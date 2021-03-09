// Copyright 2021 The Operator-SDK Authors
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

package v2

import (
	"path/filepath"

	"sigs.k8s.io/kubebuilder/v3/pkg/model/file"
)

var _ file.Template = &managerWebhookPatch{}

// managerWebhookPatch scaffolds a file that defines the patch that enables webhooks on the manager
type managerWebhookPatch struct {
	file.TemplateMixin
}

// SetTemplateDefaults implements file.Template
func (f *managerWebhookPatch) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join(patchesDir, "manager_webhook_patch.yaml")
	}

	f.TemplateBody = managerWebhookPatchTemplate

	f.IfExistsAction = file.Error

	return nil
}

const managerWebhookPatchTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
  namespace: system
spec:
  template:
    spec:
      containers:
      - name: manager
        ports:
        - containerPort: 9443
          name: webhook-server
          protocol: TCP
`
