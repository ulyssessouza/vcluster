package fromhost

import (
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/loft-sh/vcluster/test/framework"
)

var _ = ginkgo.Describe("Verify storageClass is synced from Host to vCluster", ginkgo.Ordered, func() {
	var (
		f                *framework.Framework
		storageClass     *storagev1.StorageClass
		storageClassName = "my-custom-storage-class"
	)

	ginkgo.BeforeAll(func() {
		f = framework.DefaultFramework
		storageClass = &storagev1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: storageClassName,
				Labels: map[string]string{
					"kubernetes.io/selector-from-host-sync-test": "on",
				},
			},
			Provisioner: "kubernetes.io/no-provisioner",
		}
	})
	ginkgo.AfterAll(func() {
		framework.ExpectNoError(f.HostClient.StorageV1().StorageClasses().Delete(f.Context, storageClassName, metav1.DeleteOptions{}))
	})

	ginkgo.It("Storage Class is synced from Host to vCluster", func() {
		_, err := f.HostClient.StorageV1().StorageClasses().Create(f.Context, storageClass, metav1.CreateOptions{})
		framework.ExpectNoError(err)
		var storageClass1 *storagev1.StorageClass
		gomega.Eventually(func() bool {
			storageClass1, err = f.VClusterClient.StorageV1().StorageClasses().Get(f.Context, storageClassName, metav1.GetOptions{})
			return err == nil
		}).WithPolling(time.Second).
			WithTimeout(framework.PollTimeout).
			Should(gomega.BeTrue())

		gomega.Expect(storageClass1.Name).To(gomega.Equal(storageClassName))
	})

})
