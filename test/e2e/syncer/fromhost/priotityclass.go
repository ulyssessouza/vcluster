package fromhost

import (
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	v1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/loft-sh/vcluster/test/framework"
)

var _ = ginkgo.Describe("Verify priorityClass is synced from Host to vCluster", ginkgo.Ordered, func() {
	var (
		f                 *framework.Framework
		priorityClass     *v1.PriorityClass
		priorityClassName = "my-custom-priority-class"
	)

	ginkgo.BeforeAll(func() {
		f = framework.DefaultFramework
		priorityClass = &v1.PriorityClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: priorityClassName,
				Labels: map[string]string{
					// Can name any value other than the negative list of strings
					"kubernetes.io/selector-from-host-sync-test": "positive",
				},
			},
			Value: 42,
		}
	})
	ginkgo.AfterAll(func() {
		framework.ExpectNoError(f.HostClient.SchedulingV1().PriorityClasses().Delete(f.Context, priorityClassName, metav1.DeleteOptions{}))
	})

	ginkgo.It("Priority Class is synced from Host to vCluster", func() {
		_, err := f.HostClient.SchedulingV1().PriorityClasses().Create(f.Context, priorityClass, metav1.CreateOptions{})
		framework.ExpectNoError(err)
		var priorityClass1 *v1.PriorityClass
		gomega.Eventually(func() bool {
			priorityClass1, err = f.VClusterClient.SchedulingV1().PriorityClasses().Get(f.Context, priorityClassName, metav1.GetOptions{})
			return err == nil
		}).WithPolling(time.Second).
			WithTimeout(framework.PollTimeout).
			Should(gomega.BeTrue())

		gomega.Expect(priorityClass1.Name).To(gomega.Equal(priorityClassName))
	})

})
