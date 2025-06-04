package fromhost

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/loft-sh/vcluster/test/framework"
)

var _ = Describe("Verify ingressClass is synced from Host to vCluster", Ordered, func() {
	var (
		f                *framework.Framework
		ingressClass     *networkingv1.IngressClass
		ingressClassName = "my-custom-ingress-class"
	)

	BeforeAll(func() {
		f = framework.DefaultFramework
		ingressClass = &networkingv1.IngressClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: ingressClassName,
				Labels: map[string]string{
					"kubernetes.io/selector-from-host-sync-test": "true",
				},
			},
			Spec: networkingv1.IngressClassSpec{
				Controller: "example.com/custom-ingress-controller",
			},
		}
	})
	AfterAll(func() {
		framework.ExpectNoError(f.HostClient.NetworkingV1().IngressClasses().Delete(f.Context, ingressClassName, metav1.DeleteOptions{}))
	})

	It("Ingress Class is synced from Host to vCluster", func() {
		_, err := f.HostClient.NetworkingV1().IngressClasses().Create(f.Context, ingressClass, metav1.CreateOptions{})
		framework.ExpectNoError(err)
		var ingressClassFromGet *networkingv1.IngressClass
		Eventually(func() bool {
			ingressClassFromGet, err = f.VClusterClient.NetworkingV1().IngressClasses().Get(f.Context, ingressClassName, metav1.GetOptions{})
			return err == nil
		}).WithPolling(time.Second).WithTimeout(framework.PollTimeout).Should(BeTrue())

		Expect(ingressClassFromGet.Name).To(Equal(ingressClassName))
	})

	It("Ingress Class is not synced from Host to vCluster when label is not set", func() {
		ingressClassNotToSync := &networkingv1.IngressClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-custom-ingress-class-not-to-sync",
			},
			Spec: networkingv1.IngressClassSpec{
				Controller: "example.com/custom-ingress-controller",
			},
		}

		_, err := f.HostClient.NetworkingV1().IngressClasses().Create(f.Context, ingressClassNotToSync, metav1.CreateOptions{})
		Expect(err).NotTo(HaveOccurred())

		Eventually(func() bool {
			_, err = f.VClusterClient.NetworkingV1().IngressClasses().Get(f.Context, ingressClassNotToSync.Name, metav1.GetOptions{})
			return err != nil
		}).WithPolling(time.Second).WithTimeout(framework.PollTimeout).Should(BeTrue())

		Expect(err.Error()).To(ContainSubstring("not found"))
	})
})
