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

var _ file.Template = &kustomization{}

// kustomization scaffolds a file that defines the kustomization scheme for the default overlay folder
type kustomization struct {
	file.TemplateMixin
	file.ProjectNameMixin
	file.ComponentConfigMixin

	SupportsWebhooks bool
}

// SetTemplateDefaults implements file.Template
func (f *kustomization) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = filepath.Join(manifestsDir, "kustomization.yaml")
	}

	f.TemplateBody = kustomizationTemplate

	f.IfExistsAction = file.Error

	return nil
}

const kustomizationTemplate = `# Value of this field is prepended to the
# names of all resources, e.g. a deployment named
# "wordpress" becomes "alices-wordpress".
# Note that it should also match with the prefix (text before '-') of the namespace
# field above.
namePrefix: {{ .ProjectName }}-

# Labels to add to all resources and selectors.
#commonLabels:
#  someName: someValue

resources:
- ../crd
- ../rbac
- ../manager
- ../samples
- ../scorecard
{{- if .SupportsWebhooks }}
# [WEBHOOK] To enable mutating and/or validating webhooks, uncomment the below list element.
# To enable conversion webhooks, additionally uncomment the 'WEBHOOK' section in config/crd/kustomization.yaml.
#- ../webhook
{{- end }}
# [PROMETHEUS] To enable prometheus monitor, uncomment all sections with 'PROMETHEUS'.
#- ../prometheus

patchesStrategicMerge:
# Protect the /metrics endpoint by putting it behind auth.
# If you want your controller-manager to expose the /metrics
# endpoint w/o any authn/z, please comment the following line.
- patches/manager_auth_proxy_patch.yaml
{{ if .SupportsWebhooks }}
# [WEBHOOK] To enable mutating and/or validating webhooks, uncomment the below list element.
# To enable conversion webhooks, additionally uncomment the 'WEBHOOK' section in config/crd/kustomization.yaml.
#- patches/manager_webhook_patch.yaml
{{ end -}}
`
