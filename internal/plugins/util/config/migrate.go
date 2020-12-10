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

package config

import (
	"sigs.k8s.io/kubebuilder/v2/pkg/model/config"
)

func Migrate(cfg *config.Config) error {
	switch {
	case cfg.Version == config.Version3Alpha:
		return migrate3Alpha(cfg)
	}

	return nil
}

// config3AlphaLegacy is the "old" version of kubebuilder's config.Config before it changed
// in https://github.com/kubernetes-sigs/kubebuilder/pull/1869
type config3AlphaOld struct {
	Version         string               `json:"version,omitempty"`
	Domain          string               `json:"domain,omitempty"`
	Repo            string               `json:"repo,omitempty"`
	ProjectName     string               `json:"projectName,omitempty"`
	Resources       []config.GVK         `json:"resources,omitempty"`
	MultiGroup      bool                 `json:"multigroup,omitempty"`
	ComponentConfig bool                 `json:"componentConfig,omitempty"`
	Layout          string               `json:"layout,omitempty"`
	Plugins         config.PluginConfigs `json:"plugins,omitempty"`
}

func migrate3Alpha(cfg *config.Config) error {
	return nil
}
