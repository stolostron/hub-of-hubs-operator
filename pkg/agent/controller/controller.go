package deploy

import (
	"context"

	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	hohoperatorclientset "github.com/stolostron/hub-of-hubs-operator/apis/client/clientset/versioned"
	hohoperatorv1alpha1informer "github.com/stolostron/hub-of-hubs-operator/apis/client/informers/externalversions/hubofhubs/v1alpha1"
	hohoperatorv1alpha1lister "github.com/stolostron/hub-of-hubs-operator/apis/client/listers/hubofhubs/v1alpha1"
	hohoperatorv1alpha1 "github.com/stolostron/hub-of-hubs-operator/apis/hubofhubs/v1alpha1"
)

const hoHOperatorAgentFinalizer = "hubofhubs.open-cluster-management.io/hoh-operator-agent-resources-cleanup"

type hohOperatorAgentController struct {
	clusterName             string
	addonNamespace          string
	hubKubeClient           kubernetes.Interface
	spokeKubeClient         kubernetes.Interface
	hohOperatorClient       hohoperatorclientset.Interface
	hohOperatorConfigLister hohoperatorv1alpha1lister.ConfigLister
	recorder                events.Recorder
}

func NewHoHOperatorAgentController(
	clusterName string,
	addonNamespace string,
	hubKubeClient kubernetes.Interface,
	spokeKubeClient kubernetes.Interface,
	hohOperatorClient hohoperatorclientset.Interface,
	hohOperatorConfigInformer hohoperatorv1alpha1informer.ConfigInformer,
	recorder events.Recorder,
) factory.Controller {
	c := &hohOperatorAgentController{
		clusterName:             clusterName,
		addonNamespace:          addonNamespace,
		hubKubeClient:           hubKubeClient,
		spokeKubeClient:         spokeKubeClient,
		hohOperatorClient:       hohOperatorClient,
		hohOperatorConfigLister: hohOperatorConfigInformer.Lister(),
		recorder:                recorder,
	}
	return factory.New().
		WithInformersQueueKeyFunc(
			func(obj runtime.Object) string {
				key, _ := cache.MetaNamespaceKeyFunc(obj)
				return key
			}, hohOperatorConfigInformer.Informer()).
		WithSync(c.sync).ToController("hub-of-hubs-operator-agent-controller", recorder)
}

func (c *hohOperatorAgentController) sync(ctx context.Context, syncCtx factory.SyncContext) error {
	key := syncCtx.QueueKey()
	klog.V(2).Infof("Reconciling hub-of-hubs-operator config %q", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		// ignore addon whose key is not in format: namespace/name
		return nil
	}

	hohOperatorConfig, err := c.hohOperatorConfigLister.Configs(namespace).Get(name)
	switch {
	case errors.IsNotFound(err):
		return nil
	case err != nil:
		return err
	}

	hohOperatorConfig = hohOperatorConfig.DeepCopy()
	if hohOperatorConfig.DeletionTimestamp.IsZero() {
		hasFinalizer := false
		for i := range hohOperatorConfig.Finalizers {
			if hohOperatorConfig.Finalizers[i] == hoHOperatorAgentFinalizer {
				hasFinalizer = true
				break
			}
		}
		if !hasFinalizer {
			hohOperatorConfig.Finalizers = append(hohOperatorConfig.Finalizers, hoHOperatorAgentFinalizer)
			klog.V(2).Infof("adding finalizer %q to hub-of-hubs-operator config %q/%q", hoHOperatorAgentFinalizer, namespace, name)
			_, err := c.hohOperatorClient.HubofhubsV1alpha1().Configs(namespace).Update(ctx, hohOperatorConfig, metav1.UpdateOptions{})
			return err
		}
	}

	// remove hohOperatorConfig related resources after hohOperatorConfig is deleted
	if !hohOperatorConfig.DeletionTimestamp.IsZero() {
		if err := c.removeHoHOperatorConfigResources(ctx, hohOperatorConfig); err != nil {
			return err
		}
		return c.removeHoHOperatorConfigFinalizer(ctx, hohOperatorConfig)
	}

	// TODO(morvencao): add config agent logic here

	return nil
}

func (c *hohOperatorAgentController) removeHoHOperatorConfigResources(ctx context.Context, hohOperatorConfig *hohoperatorv1alpha1.Config) error {
	// TODO(morvencao): add config agent resources remove logic here

	return nil
}

func (c *hohOperatorAgentController) removeHoHOperatorConfigFinalizer(ctx context.Context, hohOperatorConfig *hohoperatorv1alpha1.Config) error {
	copiedFinalizers := []string{}
	for _, finalizer := range hohOperatorConfig.Finalizers {
		if finalizer == hoHOperatorAgentFinalizer {
			continue
		}
		copiedFinalizers = append(copiedFinalizers, finalizer)
	}

	if len(hohOperatorConfig.Finalizers) != len(copiedFinalizers) {
		hohOperatorConfig.Finalizers = copiedFinalizers
		klog.V(2).Infof("removing finalizer %q from hub-of-hubs-operator config %q/%q", hoHOperatorAgentFinalizer, hohOperatorConfig.GetNamespace(), hohOperatorConfig.GetName())
		_, err := c.hohOperatorClient.HubofhubsV1alpha1().Configs(hohOperatorConfig.GetNamespace()).Update(ctx, hohOperatorConfig, metav1.UpdateOptions{})
		return err
	}

	return nil
}
