// Copyright 2020-2022 Intel Corporation. All Rights Reserved.
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

var _ = Describe("DsaDevicePlugin Controller", func() {

	Context("Basic CRUD operations", func() {
		It("should handle DsaDevicePlugin objects correctly", func() {
			By("creating DsaDevicePlugin successfully")
			spec := devicepluginv1.DsaDevicePluginSpec{
				NodeSelector: map[string]string{"test": "nodeselector"},
				Image:        "testimage",
				InitImage:    "testinitimage",
				LogLevel:     3,
			}
			fetched := &devicepluginv1.DsaDevicePlugin{}
			testCreateDevicePluginWithSpec(dsa, spec, fetched)

			By("creating DsaDevicePlugin without setting Spec.NodeSelector successfully")
			spec = devicepluginv1.DsaDevicePluginSpec{
				Image: "testimage",
			}
			testDelete(dsa)
			testCreateDevicePluginWithSpec(dsa, spec, fetched)

			By("updating DsaDevicePlugin successfully")
			spec = devicepluginv1.DsaDevicePluginSpec{
				Image:        "updated-testimage",
				NodeSelector: map[string]string{"test": "updated-node-selector"},
				LogLevel:     4,
			}
			fetched.Spec = spec
			fetchedUpdated := &devicepluginv1.DsaDevicePlugin{}
			testUpdateDevicePlugin(fetched)
			testUpdateImage(dsa, fetched, fetchedUpdated)
			testUpdateInitImage(dsa, fetched, fetchedUpdated)
			testUpdateArgs(dsa, fetched, fetchedUpdated)
			testUpdateNodeSelector(dsa, fetched)

			fetchedUpdated.Spec.NodeSelector = map[string]string{}
			testUpdateDevicePlugin(fetchedUpdated)
			testUpdateNodeSelector(dsa, fetchedUpdated)

			By("deleting DsaDevicePlugin successfully")
			testDelete(dsa)
		})
	})

	It("upgrades", func() {
		dp := &devicepluginv1.DsaDevicePlugin{}

		var image, initimage string

		testUpgrade("dsa", dp, &image, &initimage)

		Expect(dp.Spec.Image == image).To(BeTrue())
		Expect(dp.Spec.InitImage == initimage).To(BeTrue())
	})
})
