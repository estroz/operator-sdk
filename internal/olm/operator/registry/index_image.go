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

package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/operator-framework/operator-sdk/internal/olm/operator"
	"github.com/operator-framework/operator-sdk/internal/olm/operator/registry/index"
	registryutil "github.com/operator-framework/operator-sdk/internal/registry"
)

type IndexImageCatalogCreator struct {
	PackageName      string
	IndexImage       string
	InjectBundles    []string
	InjectBundleMode string
	BundleImage      string

	cfg *operator.Configuration
}

func NewIndexImageCatalogCreator(cfg *operator.Configuration) *IndexImageCatalogCreator {
	return &IndexImageCatalogCreator{
		cfg: cfg,
	}
}

func (c IndexImageCatalogCreator) CreateCatalog(ctx context.Context, name string) (*v1alpha1.CatalogSource, error) {
	dbPath, err := c.getDBPath(ctx)
	if err != nil {
		return nil, fmt.Errorf("get database path: %v", err)
	}

	log.Infof("IndexImageCatalogCreator.IndexImage:        %q\n", c.IndexImage)
	log.Infof("IndexImageCatalogCreator.IndexImageDBPath:  %v\n", dbPath)
	log.Infof("IndexImageCatalogCreator.InjectBundles:     %q\n", strings.Join(c.InjectBundles, ","))
	log.Infof("IndexImageCatalogCreator.InjectBundleMode:  %q\n", c.InjectBundleMode)

	// create a basic catalog source type
	cs := newCatalogSource(name, c.cfg.Namespace,
		withSDKPublisher(c.PackageName))
	if err := c.cfg.Client.Create(ctx, cs); err != nil {
		return nil, fmt.Errorf("error creating catalog source: %w", err)
	}

	// initialize and create the registry pod with provided index image
	registryPod, err := c.createRegistryPod(ctx, cs, dbPath)
	if err != nil {
		return nil, fmt.Errorf("error in creating registry pod: %v", err)
	}

	// update catalog source with source type, address and annotations
	if err := c.updateCatalogSource(ctx, registryPod.Status.PodIP, cs); err != nil {
		return nil, fmt.Errorf("error in updating catalog source: %v", err)
	}

	// wait for catalog source to be ready
	if err := waitForCatalogSource(ctx, c.cfg, cs); err != nil {
		return nil, err
	}

	return cs, nil
}

const defaultDBPath = "/database/index.db"

func (c IndexImageCatalogCreator) getDBPath(ctx context.Context) (string, error) {
	labels, err := registryutil.GetImageLabels(ctx, nil, c.IndexImage, false)
	if err != nil {
		return "", fmt.Errorf("get index image labels: %v", err)
	}
	if dbPath, ok := labels["operators.operatorframework.io.index.database.v1"]; ok {
		return dbPath, nil
	}
	return defaultDBPath, nil
}

func (c IndexImageCatalogCreator) createRegistryPod(ctx context.Context, cs *v1alpha1.CatalogSource, dbPath string) (pod *corev1.Pod, err error) {
	rp := &index.RegistryPod{
		IndexImage:  c.IndexImage,
		BundleImage: c.BundleImage,
		DBPath:      dbPath,
	}
	if rp.BundleAddMode, err = index.ParseBundleAddMode(c.InjectBundleMode); err != nil {
		return nil, err
	}

	if pod, err = rp.Create(ctx, c.cfg, cs); err != nil {
		return nil, err
	}

	podKey, err := client.ObjectKeyFromObject(pod)
	if err != nil {
		return nil, err
	}
	// upon creation of new pod, poll and verify that pod status is running
	podCheck := wait.ConditionFunc(func() (done bool, err error) {
		err = c.cfg.Client.Get(ctx, podKey, pod)
		if err != nil {
			return false, fmt.Errorf("error getting pod %s: %w", podKey.Name, err)
		}
		return pod.Status.Phase == corev1.PodRunning, nil
	})

	// check pod status to be Running
	// poll every 200 ms until podCheck is true or context is done
	err = wait.PollImmediateUntil(200*time.Millisecond, podCheck, ctx.Done())
	if err != nil {
		return nil, fmt.Errorf("error waiting for registry pod %s to run: %v", podKey.Name, err)
	}
	return pod, nil
}

func (c IndexImageCatalogCreator) updateCatalogSource(_ context.Context, podAddr string, cs *v1alpha1.CatalogSource) error {
	// Update catalog source with source type as grpc and address to point to the pod IP
	cs.Spec.SourceType = v1alpha1.SourceTypeGrpc
	cs.Spec.Address = index.GetRegistryPodHost(podAddr)

	// Update catalog source with annotations for index image,
	// injected bundle, and registry add mode
	injectedBundlesJSON, err := json.Marshal(c.InjectBundles)
	if err != nil {
		return fmt.Errorf("error in json marshal injected bundles: %v", err)
	}
	cs.ObjectMeta.Annotations = map[string]string{
		"operators.operatorframework.io/index-image":        c.IndexImage,
		"operators.operatorframework.io/inject-bundle-mode": c.InjectBundleMode,
		"operators.operatorframework.io/injected-bundles":   string(injectedBundlesJSON),
	}

	return nil
}
