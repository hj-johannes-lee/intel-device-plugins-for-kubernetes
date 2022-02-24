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
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/intel/intel-device-plugins-for-kubernetes/deployments"
	devicepluginv1 "github.com/intel/intel-device-plugins-for-kubernetes/pkg/apis/deviceplugin/v1"
	ctr "github.com/intel/intel-device-plugins-for-kubernetes/pkg/controllers"
	dlbctr "github.com/intel/intel-device-plugins-for-kubernetes/pkg/controllers/dlb"
	dsactr "github.com/intel/intel-device-plugins-for-kubernetes/pkg/controllers/dsa"
	fpgactr "github.com/intel/intel-device-plugins-for-kubernetes/pkg/controllers/fpga"
	gpuctr "github.com/intel/intel-device-plugins-for-kubernetes/pkg/controllers/gpu"
	qatctr "github.com/intel/intel-device-plugins-for-kubernetes/pkg/controllers/qat"
	sgxctr "github.com/intel/intel-device-plugins-for-kubernetes/pkg/controllers/sgx"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const (
	timeout  = time.Second * 30
	interval = time.Second * 1
	dlb      = "dlb"
	dsa      = "dsa"
	gpu      = "gpu"
	fpga     = "fpga"
	qat      = "qat"
	sgx      = "sgx"
)

var (
	cfg         *rest.Config
	k8sClient   client.Client
	k8sManager  ctrl.Manager
	testEnv     *envtest.Environment
	ctx         context.Context
	cancel      context.CancelFunc
	err         error
	ns          = metav1.NamespaceSystem
	version     = ctr.ImageMinVersion.String()
	prevVersion = ctr.ImageMinVersion.WithMinor(ctr.ImageMinVersion.Minor() - 1).String()
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t,
		"Controller Suite")
}

var _ = BeforeSuite(func() {
	By("bootstrapping test environment")

	logf.SetLogger(klogr.NewWithOptions(klogr.WithFormat(klogr.FormatKlog)))

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "..", "deployments", "operator", "crd", "bases")},
	}
	cfg, err = testEnv.Start()

	Expect(err == nil && cfg != nil).To(BeTrue())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})

	Expect(err == nil && k8sClient != nil).To(BeTrue())

	Expect(devicepluginv1.AddToScheme(scheme.Scheme)).To(BeNil())

}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")

	Expect(testEnv.Stop()).To(BeNil())
})

var _ = BeforeEach(func() {
	up()
})

var _ = AfterEach(func() {
	down()
})

func up() {
	k8sManager, _ = ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme.Scheme})

	withWebhook := true

	Expect(dlbctr.SetupReconciler(k8sManager, ns, !withWebhook)).To(BeNil())

	Expect(dsactr.SetupReconciler(k8sManager, ns, !withWebhook)).To(BeNil())

	Expect(fpgactr.SetupReconciler(k8sManager, ns, !withWebhook)).To(BeNil())

	Expect(gpuctr.SetupReconciler(k8sManager, ns, !withWebhook)).To(BeNil())

	Expect(qatctr.SetupReconciler(k8sManager, ns, !withWebhook)).To(BeNil())

	Expect(sgxctr.SetupReconciler(k8sManager, ns, !withWebhook)).To(BeNil())

	ctx, cancel = context.WithCancel(context.TODO())

	go func() {
		Expect(k8sManager.Start(ctx)).To(BeNil())
	}()

	time.Sleep(time.Second)
}

func down() {
	time.Sleep(time.Second)

	ctx = context.TODO()

	cancel()
}

func getImage(name string, dpToUpdate client.Object) string {
	switch name {
	case dlb:
		return dpToUpdate.(*devicepluginv1.DlbDevicePlugin).Spec.Image
	case dsa:
		return dpToUpdate.(*devicepluginv1.DsaDevicePlugin).Spec.Image
	case gpu:
		return dpToUpdate.(*devicepluginv1.GpuDevicePlugin).Spec.Image
	case fpga:
		return dpToUpdate.(*devicepluginv1.FpgaDevicePlugin).Spec.Image
	case qat:
		return dpToUpdate.(*devicepluginv1.QatDevicePlugin).Spec.Image
	case sgx:
		return dpToUpdate.(*devicepluginv1.SgxDevicePlugin).Spec.Image
	default:
		return ""
	}
}

func getInitImage(name string, dpToUpdate client.Object) string {
	switch name {
	case gpu:
		return dpToUpdate.(*devicepluginv1.GpuDevicePlugin).Spec.InitImage
	case fpga:
		return dpToUpdate.(*devicepluginv1.FpgaDevicePlugin).Spec.InitImage
	case sgx:
		return dpToUpdate.(*devicepluginv1.SgxDevicePlugin).Spec.InitImage
	default:
		return ""
	}
}

func getIntendedArgs(name string, dpToUpdate client.Object) []string {
	var intendedArgs []string

	switch name {
	case dlb:
		intendedArgs = dlbctr.GetPodArgs(dpToUpdate.(*devicepluginv1.DlbDevicePlugin))
	case dsa:
		intendedArgs = dsactr.GetPodArgs(dpToUpdate.(*devicepluginv1.DsaDevicePlugin))
	case fpga:
		intendedArgs = fpgactr.GetPodArgs(dpToUpdate.(*devicepluginv1.FpgaDevicePlugin))
	case gpu:
		intendedArgs = gpuctr.GetPodArgs(dpToUpdate.(*devicepluginv1.GpuDevicePlugin))
	case sgx:
		intendedArgs = sgxctr.GetPodArgs(dpToUpdate.(*devicepluginv1.SgxDevicePlugin))
	case qat:
		intendedArgs = qatctr.GetPodArgs(dpToUpdate.(*devicepluginv1.QatDevicePlugin))
	}

	return intendedArgs
}

func getIntendedNodeSelector(name string, dpToUpdate client.Object) map[string]string {
	var intendedNodeSelector map[string]string

	switch name {
	case dlb:
		intendedNodeSelector = dpToUpdate.(*devicepluginv1.DlbDevicePlugin).Spec.NodeSelector
	case dsa:
		intendedNodeSelector = dpToUpdate.(*devicepluginv1.DsaDevicePlugin).Spec.NodeSelector
	case gpu:
		intendedNodeSelector = dpToUpdate.(*devicepluginv1.GpuDevicePlugin).Spec.NodeSelector
	case fpga:
		intendedNodeSelector = dpToUpdate.(*devicepluginv1.FpgaDevicePlugin).Spec.NodeSelector
	case qat:
		intendedNodeSelector = dpToUpdate.(*devicepluginv1.QatDevicePlugin).Spec.NodeSelector
	case sgx:
		intendedNodeSelector = dpToUpdate.(*devicepluginv1.SgxDevicePlugin).Spec.NodeSelector
	}

	if len(intendedNodeSelector) == 0 {
		intendedNodeSelector = map[string]string{"kubernetes.io/arch": "amd64"}
	}

	return intendedNodeSelector
}

func getKey(name string) types.NamespacedName {
	return types.NamespacedName{
		Name: name + "deviceplugin-test",
	}
}

func isDaemonSetCreated(name string, fetched client.Object) bool {
	return Eventually(func() bool {
		_ = k8sClient.Get(context.Background(), getKey(name), fetched)
		switch name {
		case dlb:
			return len(fetched.(*devicepluginv1.DlbDevicePlugin).Status.ControlledDaemonSet.UID) > 0
		case dsa:
			return len(fetched.(*devicepluginv1.DsaDevicePlugin).Status.ControlledDaemonSet.UID) > 0
		case gpu:
			return len(fetched.(*devicepluginv1.GpuDevicePlugin).Status.ControlledDaemonSet.UID) > 0
		case fpga:
			return len(fetched.(*devicepluginv1.FpgaDevicePlugin).Status.ControlledDaemonSet.UID) > 0
		case qat:
			return len(fetched.(*devicepluginv1.QatDevicePlugin).Status.ControlledDaemonSet.UID) > 0
		case sgx:
			return len(fetched.(*devicepluginv1.SgxDevicePlugin).Status.ControlledDaemonSet.UID) > 0
		default:
			return false
		}
	}, timeout, interval).Should(BeTrue())
}

func testCreateDevicePluginWithSpec(name string, spec interface{}, fetched client.Object) {
	toCreate := makeDP(name, spec)
	testCreateDevicePlugin(name, toCreate, fetched)
}
func testCreateDevicePlugin(name string, toCreate client.Object, fetched client.Object) {
	Expect(k8sClient.Create(context.Background(), toCreate)).Should(Succeed())
	time.Sleep(time.Second * 5)
	isDaemonSetCreated(name, fetched)
	testUpdateNodeSelector(name, toCreate)
	testUpdateArgs(name, toCreate, fetched)

	if name == gpu || name == sgx {
		testUpdateInitImage(name, toCreate, fetched)
	}
}

func testDelete(name string) {
	var f client.Object

	switch name {
	case dlb:
		f = &devicepluginv1.DlbDevicePlugin{}
	case dsa:
		f = &devicepluginv1.DsaDevicePlugin{}
	case gpu:
		f = &devicepluginv1.GpuDevicePlugin{}
	case fpga:
		f = &devicepluginv1.FpgaDevicePlugin{}
	case qat:
		f = &devicepluginv1.QatDevicePlugin{}
	case sgx:
		f = &devicepluginv1.SgxDevicePlugin{}
	}

	Eventually(func() error {
		_ = k8sClient.Get(context.Background(), getKey(name), f)

		return k8sClient.Delete(context.Background(), f)
	}, timeout, interval).Should(Succeed())

	Eventually(func() error {
		return k8sClient.Get(context.Background(), getKey(name), f)
	}, timeout, interval).ShouldNot(Succeed())
}

func testUpdateArgs(name string, dpToUpdate client.Object, updatedDp client.Object) {
	ds := &apps.DaemonSet{}

	Eventually(func() []string {
		_ = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: "intel-" + name + "-plugin"}, ds)
		return ds.Spec.Template.Spec.Containers[0].Args
	}, timeout, interval).Should(Equal(getIntendedArgs(name, dpToUpdate)))
}

func testUpdateDevicePlugin(dpToUpdate client.Object) {
	Expect(k8sClient.Update(context.Background(), dpToUpdate)).Should(Succeed())
}

func testUpdateImage(name string, dpToUpdate client.Object, updatedDp client.Object) {
	Eventually(func() string {
		_ = k8sClient.Get(context.Background(), getKey(name), updatedDp)
		return getImage(name, updatedDp)
	}, timeout, interval).Should(Equal(getImage(name, dpToUpdate)))
}

func testUpdateInitImage(name string, dpToUpdate client.Object, updatedDp client.Object) {
	var initImage string

	Eventually(func() string {
		_ = k8sClient.Get(context.Background(), getKey(name), updatedDp)
		initImage = getInitImage(name, updatedDp)

		return initImage
	}, timeout, interval).Should(Equal(getInitImage(name, dpToUpdate)))
}

func testUpdateNodeSelector(name string, dpToUpdate client.Object) {
	ds := &apps.DaemonSet{}

	Eventually(func() map[string]string {
		_ = k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: "intel-" + name + "-plugin"}, ds)
		return ds.Spec.Template.Spec.NodeSelector
	}, timeout, interval).Should(Equal(getIntendedNodeSelector(name, dpToUpdate)))
}

func testUpgrade(name string, dp interface{}, pimage, pinitimage *string) {
	down()

	prefix := "intel/intel-" + name
	image0 := prefix + "-plugin:" + prevVersion
	initimage0 := prefix + "-initcontainer:" + prevVersion
	image := prefix + "-plugin:" + version
	initimage := prefix + "-initcontainer:" + version

	*pimage = image

	if pinitimage != nil {
		*pinitimage = initimage
	}

	ds0 := makeDaemonSet(name, image0, initimage0)

	Expect(k8sClient.Create(ctx, ds0)).To(BeNil())

	dp0 := makeDevicePlugin(name, image0, initimage0)

	Expect(k8sClient.Create(ctx, dp0)).To(BeNil())

	up()

	Expect(k8sClient.Get(ctx, types.NamespacedName{Name: name}, dp.(client.Object))).To(BeNil())

	ds := &apps.DaemonSet{}

	Expect(k8sClient.Get(ctx, types.NamespacedName{Namespace: ns, Name: "intel-" + name + "-plugin"}, ds)).To(BeNil())

	Expect(ds.Spec.Template.Spec.Containers[0].Image == image).To(BeTrue())

	if pinitimage != nil {
		Expect(ds.Spec.Template.Spec.InitContainers[0].Image == initimage).To(BeTrue())
	}

	Expect(k8sClient.Delete(ctx, dp.(client.Object))).To(BeNil())
}

func makeDevicePlugin(name, image, initimage string) client.Object {
	var obj client.Object

	switch name {
	case dlb:
		obj = &devicepluginv1.DlbDevicePlugin{
			Spec: devicepluginv1.DlbDevicePluginSpec{
				Image: image,
			},
		}
	case dsa:
		obj = &devicepluginv1.DsaDevicePlugin{
			Spec: devicepluginv1.DsaDevicePluginSpec{
				Image:     image,
				InitImage: initimage,
			},
		}
	case fpga:
		obj = &devicepluginv1.FpgaDevicePlugin{
			Spec: devicepluginv1.FpgaDevicePluginSpec{
				Image:     image,
				InitImage: initimage,
			},
		}
	case gpu:
		obj = &devicepluginv1.GpuDevicePlugin{
			Spec: devicepluginv1.GpuDevicePluginSpec{
				Image:     image,
				InitImage: initimage,
			},
		}
	case qat:
		obj = &devicepluginv1.QatDevicePlugin{
			Spec: devicepluginv1.QatDevicePluginSpec{
				Image: image,
			},
		}
	case sgx:
		obj = &devicepluginv1.SgxDevicePlugin{
			Spec: devicepluginv1.SgxDevicePluginSpec{
				Image:     image,
				InitImage: initimage,
			},
		}
	}

	obj.SetName(name)

	return obj
}

func makeDP(name string, spec interface{}) client.Object {
	switch name {
	case dlb:
		return &devicepluginv1.DlbDevicePlugin{
			ObjectMeta: metav1.ObjectMeta{
				Name: getKey(name).Name,
			},
			Spec: spec.(devicepluginv1.DlbDevicePluginSpec),
		}
	case dsa:
		return &devicepluginv1.DsaDevicePlugin{
			ObjectMeta: metav1.ObjectMeta{
				Name: getKey(name).Name,
			},
			Spec: spec.(devicepluginv1.DsaDevicePluginSpec),
		}
	case fpga:
		return &devicepluginv1.FpgaDevicePlugin{
			ObjectMeta: metav1.ObjectMeta{
				Name: getKey(name).Name,
			},
			Spec: spec.(devicepluginv1.FpgaDevicePluginSpec),
		}
	case gpu:
		return &devicepluginv1.GpuDevicePlugin{
			ObjectMeta: metav1.ObjectMeta{
				Name: getKey(name).Name,
			},
			Spec: spec.(devicepluginv1.GpuDevicePluginSpec),
		}
	case qat:
		return &devicepluginv1.QatDevicePlugin{
			ObjectMeta: metav1.ObjectMeta{
				Name: getKey(name).Name,
			},
			Spec: spec.(devicepluginv1.QatDevicePluginSpec),
		}
	case sgx:
		return &devicepluginv1.SgxDevicePlugin{
			ObjectMeta: metav1.ObjectMeta{
				Name: getKey(name).Name,
			},
			Spec: spec.(devicepluginv1.SgxDevicePluginSpec),
		}
	default:
		return nil
	}
}

func makeDaemonSet(name, image, initimage string) *apps.DaemonSet {
	var ds *apps.DaemonSet

	initcontainerName := "intel-" + name + "-initcontainer"

	switch name {
	case dlb:
		ds = deployments.DLBPluginDaemonSet()
	case dsa:
		ds = deployments.DSAPluginDaemonSet()
		initcontainerName = "intel-idxd-config-initcontainer"
	case gpu:
		ds = deployments.GPUPluginDaemonSet()
	case fpga:
		ds = deployments.FPGAPluginDaemonSet()
	case qat:
		ds = deployments.QATPluginDaemonSet()
	case sgx:
		ds = deployments.SGXPluginDaemonSet()
	}

	ds.ObjectMeta.Namespace = ns
	ds.Spec.Template.Spec.Containers[0].Image = image

	if len(initimage) > 0 {
		ds.Spec.Template.Spec.InitContainers = []corev1.Container{{
			Name:  initcontainerName,
			Image: initimage,
		}}
	}

	yes := true
	ds.OwnerReferences = []metav1.OwnerReference{{
		APIVersion: "deviceplugin.intel.com/v1",
		Kind:       strings.Title(name) + "DevicePlugin",
		Name:       name,
		Controller: &yes,
		UID:        uuid.NewUUID(),
	}}

	return ds
}
