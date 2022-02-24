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

var _ = Describe("GpuDevicePlugin Controller", func() {

	Context("Basic CRUD operations", func() {
		It("should handle GpuDevicePlugin objects correctly", func() {
			By("creating GpuDevicePlugin successfully")
			spec := devicepluginv1.GpuDevicePluginSpec{
				NodeSelector: map[string]string{"test": "nodeselector"},
				Image:        "testimage",
				InitImage:    "testinitimage",
				LogLevel:     3,
			}
			fetched := &devicepluginv1.GpuDevicePlugin{}
			testCreateDevicePluginWithSpec(gpu, spec, fetched)

			By("creating GpuDevicePlugin without setting Spec.NodeSelector successfully")
			spec = devicepluginv1.GpuDevicePluginSpec{
				Image: "testimage",
			}
			testDelete(gpu)
			testCreateDevicePluginWithSpec(gpu, spec, fetched)

			By("updating GpuDevicePlugin successfully")
			spec = devicepluginv1.GpuDevicePluginSpec{
				Image:        "updated-testimage",
				NodeSelector: map[string]string{"test": "updated-node-selector"},
				LogLevel:     4,
			}
			fetched.Spec = spec
			fetchedUpdated := &devicepluginv1.GpuDevicePlugin{}
			testUpdateDevicePlugin(fetched)
			testUpdateImage(gpu, fetched, fetchedUpdated)
			testUpdateInitImage(gpu, fetched, fetchedUpdated)
			testUpdateArgs(gpu, fetched, fetchedUpdated)
			testUpdateNodeSelector(gpu, fetched)

			fetchedUpdated.Spec.NodeSelector = map[string]string{}
			testUpdateDevicePlugin(fetchedUpdated)
			testUpdateNodeSelector(gpu, fetchedUpdated)

			By("deleting GpuDevicePlugin successfully")
			testDelete(gpu)
		})
	})

	It("upgrades", func() {
		dp := &devicepluginv1.GpuDevicePlugin{}

		var image, initimage string

		testUpgrade("gpu", dp, &image, &initimage)

		Expect(dp.Spec.Image == image).To(BeTrue())
		Expect(dp.Spec.InitImage == initimage).To(BeTrue())
	})
})
