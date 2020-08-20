// Copyright 2019 The Operator-SDK Authors
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
	"fmt"

	v1 "github.com/operator-framework/api/pkg/operators/v1"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-sdk/internal/olm/operator"
	"github.com/operator-framework/operator-sdk/internal/util/k8sutil"
)

func getSubscriptionName(csvName string) string {
	name := k8sutil.FormatOperatorNameDNS1123(csvName)
	return fmt.Sprintf("%s-sub", name)
}

// withCatalogSource returns a function that sets the Subscription argument's
// target CatalogSource's name and namespace.
func withCatalogSource(csName, csNamespace string) func(*v1alpha1.Subscription) {
	return func(sub *v1alpha1.Subscription) {
		sub.Spec.CatalogSource = csName
		sub.Spec.CatalogSourceNamespace = csNamespace
	}
}

// withPackageChannel returns a function that sets the Subscription argument's
// target package, channel, and starting CSV to those in channel.
func withPackageChannel(pkgName, channelName, startingCSV string) func(*v1alpha1.Subscription) {
	return func(sub *v1alpha1.Subscription) {
		if sub.Spec == nil {
			sub.Spec = &v1alpha1.SubscriptionSpec{}
		}
		sub.Spec.Package = pkgName
		sub.Spec.Channel = channelName
		sub.Spec.StartingCSV = startingCSV
	}
}

// newSubscription creates a new Subscription for a CSV with a name derived
// from csvName, the CSV's objectmeta.name, in namespace. opts will be applied
// to the Subscription object.
func newSubscription(csvName, namespace string,
	opts ...func(*v1alpha1.Subscription)) *v1alpha1.Subscription {
	sub := &v1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.SubscriptionKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getSubscriptionName(csvName),
			Namespace: namespace,
		},
	}
	for _, opt := range opts {
		opt(sub)
	}
	return sub
}

func withSDKPublisher(pkgName string) func(*v1alpha1.CatalogSource) {
	return func(cs *v1alpha1.CatalogSource) {
		cs.Spec.DisplayName = pkgName
		cs.Spec.Publisher = "operator-sdk"
	}
}

// newCatalogSource creates a new CatalogSource with a name derived from
// pkgName, the package manifest's packageName, in namespace. opts will
// be applied to the CatalogSource object.
func newCatalogSource(name, namespace string,
	opts ...func(*v1alpha1.CatalogSource)) *v1alpha1.CatalogSource {
	cs := &v1alpha1.CatalogSource{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       v1alpha1.CatalogSourceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	for _, opt := range opts {
		opt(cs)
	}
	return cs
}

// withGRPC returns a function that sets the OperatorGroup argument's
// targetNamespaces to namespaces. namespaces can be length 0..N; if
// namespaces length is 0, targetNamespaces is set to an empty string,
// indicating a global scope.
func withTargetNamespaces(namespaces ...string) func(*v1.OperatorGroup) {
	return func(og *v1.OperatorGroup) {
		if len(namespaces) != 0 && namespaces[0] != "" {
			og.Spec.TargetNamespaces = namespaces
		}
	}
}

// newSDKOperatorGroup creates a new OperatorGroup with name
// sdkOperatorGroupName in namespace. opts will be applied to the
// OperatorGroup object. Note that the default OperatorGroup has a global
// scope.
func newSDKOperatorGroup(namespace string, opts ...func(*v1.OperatorGroup)) *v1.OperatorGroup {
	og := &v1.OperatorGroup{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       v1.OperatorGroupKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      operator.SDKOperatorGroupName,
			Namespace: namespace,
		},
	}
	for _, opt := range opts {
		opt(og)
	}
	return og
}
