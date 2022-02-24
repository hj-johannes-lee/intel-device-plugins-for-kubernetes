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

var _ = Describe("FpgaDevicePlugin Controller", func() {

	Context("Basic CRUD operations", func() {
		It("should handle FpgaDevicePlugin objects correctly", func() {
			By("creating FpgaDevicePlugin successfully")
			spec := devicepluginv1.FpgaDevicePluginSpec{
				NodeSelector: map[string]string{"test": "nodeselector"},
				Image:        "testimage",
				InitImage:    "testinitimage",
				LogLevel:     3,
			}
			fetched := &devicepluginv1.FpgaDevicePlugin{}
			testCreateDevicePluginWithSpec(fpga, spec, fetched)

			By("creating FpgaDevicePlugin without setting Spec.NodeSelector successfully")
			spec = devicepluginv1.FpgaDevicePluginSpec{
				Image:     "testimage",
				InitImage: "testinitimage",
			}
			testDelete(fpga)
			testCreateDevicePluginWithSpec(fpga, spec, fetched)

			By("updating FpgaDevicePlugin successfully")
			spec = devicepluginv1.FpgaDevicePluginSpec{
				Image:        "updated-testimage",
				InitImage:    "updated-testinitimage",
				NodeSelector: map[string]string{"test": "updated-node-selector"},
				LogLevel:     4,
			}
			fetched.Spec = spec
			fetchedUpdated := &devicepluginv1.FpgaDevicePlugin{}
			testUpdateDevicePlugin(fetched)
			testUpdateImage(fpga, fetched, fetchedUpdated)
			testUpdateInitImage(fpga, fetched, fetchedUpdated)
			testUpdateArgs(fpga, fetched, fetchedUpdated)
			testUpdateNodeSelector(fpga, fetched)

			fetchedUpdated.Spec.NodeSelector = map[string]string{}
			testUpdateDevicePlugin(fetchedUpdated)
			testUpdateNodeSelector(fpga, fetchedUpdated)

			By("deleting FpgaDevicePlugin successfully")
			testDelete(fpga)
		})
	})

	It("upgrades", func() {
		dp := &devicepluginv1.FpgaDevicePlugin{}

		var image, initimage string

		testUpgrade("fpga", dp, &image, &initimage)

		Expect(dp.Spec.Image == image).To(BeTrue())
		Expect(dp.Spec.InitImage == initimage).To(BeTrue())
	})
})
