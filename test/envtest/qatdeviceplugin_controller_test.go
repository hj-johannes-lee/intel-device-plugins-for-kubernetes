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
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	devicepluginv1 "github.com/intel/intel-device-plugins-for-kubernetes/pkg/apis/deviceplugin/v1"
)

var _ = Describe("QatDevicePlugin Controller", func() {

	Context("Basic CRUD operations", func() {
		It("should handle QatDevicePlugin objects correctly", func() {
			By("creating QatDevicePlugin successfully")
			spec := devicepluginv1.QatDevicePluginSpec{
				NodeSelector: map[string]string{"test": "nodeselector"},
				Image:        "testimage",
				LogLevel:     3,
			}
			annotations := map[string]string{
				"container.apparmor.security.beta.kubernetes.io/intel-qat-plugin": "unconfined",
			}
			toCreate := &devicepluginv1.QatDevicePlugin{
				ObjectMeta: metav1.ObjectMeta{
					Name:        getKey(qat).Name,
					Annotations: annotations,
				},
				Spec: spec,
			}
			fetched := &devicepluginv1.QatDevicePlugin{}
			testCreateDevicePlugin(qat, toCreate, fetched)

			By("copy annotations successfully")
			Expect(&(fetched.Annotations) == &annotations).ShouldNot(BeTrue())
			Eventually(fetched.Annotations).Should(Equal(annotations))

			By("updating annotations successfully")
			updatedAnnotations := map[string]string{"key": "value"}
			fetched.Annotations = updatedAnnotations
			updated := &devicepluginv1.QatDevicePlugin{}
			testUpdateDevicePlugin(fetched)
			Eventually(func() map[string]string {
				_ = k8sClient.Get(context.Background(), getKey(qat), updated)
				return updated.Annotations
			}, timeout, interval).Should(Equal(updatedAnnotations))

			By("creating QatDevicePlugin without setting Spec.NodeSelector successfully")
			spec = devicepluginv1.QatDevicePluginSpec{
				Image: "testimage",
			}
			toCreate = &devicepluginv1.QatDevicePlugin{
				ObjectMeta: metav1.ObjectMeta{
					Name: getKey(qat).Name,
				},
				Spec: spec,
			}
			testDelete(qat)
			testCreateDevicePlugin(qat, toCreate, fetched)

			By("updating QatDevicePlugin successfully")
			spec = devicepluginv1.QatDevicePluginSpec{
				Image:        "updated-testimage",
				NodeSelector: map[string]string{"test": "updated-node-selector"},
				LogLevel:     4,
			}
			fetched.Spec = spec
			fetchedUpdated := &devicepluginv1.QatDevicePlugin{}
			testUpdateDevicePlugin(fetched)
			testUpdateImage(qat, fetched, fetchedUpdated)
			testUpdateInitImage(qat, fetched, fetchedUpdated)
			testUpdateArgs(qat, fetched, fetchedUpdated)
			testUpdateNodeSelector(qat, fetched)

			fetchedUpdated.Spec.NodeSelector = map[string]string{}
			testUpdateDevicePlugin(fetchedUpdated)
			testUpdateNodeSelector(qat, fetchedUpdated)

			By("deleting QatDevicePlugin successfully")
			testDelete(qat)
		})
	})

	It("upgrades", func() {
		dp := &devicepluginv1.QatDevicePlugin{}

		var image string

		testUpgrade("qat", dp, &image, nil)

		Expect(dp.Spec.Image == image).To(BeTrue())
	})
})
