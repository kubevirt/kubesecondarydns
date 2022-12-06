package tests

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "kubevirt.io/api/core/v1"

	networkv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
)

const timeout = 3 * time.Minute
const pollingInterval = 5 * time.Second
const testNamespacePrefix = "secondary-test"
const dnsPort = "31111"
const dnsIP = "127.0.0.1" // Forwarded to the node port - https://github.com/kubevirt/kubevirtci/pull/867
const domain = "vm.secondary.io"

var testNamespace string

var _ = Describe("Virtual Machines Startup", func() {
	BeforeEach(func() {
		By("Creating test namespace")
		testNamespace = randName(testNamespacePrefix)
		ns := k8sv1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: testNamespace,
			},
		}
		err := getClient().Create(context.Background(), &ns)
		Expect(err).ToNot(HaveOccurred(), "couldn't create namespace for the test")
	})

	AfterEach(func() {
		By("Removing the test namespace")
		namespace := &k8sv1.Namespace{}
		err := getClient().Get(context.Background(), types.NamespacedName{Name: testNamespace}, namespace)
		Expect(err).ToNot(HaveOccurred())
		err = getClient().Delete(context.Background(), namespace)
		Expect(err).ToNot(HaveOccurred())
		Eventually(func() error {
			return getClient().Get(context.Background(), types.NamespacedName{Name: testNamespace}, namespace)
		}, 240*time.Second, 1*time.Second).Should(SatisfyAll(HaveOccurred(), WithTransform(errors.IsNotFound, BeTrue())), fmt.Sprintf("should successfully delete namespace '%s'", testNamespace))
	})

	Context("with one secondary and no default interface", func() {
		const interfaceName = "nic1"
		var vmiName string
		var vmi *v1.VirtualMachineInstance
		BeforeEach(func() {
			By("Creating a NetworkAttachmentDefinition")
			err := createNetworkAttachmentDefinition(testNamespace, "ptp-conf")
			Expect(err).ToNot(HaveOccurred(), "Should successfully create the NAD")
		})
		BeforeEach(func() {
			By("Creating a VirtualMachineInstance")
			interfaces := []v1.Interface{{Name: interfaceName, InterfaceBindingMethod: v1.InterfaceBindingMethod{Bridge: &v1.InterfaceBridge{}}}}
			networks := []v1.Network{
				{Name: interfaceName, NetworkSource: v1.NetworkSource{
					Multus: &v1.MultusNetwork{NetworkName: "ptp-conf"},
				}},
			}
			vmi := CreateVmiObject("vmi-sec", testNamespace, interfaces, networks)
			vmiName = vmi.Name
			err := getClient().Create(context.Background(), vmi)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should get the correct IP when nslookup to the secondary interface FQDN", func() {
			By("Invoking nslookup on the VMI interface FQDN")
			Eventually(func() bool {
				vmi = &v1.VirtualMachineInstance{}
				err := getClient().Get(context.Background(), types.NamespacedName{Namespace: testNamespace, Name: vmiName}, vmi)
				Expect(err).ToNot(HaveOccurred())
				return len(vmi.Status.Interfaces) == 1 && vmi.Status.Interfaces[0].IP != ""
			}, timeout, pollingInterval).Should(BeTrue(), "failed to get VMI interface IP")
			vmiIp := vmi.Status.Interfaces[0].IP

			var nslookupOutput []byte
			Eventually(func() error {
				var nslookupErr error
				nslookupOutput, nslookupErr = exec.Command("nslookup", fmt.Sprintf("-port=%s", dnsPort), fmt.Sprintf("%s.%s.%s.%s", interfaceName, vmiName, testNamespace, domain), dnsIP).CombinedOutput()
				return nslookupErr
			}, time.Minute, pollingInterval).ShouldNot(HaveOccurred(), fmt.Sprintf("nslookup failed. Output - %s", nslookupOutput))

			By("Comparing the VirtualMachineInstance IP to the nslookup result")
			Expect(nslookupOutput).To(ContainSubstring(fmt.Sprintf("Address: %s\n", vmiIp)), fmt.Sprintf("nsloookup doesn't return the VMI IP address - %s. nslookup output - %s", vmiIp, nslookupOutput))
		})
	})
})

func createNetworkAttachmentDefinition(namespace, name string) error {
	nad := networkv1.NetworkAttachmentDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: networkv1.NetworkAttachmentDefinitionSpec{
			Config: `{
      "cniVersion": "0.3.1",
      "name": "mynet",
      "plugins": [
        {
          "type": "ptp",
          "ipam": {
		      "type": "static",
		      "addresses": [
			      {
				       "address": "10.10.0.5/24",
				       "gateway": "10.10.0.254"
			      }
		      ]
	      }
        }
      ]
}`,
		},
	}
	return getClient().Create(context.Background(), &nad)
}
