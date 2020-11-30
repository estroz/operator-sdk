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

package flags

import (
	"flag"
	"fmt"
	"strings"

	"github.com/blang/semver"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/util/validation"

	"github.com/operator-framework/operator-sdk/internal/generate/clusterserviceversion"
)

const (
	replaces  = "replaces"
	skipRange = "skipRange"
)

type ResolveTypes struct {
	SkipRange string
	Replaces  string
}

var _ flag.Value = &ResolveTypes{}
var _ clusterserviceversion.Updater = &ResolveTypes{}

// Set is called when the --resolve-mode flag is passed to the CLI. It will
// configure the ResolveTypes based on the values passed in.
func (m *ResolveTypes) Set(str string) error {
	for _, mode := range strings.Split(str, ",") {
		splitEq := strings.SplitN(mode, "=", 2)
		if len(splitEq) != 2 {
			return fmt.Errorf("invalid resolve mode %q: must have format \"<type>=<value>\"", mode)
		}
		modeType, value := strings.TrimSpace(splitEq[0]), strings.TrimSpace(splitEq[1])
		if modeType == "" {
			return fmt.Errorf("resolve mode type cannot be empty")
		}
		if value == "" {
			return fmt.Errorf("resolve mode value for type %q cannot be empty", modeType)
		}
		valueUnquoted, err := removeQuotes(value)
		if err != nil {
			return fmt.Errorf("error unquoting %q: %v", value, err)
		}
		value = valueUnquoted
		switch strings.TrimSpace(modeType) {
		case replaces:
			if m.Replaces != "" {
				return fmt.Errorf("duplicate %q resolve mode", modeType)
			}
			m.Replaces = value
		case skipRange:
			if m.SkipRange != "" {
				return fmt.Errorf("duplicate %q resolve mode", modeType)
			}
			m.SkipRange = value
		default:
			return fmt.Errorf("invalid resolve mode type %q", modeType)
		}
	}
	return m.Validate()
}

// IsEmpty returns true if ResolveTypes is empty.
func (m ResolveTypes) IsEmpty() bool {
	return m == (ResolveTypes{})
}

func (m ResolveTypes) String() string {
	var types []string
	if m.Replaces != "" {
		types = append(types, fmt.Sprintf("%s=%q", replaces, m.Replaces))
	}
	if m.SkipRange != "" {
		types = append(types, fmt.Sprintf("%s=%q", skipRange, m.SkipRange))
	}
	return strings.Join(types, ",")
}

func (ResolveTypes) Type() string {
	return "ResolveTypeValue"
}

func (ResolveTypes) Description() string {
	exampleFormat := ResolveTypes{
		Replaces:  "<existing-csv-name>",
		SkipRange: "<semver-range>",
	}
	return fmt.Sprintf("Types of upgrade graph fields to populate in a CSV. Format: %s", exampleFormat)
}

func (m ResolveTypes) Validate() error {
	if m.Replaces != "" {
		if errs := validation.IsDNS1123Subdomain(m.Replaces); len(errs) > 0 {
			return fmt.Errorf("invalid replaces %q: %+q", m.Replaces, errs)
		}
	}
	if m.SkipRange != "" {
		if _, err := semver.ParseRange(m.SkipRange); err != nil {
			return fmt.Errorf("invalid skipRange %q: %v", m.SkipRange, err)
		}
	}
	return nil
}

func removeQuotes(s string) (string, error) {
	lenS := len(s)
	if lenS < 2 {
		return s, nil
	}
	trimChar := s[0]
	switch trimChar {
	case '\'', '"', '`':
	default:
		return s, nil
	}
	if s[lenS-1] != trimChar {
		return "", fmt.Errorf("string %s is incorrectly quoted", s)
	}
	return s[1 : lenS-1], nil
}

const skipRangeKey = "olm.skipRange"

func (m ResolveTypes) Update(csv *v1alpha1.ClusterServiceVersion) error {
	if m.Replaces != "" {
		csv.Spec.Replaces = m.Replaces
	}
	if m.SkipRange != "" {
		annotations := csv.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations[skipRangeKey] = m.SkipRange
		csv.SetAnnotations(annotations)
	}
	return nil
}
