// Copyright 2020 Intel Corporation. All Rights Reserved.
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

package fpgacontroller

import (
	"errors"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"

	fpgav2 "github.com/intel/intel-device-plugins-for-kubernetes/pkg/apis/fpga/v2"
	"github.com/intel/intel-device-plugins-for-kubernetes/pkg/fpgacontroller/patcher"
	"k8s.io/klog/v2/ktesting"
)

var (
	errClient = errors.New("client error")
	scheme    = runtime.NewScheme()
)

func init() {
	_ = fpgav2.AddToScheme(scheme)
}

func TestAcceleratorFunctionReconcile(t *testing.T) {
	tcases := []struct {
		getError    error
		name        string
		expectedErr bool
	}{
		{
			name: "empty af",
		},
		{
			name:        "client error",
			getError:    errClient,
			expectedErr: true,
		},
		{
			name:     "af not found",
			getError: apierrors.NewNotFound(schema.GroupResource{}, "fake"),
		},
	}

	for _, tt := range tcases {
		t.Run(tt.name, func(t *testing.T) {
			logger, ctx := ktesting.NewTestContext(t)
			reconciler := &AcceleratorFunctionReconciler{
				Client: &mockClient{
					getError: tt.getError,
				},
				PatcherManager: patcher.NewPatcherManager(logger),
			}
			_, err := reconciler.Reconcile(ctx, ctrl.Request{})
			if err != nil && !tt.expectedErr {
				t.Errorf("unexpected error: %+v", err)
			}
			if err == nil && tt.expectedErr {
				t.Error("expected error but got success")
			}
		})
	}
}

func TestAcceleratorFunctionSetupWithManager(t *testing.T) {
	r := &AcceleratorFunctionReconciler{}

	logger, _ := ktesting.NewTestContext(t)
	err := r.SetupWithManager(&mockManager{
		scheme: scheme,
		log:    logger,
	})

	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}

func TestFpgaRegionReconcile(t *testing.T) {
	tcases := []struct {
		getError    error
		name        string
		expectedErr bool
	}{
		{
			name: "empty region",
		},
		{
			name:        "client error",
			getError:    errClient,
			expectedErr: true,
		},
		{
			name:     "region not found",
			getError: apierrors.NewNotFound(schema.GroupResource{}, "fake"),
		},
	}

	for _, tt := range tcases {
		t.Run(tt.name, func(t *testing.T) {
			logger, ctx := ktesting.NewTestContext(t)
			reconciler := &FpgaRegionReconciler{
				Client: &mockClient{
					getError: tt.getError,
				},
				PatcherManager: patcher.NewPatcherManager(logger),
			}
			_, err := reconciler.Reconcile(ctx, ctrl.Request{})
			if err != nil && !tt.expectedErr {
				t.Errorf("unexpected error: %+v", err)
			}
			if err == nil && tt.expectedErr {
				t.Error("expected error but got success")
			}
		})
	}
}

func TestFpgaRegionSetupWithManager(t *testing.T) {
	r := &FpgaRegionReconciler{}

	logger, _ := ktesting.NewTestContext(t)
	err := r.SetupWithManager(&mockManager{
		scheme: scheme,
		log:    logger,
	})

	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}
