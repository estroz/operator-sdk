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
	"fmt"
	"time"

	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/operator-framework/operator-sdk/internal/olm/operator"
)

type CatalogCreator interface {
	CreateCatalog(ctx context.Context, name string) (*v1alpha1.CatalogSource, error)
}

// TODO: modify this as necessary.
type InstallPlanApprover interface {
	Approve(ctx context.Context, name string) error
}

func waitForCatalogSource(ctx context.Context, cfg *operator.Configuration, cs *v1alpha1.CatalogSource) error {
	catSrcKey, err := client.ObjectKeyFromObject(cs)
	if err != nil {
		return fmt.Errorf("error in getting catalog source key: %v", err)
	}

	// verify that catalog source connection status is READY
	catSrcCheck := wait.ConditionFunc(func() (done bool, err error) {
		if err := cfg.Client.Get(ctx, catSrcKey, cs); err != nil {
			return false, err
		}
		if cs.Status.GRPCConnectionState != nil {
			if cs.Status.GRPCConnectionState.LastObservedState == "READY" {
				return true, nil
			}
		}
		return false, nil
	})

	if err := wait.PollImmediateUntil(200*time.Millisecond, catSrcCheck, ctx.Done()); err != nil {
		return fmt.Errorf("catalog source connection is not ready: %v", err)
	}

	return nil
}
