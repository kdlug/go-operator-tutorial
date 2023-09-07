/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	cachev1alpha1 "github.com/kdlug/go-operator-tutorial/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	Expect(os.Setenv("MEMCACHED_IMAGE", "memcached:1.4.36-alpine")).To(Succeed())
	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = cachev1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&MemcachedReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	// anonymous function run as gouroutine which will run controller in background
	go func() {
		defer GinkgoRecover()       // when anonymous function finishes GinkoRecover() will be run
		err = k8sManager.Start(ctx) // runs k8s manager
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = Describe("MemcachedController", func() {
	Context("testing memcache controller", func() {
		var memcached *cachev1alpha1.Memcached
		BeforeEach(func() {
			memcached = getMemcached("default", "test-memcache")
		})

		// Integration tests using It blocks are written here.
		It("should create deployment", func() {
			Expect(k8sClient.Create(ctx, memcached)).To(BeNil()) // create memcache CR
			createdDeploy := &appsv1.Deployment{}
			deployKey := types.NamespacedName{Name: memcached.Name, Namespace: memcached.Namespace}
			// after creating a CR controller should create a Deployment
			// we call testEnv Kubernetes API Server to get the deployment
			// Eventually block is a retry block with a timeout
			// we expect to get deployment in that time
			Eventually(func() bool {
				err := k8sClient.Get(ctx, deployKey, createdDeploy) //
				return err == nil
			}, time.Second*10, time.Millisecond*250).Should(BeTrue())
		})

		It("verify replicas for deployment", func() {
			createdDeploy := &appsv1.Deployment{}
			deployKey := types.NamespacedName{Name: memcached.Name, Namespace: memcached.Namespace}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, deployKey, createdDeploy)
				return err == nil
			}, time.Second*10, time.Millisecond*250).Should(BeTrue())
			Expect(createdDeploy.Spec.Replicas).To(Equal(&memcached.Spec.Size))
		})

		It("should update deployment, once memcached size is changed", func() {
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: memcached.Name, Namespace: memcached.Namespace},
				memcached)).Should(Succeed())
			// update size to 3
			memcached.Spec.Size = 3
			Expect(k8sClient.Update(ctx, memcached)).Should(Succeed())
			Eventually(func() bool {
				k8sClient.Get(ctx,
					types.NamespacedName{Name: memcached.Name, Namespace: memcached.Namespace},
					memcached)
				return memcached.Spec.Size == 3
			}, time.Second*10, time.Millisecond*250).Should(BeTrue())
			createdDeploy := &appsv1.Deployment{}
			deployKey := types.NamespacedName{Name: memcached.Name, Namespace: memcached.Namespace}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, deployKey, createdDeploy)
				return err == nil
			}, time.Second*20, time.Millisecond*250).Should(BeTrue())
			Expect(createdDeploy.Spec.Replicas).To(Equal(&memcached.Spec.Size))
		})

	})
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func getMemcached(namespace string, name string) *cachev1alpha1.Memcached {
	return &cachev1alpha1.Memcached{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: cachev1alpha1.MemcachedSpec{
			Size:          2,
			ContainerPort: 8090,
		},
	}
}
