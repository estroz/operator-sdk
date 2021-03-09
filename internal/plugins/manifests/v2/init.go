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
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/kubebuilder/v3/pkg/config"
	"sigs.k8s.io/kubebuilder/v3/pkg/model"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/file"

	"github.com/operator-framework/operator-sdk/internal/kubebuilder/machinery"
	"github.com/operator-framework/operator-sdk/internal/util/projutil"
)

// runInit runs the manifests SDK phase 2 plugin.
func runInit(cfg config.Config) error {

	if err := newInitScaffolder(cfg).scaffold(); err != nil {
		return err
	}

	return nil
}

type initScaffolder struct {
	config config.Config
}

func newInitScaffolder(config config.Config) *initScaffolder {
	return &initScaffolder{
		config: config,
	}
}

func (s *initScaffolder) newUniverse() *model.Universe {
	return model.NewUniverse(
		model.WithConfig(s.config),
	)
}

func (s *initScaffolder) scaffold() error {

	if err := os.MkdirAll(patchesDir, 0755); err != nil {
		return fmt.Errorf("error creating manifests patches dir: %v", err)
	}

	var builders []file.Builder
	operatorType := projutil.PluginKeyToOperatorType(s.config.GetLayout())
	switch operatorType {
	case projutil.OperatorTypeUnknown:
		return fmt.Errorf("unsupported plugin key %q", s.config.GetLayout())
	case projutil.OperatorTypeGo:
		builders = append(builders,
			&kustomization{SupportsWebhooks: true},
			&managerWebhookPatch{},
		)
	default:
		builders = append(builders,
			&kustomization{SupportsWebhooks: false},
		)
	}

	err := machinery.NewScaffold().Execute(s.newUniverse(), builders...)
	if err != nil {
		return fmt.Errorf("error scaffolding manifests: %v", err)
	}

	for _, f := range []string{"manager_auth_proxy_patch.yaml"} {
		if err := copyFile(filepath.Join("config", "default", f), filepath.Join(patchesDir, f)); err != nil {
			return fmt.Errorf("error copying file: %v", err)
		}
	}

	return nil
}

func copyFile(fromPath, toPath string) error {
	r, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer closeFile(r)
	info, err := r.Stat()
	if err != nil {
		return err
	}

	w, err := os.OpenFile(toPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode())
	if err != nil {
		return err
	}
	defer closeFile(w)

	_, err = io.Copy(w, r)
	return err
}

func closeFile(f *os.File) {
	if err := f.Close(); err != nil {
		log.Errorf("Failed to close file: %v", err)
	}
}
