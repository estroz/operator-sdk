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
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-sdk/internal/olm/operator"
	"k8s.io/client-go/kubernetes/scheme"

	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreateRegistryPod(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Registry Pod Suite")
}

var _ = Describe("RegistryPod", func() {

	Describe("creating registry pod", func() {

		Context("with valid registry pod values", func() {
			expectedPodName := "quay-io-example-example-operator-bundle-0-2-0"
			expectedOutput := "/bin/mkdir -p /database/index.db &&" +
				"/bin/opm registry add -d /database/index.db -b quay.io/example/example-operator-bundle:0.2.0 --mode=semver &&" +
				"/bin/opm registry serve -d /database/index.db -p 50051"

			var rp *RegistryPod
			var err error

			BeforeEach(func() {
				rp = newRegistryPod("/database/index.db", "quay.io/example/example-operator-bundle:0.2.0")
				Expect(err).To(BeNil())
			})

			It("should validate the RegistryPod successfully", func() {
				err := rp.validate()

				Expect(err).To(BeNil())
			})

			It("should create the RegistryPod successfully", func() {
				cfg := newFakeConfig()
				pod, err := rp.Create(context.TODO(), cfg, newCatalogSource())
				Expect(err).To(BeNil())
				Expect(pod).NotTo(BeNil())
				Expect(pod.GetName()).To(Equal(expectedPodName))
				Expect(pod.GetNamespace()).To(Equal(cfg.Namespace))
				Expect(pod.Spec.Containers[0].Name).To(Equal(registryContainerName))
				if len(pod.Spec.Containers) > 0 {
					if len(pod.Spec.Containers[0].Ports) > 0 {
						Expect(pod.Spec.Containers[0].Ports[0].ContainerPort).To(BeEquivalentTo(registryGRPCPort))
					}
				}
			})

			It("should return a valid container command", func() {
				output, err := rp.getContainerCmd()

				Expect(err).To(BeNil())
				Expect(output).Should(Equal(expectedOutput))
			})

			It("should return a pod definition successfully", func() {
				cfg := newFakeConfig()
				pod, err := rp.podForBundleRegistry(cfg.Namespace)
				Expect(err).To(BeNil())
				Expect(pod).NotTo(BeNil())
				Expect(pod.GetName()).To(Equal(expectedPodName))
				Expect(pod.GetNamespace()).To(Equal(cfg.Namespace))
				Expect(pod.Spec.Containers[0].Name).To(Equal(registryContainerName))
				if len(pod.Spec.Containers) > 0 {
					if len(pod.Spec.Containers[0].Ports) > 0 {
						Expect(pod.Spec.Containers[0].Ports[0].ContainerPort).To(BeEquivalentTo(registryGRPCPort))
					}
				}
			})
		})

		Context("with invalid registry pod values", func() {

			It("should error when bundle image is not provided", func() {
				expectedErr := "bundle image cannot be empty"

				rp := newRegistryPod("/database/index.db", "")
				err := rp.validate()
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).Should(ContainSubstring(expectedErr))
			})

			It("should not create a registry pod when database path is not provided", func() {
				expectedErr := "registry database path cannot be empty"

				rp := newRegistryPod("", "quay.io/example/example-operator-bundle:0.2.0")
				err := rp.validate()
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).Should(ContainSubstring(expectedErr))
			})

			It("should not create a registry pod when bundle add mode is empty", func() {
				expectedErr := "bundle add mode cannot be empty"

				rp := newRegistryPod("/database/index.db", "quay.io/example/example-operator-bundle:0.2.0")
				rp.BundleAddMode = ""

				err := rp.validate()
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).Should(ContainSubstring(expectedErr))
			})

			It("should not accept any other bundle add mode other than semver or replaces", func() {
				expectedErr := `unknown bundle add mode "invalid"`

				_, err := ParseBundleAddMode("invalid")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).Should(ContainSubstring(expectedErr))
			})

			// todo(rashmigottipati): add test to check VerifyPodRunning returning error
		})
	})
})

func newRegistryPod(dbPath, bundleImage string) *RegistryPod {
	return &RegistryPod{
		BundleImage:   bundleImage,
		IndexImage:    defaultIndexImage,
		BundleAddMode: SemverBundleAddMode,
		DBPath:        dbPath,
	}
}

func newCatalogSource() *v1alpha1.CatalogSource {
	cs := &v1alpha1.CatalogSource{}
	cs.SetGroupVersionKind(v1alpha1.SchemeGroupVersion.WithKind("CatalogSource"))
	return cs
}

// newFakeConfig() returns a fake controller runtime client
func newFakeConfig() *operator.Configuration {
	sch := scheme.Scheme
	_ = v1alpha1.AddToScheme(sch)
	return &operator.Configuration{
		Namespace: "default",
		Client:    fakeclient.NewFakeClient(),
		Scheme:    sch,
	}
}
