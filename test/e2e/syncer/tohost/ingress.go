package tohost

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/loft-sh/vcluster/test/framework"
)

var _ = Describe("Verify ingress is synced from vCluster to Host", Ordered, func() {
	var (
		f                      *framework.Framework
		ingressClassToSync     *networkingv1.IngressClass
		ingressClassNameToSync = "my-custom-ingress-class"

		ingressClassNotToSync     *networkingv1.IngressClass
		ingressClassNameNotToSync = "my-custom-ingress-class-not-to-sync"

		vClusterNamespace *corev1.Namespace
	)

	BeforeAll(func() {
		f = framework.DefaultFramework

		ingressClassToSync = &networkingv1.IngressClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: ingressClassNameToSync,
				Labels: map[string]string{
					"kubernetes.io/selector-from-host-sync-test": "true",
				},
			},
			Spec: networkingv1.IngressClassSpec{
				Controller: "example.com/custom-ingress-controller",
			},
		}

		ingressClassNotToSync = &networkingv1.IngressClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: ingressClassNameNotToSync,
				Labels: map[string]string{
					"kubernetes.io/selector-from-host-sync-test": "false",
				},
			},
			Spec: networkingv1.IngressClassSpec{
				Controller: "example.com/custom-ingress-controller",
			},
		}

		vClusterNamespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: f.VClusterNamespace,
			},
		}
	})
	AfterAll(func() {
		framework.ExpectNoError(f.HostClient.NetworkingV1().IngressClasses().Delete(f.Context, ingressClassNameToSync, metav1.DeleteOptions{}))
		framework.ExpectNoError(f.HostClient.NetworkingV1().IngressClasses().Delete(f.Context, ingressClassNameNotToSync, metav1.DeleteOptions{}))
	})

	It("Ingress Class is synced from Host to vCluster", func() {
		var ingressClassFromGet *networkingv1.IngressClass

		_, err := f.HostClient.NetworkingV1().IngressClasses().Create(f.Context, ingressClassToSync, metav1.CreateOptions{})
		framework.ExpectNoError(err)

		_, err = f.HostClient.NetworkingV1().IngressClasses().Create(f.Context, ingressClassNotToSync, metav1.CreateOptions{})
		framework.ExpectNoError(err)

		Eventually(func() bool {
			ingressClassFromGet, err = f.VClusterClient.NetworkingV1().IngressClasses().Get(f.Context, ingressClassNameToSync, metav1.GetOptions{})
			return err == nil
		}).WithPolling(time.Second).WithTimeout(framework.PollTimeout).Should(BeTrue())
		Expect(ingressClassFromGet.Name).To(Equal(ingressClassNameToSync))

		ingressClassFromGet, err = f.VClusterClient.NetworkingV1().IngressClasses().Get(f.Context, ingressClassNameNotToSync, metav1.GetOptions{})
		framework.ExpectError(err)
	})

	It("Ingress is synced from Host to vCluster", func() {
		var ingressFromGet *networkingv1.Ingress
		var ingressNameToSync = "my-custom-ingress"
		var ingressNameNotToSync = ingressNameToSync + "-not-to-sync"

		_, err := f.VClusterClient.CoreV1().Namespaces().Create(f.Context, vClusterNamespace, metav1.CreateOptions{})
		framework.ExpectNoError(err)

		prefix := networkingv1.PathTypePrefix
		rules := []networkingv1.IngressRule{
			{
				Host: "example.com",
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &prefix,
								Backend: networkingv1.IngressBackend{
									Service: &networkingv1.IngressServiceBackend{
										Name: "web",
										Port: networkingv1.ServiceBackendPort{
											Number: 8080,
										},
									},
								},
							},
						},
					},
				},
			},
		}

		ingressToSync := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name: ingressNameToSync,
			},
			Spec: networkingv1.IngressSpec{
				IngressClassName: &ingressClassNameToSync,
				Rules:            rules,
			},
		}

		ingressNotToSync := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name: ingressNameNotToSync,
			},
			Spec: networkingv1.IngressSpec{
				IngressClassName: &ingressClassNameNotToSync,
				Rules:            rules,
			},
		}

		_, err = f.VClusterClient.NetworkingV1().Ingresses(f.VClusterNamespace).Create(f.Context, ingressToSync, metav1.CreateOptions{})
		framework.ExpectNoError(err)

		_, err = f.VClusterClient.NetworkingV1().Ingresses(f.VClusterNamespace).Create(f.Context, ingressNotToSync, metav1.CreateOptions{})
		framework.ExpectNoError(err)

		Eventually(func() bool {
			ingressFromGet, err = f.HostClient.NetworkingV1().Ingresses(f.VClusterNamespace).Get(f.Context, getHostName(ingressNameToSync), metav1.GetOptions{})
			return err == nil
		}).WithPolling(time.Second).WithTimeout(framework.PollTimeout).Should(BeTrue())
		Expect(ingressFromGet.Name).To(Equal(getHostName(ingressNameToSync)))

		_, err = f.HostClient.NetworkingV1().Ingresses(f.VClusterNamespace).Get(f.Context, getHostName(ingressNameNotToSync), metav1.GetOptions{})
		framework.ExpectNotFound(err) // Cannot find the ingress because it should not be synced
	})
})

func getHostName(name string) string {
	return name + "-x-vcluster-x-vcluster"
}
