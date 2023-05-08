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

package qat

import (
	"context"
	"path/filepath"
	"time"

	"github.com/intel/intel-device-plugins-for-kubernetes/test/e2e/utils"
	"github.com/onsi/ginkgo/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kubernetes/test/e2e/framework"
	e2edebug "k8s.io/kubernetes/test/e2e/framework/debug"
	e2ekubectl "k8s.io/kubernetes/test/e2e/framework/kubectl"
	e2epod "k8s.io/kubernetes/test/e2e/framework/pod"
	admissionapi "k8s.io/pod-security-admission/api"
)

const (
	qatPluginKustomizationYaml = "deployments/qat_plugin/overlays/e2e/kustomization.yaml"
	compressTestYaml           = "deployments/qat_dpdk_app/test-compress1/kustomization.yaml"
	cryptoTestYaml             = "deployments/qat_dpdk_app/test-crypto1/kustomization.yaml"
	cryptoTestGen4Yaml         = "deployments/qat_dpdk_app/test-crypto1-gen4/kustomization.yaml"
	compressTestGen4Yaml       = "deployments/qat_dpdk_app/test-compress1-gen4/kustomization.yaml"
)

func init() {
	ginkgo.Describe("QAT plugin in DPDK mode", describeQatDpdkPlugin)
}

func describeQatDpdkPlugin() {
	f := framework.NewDefaultFramework("qatplugindpdk")
	f.NamespacePodSecurityEnforceLevel = admissionapi.LevelPrivileged

	kustomizationPath, err := utils.LocateRepoFile(qatPluginKustomizationYaml)
	if err != nil {
		framework.Failf("unable to locate %q: %v", qatPluginKustomizationYaml, err)
	}

	compressTestYamlPath, err := utils.LocateRepoFile(compressTestYaml)
	if err != nil {
		framework.Failf("unable to locate %q: %v", compressTestYaml, err)
	}

	cryptoTestYamlPath, err := utils.LocateRepoFile(cryptoTestYaml)
	if err != nil {
		framework.Failf("unable to locate %q: %v", cryptoTestYaml, err)
	}

	cryptoTestGen4YamlPath, err := utils.LocateRepoFile(cryptoTestGen4Yaml)
	if err != nil {
		framework.Failf("unable to locate %q: %v", cryptoTestGen4Yaml, err)
	}

	compressTestGen4YamlPath, err := utils.LocateRepoFile(compressTestGen4Yaml)
	if err != nil {
		framework.Failf("unable to locate %q: %v", cryptoTestGen4Yaml, err)
	}

	var dpPodName string

	var resourceName v1.ResourceName

	ginkgo.JustBeforeEach(func() {
		ginkgo.By("deploying QAT plugin in DPDK mode")
		e2ekubectl.RunKubectlOrDie(f.Namespace.Name, "apply", "-k", filepath.Dir(kustomizationPath))

		ginkgo.By("waiting for QAT plugin's availability")
		podList, err := e2epod.WaitForPodsWithLabelRunningReady(f.ClientSet, f.Namespace.Name,
			labels.Set{"app": "intel-qat-plugin"}.AsSelector(), 1 /* one replica */, 100*time.Second)
		if err != nil {
			e2edebug.DumpAllNamespaceInfo(f.ClientSet, f.Namespace.Name)
			e2ekubectl.LogFailedContainers(f.ClientSet, f.Namespace.Name, framework.Logf)
			framework.Failf("unable to wait for all pods to be running and ready: %v", err)
		}
		dpPodName = podList.Items[0].Name

		ginkgo.By("checking QAT plugin's securityContext")
		if err := utils.TestPodsFileSystemInfo(podList.Items); err != nil {
			framework.Failf("container filesystem info checks failed: %v", err)
		}

		ginkgo.By("checking if the resource is allocatable")
		if err := utils.WaitForNodesWithResource(f.ClientSet, resourceName, 30*time.Second); err != nil {
			framework.Failf("unable to wait for nodes to have positive allocatable resource: %v", err)
		}
	})

	ginkgo.AfterEach(func() {
		ginkgo.By("undeploying QAT plugin")
		e2ekubectl.RunKubectlOrDie(f.Namespace.Name, "delete", "-k", filepath.Dir(kustomizationPath))
		if err := e2epod.WaitForPodNotFoundInNamespace(f.ClientSet, dpPodName, f.Namespace.Name, 30*time.Second); err != nil {
			framework.Failf("failed to terminate pod: %v", err)
		}
	})

	ginkgo.Context("When QAT Gen4 resources are available with crypto (cy) services enabled", func() {
		// This BeforeEach runs even before the JustBeforeEach above.
		ginkgo.BeforeEach(func() {
			ginkgo.By("creating a configMap before plugin gets deployed")
			e2ekubectl.RunKubectlOrDie(f.Namespace.Name, "create", "configmap", "--from-literal", "qat.conf=ServicesEnabled=sym;asym", "qat-config")

			ginkgo.By("setting resourceName for cy services")
			resourceName = "qat.intel.com/cy"
		})

		ginkgo.It("deploys a crypto pod (openssl) requesting QAT resources", func() {
			runCpaSampleCode(f, "4", resourceName)
		})

		ginkgo.It("deploys a crypto pod (dpdk crypto-perf) requesting QAT resources", func() {
			ginkgo.By("submitting a crypto pod requesting QAT resources")
			e2ekubectl.RunKubectlOrDie(f.Namespace.Name, "apply", "-k", filepath.Dir(cryptoTestGen4YamlPath))

			ginkgo.By("waiting the crypto pod to finish successfully")
			e2epod.NewPodClient(f).WaitForSuccess("qat-dpdk-test-crypto-perf-tc1-gen4", 300*time.Second)

			output, _ := e2epod.GetPodLogs(f.ClientSet, f.Namespace.Name, "qat-dpdk-test-crypto-perf-tc1-gen4", "crypto-perf")

			framework.Logf("crypto-perf output:\n %s", output)
		})
	})

	ginkgo.Context("When QAT Gen4 resources are available with compress (dc) services enabled", func() {
		// This BeforeEach runs even before the JustBeforeEach above.
		ginkgo.BeforeEach(func() {
			ginkgo.By("creating a configMap before plugin gets deployed")
			e2ekubectl.RunKubectlOrDie(f.Namespace.Name, "create", "configmap", "--from-literal", "qat.conf=ServicesEnabled=dc", "qat-config")

			ginkgo.By("setting resourceName for dc services")
			resourceName = "qat.intel.com/dc"
		})

		ginkgo.It("deploys a compress pod (openssl) requesting QAT resources", func() {
			runCpaSampleCode(f, "32", resourceName)
		})

		ginkgo.It("deploys a compress pod (dpdk compress-perf) requesting QAT resources", func() {
			ginkgo.By("submitting a compress pod requesting QAT resources")
			e2ekubectl.RunKubectlOrDie(f.Namespace.Name, "apply", "-k", filepath.Dir(compressTestGen4YamlPath))

			ginkgo.By("waiting the crypto pod to finish successfully")
			e2epod.NewPodClient(f).WaitForSuccess("qat-dpdk-test-compress-perf-tc1-gen4", 300*time.Second)

			output, _ := e2epod.GetPodLogs(f.ClientSet, f.Namespace.Name, "qat-dpdk-test-compress-perf-tc1-gen4", "compress-perf")

			framework.Logf("compress-perf output:\n %s", output)
		})
	})

	ginkgo.Context("When QAT Gen2 resources are available", func() {
		ginkgo.BeforeEach(func() {
			ginkgo.By("setting resourceName for Gen2 resources")
			resourceName = "qat.intel.com/generic"
		})

		ginkgo.It("deploys a crypto pod requesting QAT resources", func() {
			ginkgo.By("submitting a crypto pod requesting QAT resources")
			e2ekubectl.RunKubectlOrDie(f.Namespace.Name, "apply", "-k", filepath.Dir(cryptoTestYamlPath))

			ginkgo.By("waiting the crypto pod to finish successfully")
			e2epod.NewPodClient(f).WaitForSuccess("qat-dpdk-test-crypto-perf-tc1", 60*time.Second)
		})

		ginkgo.It("deploys a compress pod requesting QAT resources", func() {
			ginkgo.By("submitting a compress pod requesting QAT resources")
			e2ekubectl.RunKubectlOrDie(f.Namespace.Name, "apply", "-k", filepath.Dir(compressTestYamlPath))

			ginkgo.By("waiting the compress pod to finish successfully")
			e2epod.NewPodClient(f).WaitForSuccess("qat-dpdk-test-compress-perf-tc1", 60*time.Second)
		})
	})
}

func runCpaSampleCode(f *framework.Framework, runTests string, resourceName v1.ResourceName) {
	ginkgo.By("submitting a pod requesting QAT" + resourceName.String() + "resources")
	podSpec := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "openssl-qat-engine"},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            "openssl-qat-engine",
					Image:           "intel/openssl-qat-engine:devel",
					ImagePullPolicy: "IfNotPresent",
					Command:         []string{"cpa_sample_code", "runTests=" + runTests, "signOfLife=1"},
					SecurityContext: &v1.SecurityContext{
						Capabilities: &v1.Capabilities{
							Add: []v1.Capability{"IPC_LOCK"}},
					},
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{resourceName: resource.MustParse("1")},
						Limits:   v1.ResourceList{resourceName: resource.MustParse("1")},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}
	pod, err := f.ClientSet.CoreV1().Pods(f.Namespace.Name).Create(context.TODO(), podSpec, metav1.CreateOptions{})
	framework.ExpectNoError(err, "pod Create API error")

	ginkgo.By("waiting the cpa_sample_code pod for the resource" + resourceName.String() + "to finish successfully")
	e2epod.NewPodClient(f).WaitForSuccess(pod.ObjectMeta.Name, 300*time.Second)

	output, _ := e2epod.GetPodLogs(f.ClientSet, f.Namespace.Name, pod.ObjectMeta.Name, pod.Spec.Containers[0].Name)

	framework.Logf("cpa_sample_code output:\n %s", output)
}
