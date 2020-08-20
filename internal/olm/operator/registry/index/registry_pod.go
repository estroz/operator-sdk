// Copyright 2020 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package index

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/operator-framework/operator-sdk/internal/olm/operator"
	"github.com/operator-framework/operator-sdk/internal/util/k8sutil"
)

const (
	defaultIndexImage = "quay.io/operator-framework/upstream-opm-builder:latest"

	// registryGRPCPort is the default grpc container port that the registry pod exposes
	registryGRPCPort          = 50051
	registryContainerName     = "registry-grpc"
	registryContainerPortName = "grpc"
)

// BundleAddModeType - type of BundleAddMode in RegistryPod struct
type BundleAddModeType string

const (
	// SemverBundleAddMode - bundle add mode for semver
	SemverBundleAddMode BundleAddModeType = "semver"
	// ReplacesBundleAddMode - bundle add mode for replaces
	ReplacesBundleAddMode BundleAddModeType = "replaces"
)

func ParseBundleAddMode(modeRaw string) (BundleAddModeType, error) {
	switch modeRaw {
	case string(SemverBundleAddMode), string(ReplacesBundleAddMode):
		return BundleAddModeType(modeRaw), nil
	}
	return "", fmt.Errorf("unknown bundle add mode %q", modeRaw)
}

// RegistryPod holds resources necessary for creation of a registry server
type RegistryPod struct {
	// BundleAddMode specifies the graph update mode that defines how channel graphs are updated
	// It is of the type BundleAddModeType
	BundleAddMode BundleAddModeType

	// BundleImage specifies the container image that opm uses to generate and incrementally update the database
	BundleImage string

	// Index image contains a database of pointers to operator manifest content that is queriable via an API.
	// new version of an operator bundle when published can be added to an index image
	IndexImage string

	// DBPath refers to the registry DB;
	// if an index image is provided, the existing registry DB is located at /database/index.db
	DBPath string
}

// Create creates a bundle registry pod built from an index image and returns error
func (rp *RegistryPod) Create(ctx context.Context, cfg *operator.Configuration, cs *v1alpha1.CatalogSource) (*corev1.Pod, error) {
	rp.setDefaults()

	// validate the RegistryPod struct and ensure required fields are set
	if err := rp.validate(); err != nil {
		return nil, fmt.Errorf("invalid registry pod: %v", err)
	}

	// call podForBundleRegistry() to make the pod definition
	pod, err := rp.podForBundleRegistry(cfg.Namespace)
	if err != nil {
		return nil, fmt.Errorf("error in building registry pod definition: %v", err)
	}
	podKey, err := client.ObjectKeyFromObject(pod)
	if err != nil {
		return nil, fmt.Errorf("error in getting object key from the registry pod name %s: %v", pod.GetName(), err)
	}

	existingPod := &corev1.Pod{}
	if err := cfg.Client.Get(ctx, podKey, existingPod); err == nil {
		return existingPod, nil
	} else if !k8serrors.IsNotFound(err) {
		return nil, err
	}

	// make catalog source the owner of registry pod object
	if err := controllerutil.SetOwnerReference(cs, pod, cfg.Scheme); err != nil {
		return nil, fmt.Errorf("error in setting registry pod owner reference: %v", err)
	}
	if err = cfg.Client.Create(ctx, pod); err != nil {
		return nil, fmt.Errorf("error creating registry pod: %v", err)
	}

	return pod, nil
}

func (rp *RegistryPod) setDefaults() {
	if rp.IndexImage == "" {
		rp.IndexImage = defaultIndexImage
	}

	if rp.BundleAddMode == "" {
		if rp.IndexImage == defaultIndexImage {
			rp.BundleAddMode = SemverBundleAddMode
		} else {
			rp.BundleAddMode = ReplacesBundleAddMode
		}
	}
}

// validate will ensure that RegistryPod required fields are set
// and throws error if not set
func (rp *RegistryPod) validate() error {
	if len(strings.TrimSpace(rp.BundleImage)) < 1 {
		return errors.New("bundle image cannot be empty")
	}
	if len(strings.TrimSpace(rp.DBPath)) < 1 {
		return errors.New("registry database path cannot be empty")
	}
	if len(strings.TrimSpace(string(rp.BundleAddMode))) < 1 {
		return errors.New("bundle add mode cannot be empty")
	}

	if rp.BundleAddMode != SemverBundleAddMode && rp.BundleAddMode != ReplacesBundleAddMode {
		return fmt.Errorf("invalid bundle mode %q: must be one of [%q, %q]",
			rp.BundleAddMode, ReplacesBundleAddMode, SemverBundleAddMode)
	}

	return nil
}

func GetRegistryPodHost(ipStr string) string {
	return fmt.Sprintf("%s:%d", ipStr, registryGRPCPort)
}

// getPodName will return a string constructed from the bundle Image name
func getPodName(bundleImage string) string {
	// todo(rashmigottipati): need to come up with human-readable references
	// to be able to handle SHA references in the bundle images
	return k8sutil.TrimDNS1123Label(k8sutil.FormatOperatorNameDNS1123(bundleImage))
}

// podForBundleRegistry constructs and returns the registry pod definition
// and throws error when unable to build the pod definition successfully
func (rp *RegistryPod) podForBundleRegistry(namespace string) (*corev1.Pod, error) {
	// construct the container command for pod spec
	containerCmd, err := rp.getContainerCmd()
	if err != nil {
		return nil, fmt.Errorf("error in parsing container command: %v", err)
	}

	// make the pod definition
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getPodName(rp.BundleImage),
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  registryContainerName,
					Image: rp.IndexImage,
					Command: []string{
						"/bin/sh",
						"-c",
						containerCmd,
					},
					Ports: []corev1.ContainerPort{
						{Name: registryContainerPortName, ContainerPort: registryGRPCPort},
					},
				},
			},
		},
	}

	return pod, nil
}

// getContainerCmd uses templating to construct the container command
// and throws error if unable to parse and execute the container command
func (rp *RegistryPod) getContainerCmd() (string, error) {
	const containerCommand = "/bin/mkdir -p {{ .DBPath }} &&" +
		"/bin/opm registry add -d {{ .DBPath }} -b {{.BundleImage}} --mode={{.BundleAddMode}} &&" +
		"/bin/opm registry serve -d {{ .DBPath }} -p {{.GRPCPort}}"
	type bundleCmd struct {
		BundleAddMode       BundleAddModeType
		BundleImage, DBPath string
		GRPCPort            int32
	}

	var command = bundleCmd{rp.BundleAddMode, rp.BundleImage, rp.DBPath, registryGRPCPort}

	out := &bytes.Buffer{}

	// add the custom basename template function to the
	// template's FuncMap and parse the containerCommand
	tmp := template.Must(template.New("containerCommand").Parse(containerCommand))

	// execute the command by applying the parsed tmp to command
	// and write command output to out
	if err := tmp.Execute(out, command); err != nil {
		return "", fmt.Errorf("error in parsing container command: %w", err)
	}

	return out.String(), nil
}
