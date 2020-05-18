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

package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	valerrors "github.com/operator-framework/api/pkg/validation/errors"
	"github.com/operator-framework/operator-registry/pkg/lib/bundle"
	"github.com/sirupsen/logrus"
)

// Output represents the final result
type Output struct {
	Passed  bool     `json:"passed"`
	Objects []Object `json:"objects"`
}

// Object is an object which is used to return results in the JSON format
type Object struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// NewOutput creates an output for errs, sometimes unwrapping them for their
// internal data.
func NewOutput(errs ...error) (o Output) {

	// Collect validation errors, if any.
	for _, err := range errs {
		if err == nil {
			continue
		}
		verr := bundle.ValidationError{}
		if errors.As(err, &verr) {
			for _, valErr := range verr.Errors {
				o.Objects = append(o.Objects, Object{
					Type:    string(valerrors.LevelError),
					Message: valErr.Error(),
				})
			}
		} else {
			o.Objects = append(o.Objects, Object{
				Type:    string(valerrors.LevelError),
				Message: err.Error(),
			})
		}
	}

	// TODO: when using api library validation errors directly, they can be
	// warnings or errors. Only set this to false when more than 0 error occurred.
	o.Passed = len(o.Objects) == 0

	return o
}

// LogText will print the output in human readable format
func (o Output) LogText(logger *logrus.Entry) error {

	for _, obj := range o.Objects {
		lvl, err := logrus.ParseLevel(obj.Type)
		if err != nil {
			return err
		}
		switch lvl {
		case logrus.InfoLevel:
			logger.Info(obj.Message)
		case logrus.WarnLevel:
			logger.Warn(obj.Message)
		case logrus.ErrorLevel:
			logger.Error(obj.Message)
		default:
			return fmt.Errorf("unknown output level %q", obj.Type)
		}
	}

	return nil
}

// WriteJSON will print the output in JSON format
func (o Output) WriteJSON(w io.Writer) error {

	if err := o.prepare(); err != nil {
		return err
	}

	prettyJSON, err := json.MarshalIndent(o, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON output: %v", err)
	}
	fmt.Fprintln(w, string(prettyJSON))

	return nil
}

// prepare should be used when writing an Output to a non-log writer.
func (o *Output) prepare() error {
	o.Passed = true
	for i, obj := range o.Objects {
		lvl, err := logrus.ParseLevel(obj.Type)
		if err != nil {
			return err
		}
		if o.Passed && lvl == logrus.ErrorLevel {
			o.Passed = false
		}
		lvlBytes, _ := lvl.MarshalText()
		o.Objects[i].Type = string(lvlBytes)
	}
	return nil
}
