// Copyright 2021 Google Inc. All Rights Reserved.
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

package time_sharing

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/GoogleCloudPlatform/container-engine-accelerators/pkg/gpu/nvidia/mig"
)

func TestHasTimeSharingStrategy(t *testing.T) {
	cases := []struct {
		name               string
		gpuSharingStrategy string
		want               bool
	}{{
		name:               "include time-sharing solution",
		gpuSharingStrategy: "mig,time-sharing",
		want:               true,
	}, {
		name:               "don't include time-sharing solution",
		gpuSharingStrategy: "mig,mps",
		want:               false,
	}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			get := HasTimeSharingStrategy(tc.gpuSharingStrategy)
			if diff := cmp.Diff(tc.want, get); diff != "" {
				t.Error("unexpected error (-want, +got) = ", diff)
			}
		})
	}
}

func TestTimeSharingRequestValidation(t *testing.T) {
	cases := []struct {
		name              string
		requestDevicesIDs []string
		deviceCount       int
		migDeviceManager  *mig.DeviceManager
		wantError         error
	}{{
		name:              "don't have virtual device IDs",
		requestDevicesIDs: []string{"nvidia0", "nvidia1"},
		deviceCount:       1,
		wantError:         nil,
	}, {
		name:              "only have one physical device",
		requestDevicesIDs: []string{"nvidia0/vgpu0", "nvidia1/vgpu1"},
		deviceCount:       1,
		wantError:         nil,
	}, {
		name:              "only request one  virtual device",
		requestDevicesIDs: []string{"nvidia0/vgpu0"},
		deviceCount:       2,
		wantError:         nil,
	}, {
		name:              "request multiple virtual devices and have multiple physical devices",
		requestDevicesIDs: []string{"nvidia0/vgpu0", "nvidia1/vgpu1"},
		deviceCount:       2,
		wantError:         errors.New("invalid request for time-sharing solution, at most 1 nvidia.com/gpu can be requested when there are more than 1 physical GPUs in a node"),
	}, {
		name:              "request multiple virtual devices in MIG mode, but no mig device manager presents",
		requestDevicesIDs: []string{"nvidia0/gi10/vgpu0", "nvidia1/gi11/vgpu1"},
		deviceCount:       1,
		wantError:         errors.New("invalid request for time-sharing solution, node suppose to be in MIG mode, but can't find MIG device manager"),
		migDeviceManager:  nil,
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := TimeSharingRequestValidation(tc.requestDevicesIDs, tc.deviceCount, tc.migDeviceManager)
			if err != nil && tc.wantError != nil {
				if diff := cmp.Diff(tc.wantError.Error(), err.Error()); diff != "" {
					t.Error("unexpected error (-want, +got) = ", diff)
				}
			} else if err != nil {
				t.Error("unexpected error: ", err)
			} else if tc.wantError != nil {
				t.Error("unexpected want error:", err)
			}
		})
	}
}

func TestVirtualToPhysicalDeviceID(t *testing.T) {
	cases := []struct {
		name            string
		virtualDeviceID string
		wantDeviceID    string
		wantError       error
	}{{
		name:            "invalid virtual device ID",
		virtualDeviceID: "nvidia0",
		wantDeviceID:    "",
		wantError:       errors.New("virtual device ID (nvidia0) is not valid"),
	}, {
		name:            "virtual device ID for common cases",
		virtualDeviceID: "nvidia0/vgpu0",
		wantDeviceID:    "nvidia0",
		wantError:       nil,
	}, {
		name:            "only request one  virtual device",
		virtualDeviceID: "nvidia0/gi0/vgpu0",
		wantDeviceID:    "nvidia0/gi0",
		wantError:       nil,
	}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			deviceID, err := VirtualToPhysicalDeviceID(tc.virtualDeviceID)
			if diff := cmp.Diff(tc.wantDeviceID, deviceID); diff != "" {
				t.Error("unexpected deviceID (-want, +got) = ", diff)
			}
			if err != nil && tc.wantError != nil {
				if diff := cmp.Diff(tc.wantError.Error(), err.Error()); diff != "" {
					t.Error("unexpected error (-want, +got) = ", diff)
				}
			} else if err != nil {
				t.Error("unexpected error: ", err)
			} else if tc.wantError != nil {
				t.Error("unexpected want error:", err)
			}
		})
	}
}
