// Copyright 2021-2022 Intel Corporation. All Rights Reserved.
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

package envtest

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	devicepluginv1 "github.com/intel/intel-device-plugins-for-kubernetes/pkg/apis/deviceplugin/v1"
)

var _ = Describe("DlbDevicePlugin Controller", func() {

	Context("Basic CRUD operations", func() {
		It("should handle DlbDevicePlugin objects correctly", func() {
			By("creating DlbDevicePlugin successfully")
			spec := devicepluginv1.DlbDevicePluginSpec{
				Image:        "testimage",
				NodeSelector: map[string]string{"feature.node.kubernetes.io/dlb": "true"},
				LogLevel:     4,
			}
			fetched := &devicepluginv1.DlbDevicePlugin{}

			testCreateDevicePluginWithSpec(dlb, spec, fetched)

			By("creating DlbDevicePlugin without setting Spec.NodeSelector successfully")
			spec = devicepluginv1.DlbDevicePluginSpec{
				Image: "testimage",
			}
			testDelete(dlb)
			testCreateDevicePluginWithSpec(dlb, spec, fetched)

			By("updating DlbDevicePlugin successfully")
			spec = devicepluginv1.DlbDevicePluginSpec{
				Image:        "updated-testimage",
				NodeSelector: map[string]string{"test": "updated-node-selector"},
				LogLevel:     3,
			}
			fetched.Spec = spec
			fetchedUpdated := &devicepluginv1.DlbDevicePlugin{}
			testUpdateDevicePlugin(fetched)
			testUpdateImage(dlb, fetched, fetchedUpdated)
			testUpdateArgs(dlb, fetched, fetchedUpdated)
			testUpdateNodeSelector(dlb, fetched)

			fetchedUpdated.Spec.NodeSelector = map[string]string{}
			testUpdateDevicePlugin(fetchedUpdated)
			testUpdateNodeSelector(dlb, fetchedUpdated)

			By("deleting DlbDevicePlugin successfully")
			testDelete("dlb")
		})
	})

	It("upgrades", func() {
		dp := &devicepluginv1.DlbDevicePlugin{}

		var image string

		testUpgrade("dlb", dp, &image, nil)

		Expect(dp.Spec.Image == image).To(BeTrue())
	})
})
