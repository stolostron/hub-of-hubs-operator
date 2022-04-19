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

const hohOperatorAgentFinalizer = "hubofhubs.open-cluster-management.io/operator-agent-resources-cleanup"

type hohOperatorAgentController struct {
	clusterName                  string
	addonNamespace               string
	hubKubeClient                kubernetes.Interface
	spokeKubeClient              kubernetes.Interface
	hohOperatorClient            hohoperatorclientset.Interface
	hohOperatorAgentConfigLister hohoperatorv1alpha1lister.AgentConfigLister
	recorder                     events.Recorder
}

func NewHohOperatorAgentController(
	clusterName string,
	addonNamespace string,
	hubKubeClient kubernetes.Interface,
	spokeKubeClient kubernetes.Interface,
	hohOperatorClient hohoperatorclientset.Interface,
	hohOperatorAgentConfigInformer hohoperatorv1alpha1informer.AgentConfigInformer,
	recorder events.Recorder,
) factory.Controller {
	c := &hohOperatorAgentController{
		clusterName:                  clusterName,
		addonNamespace:               addonNamespace,
		hubKubeClient:                hubKubeClient,
		spokeKubeClient:              spokeKubeClient,
		hohOperatorClient:            hohOperatorClient,
		hohOperatorAgentConfigLister: hohOperatorAgentConfigInformer.Lister(),
		recorder:                     recorder,
	}
	return factory.New().
		WithInformersQueueKeyFunc(
			func(obj runtime.Object) string {
				key, _ := cache.MetaNamespaceKeyFunc(obj)
				return key
			}, hohOperatorAgentConfigInformer.Informer()).
		WithSync(c.sync).ToController("hub-of-hubs-operator-agent-controller", recorder)
}

func (c *hohOperatorAgentController) sync(ctx context.Context, syncCtx factory.SyncContext) error {
	key := syncCtx.QueueKey()
	klog.V(2).Infof("Reconciling hub-of-hubs-operator agentconfig %q", key)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		// ignore addon whose key is not in format: namespace/name
		return nil
	}

	hohOperatorAgentConfig, err := c.hohOperatorAgentConfigLister.AgentConfigs(namespace).Get(name)
	switch {
	case errors.IsNotFound(err):
		return nil
	case err != nil:
		return err
	}

	hohOperatorAgentConfig = hohOperatorAgentConfig.DeepCopy()
	if hohOperatorAgentConfig.DeletionTimestamp.IsZero() {
		hasFinalizer := false
		for i := range hohOperatorAgentConfig.Finalizers {
			if hohOperatorAgentConfig.Finalizers[i] == hohOperatorAgentFinalizer {
				hasFinalizer = true
				break
			}
		}
		if !hasFinalizer {
			hohOperatorAgentConfig.Finalizers = append(hohOperatorAgentConfig.Finalizers, hohOperatorAgentFinalizer)
			klog.V(2).Infof("adding finalizer %q to hub-of-hubs-operator agentconfig %q/%q", hohOperatorAgentFinalizer, namespace, name)
			_, err := c.hohOperatorClient.HubofhubsV1alpha1().AgentConfigs(namespace).Update(ctx, hohOperatorAgentConfig, metav1.UpdateOptions{})
			return err
		}
	}

	// remove hohOperatorAgentConfig related resources after hohOperatorAgentConfig is deleted
	if !hohOperatorAgentConfig.DeletionTimestamp.IsZero() {
		if err := c.removeHoHOperatorAgentConfigResources(ctx, hohOperatorAgentConfig); err != nil {
			return err
		}
		return c.removeHoHOperatorAgentConfigFinalizer(ctx, hohOperatorAgentConfig)
	}

	// TODO(morvencao): add config agent logic here

	return nil
}

func (c *hohOperatorAgentController) removeHoHOperatorAgentConfigResources(ctx context.Context, hohOperatorAgentConfig *hohoperatorv1alpha1.AgentConfig) error {
	// TODO(morvencao): add config agent resources remove logic here

	return nil
}

func (c *hohOperatorAgentController) removeHoHOperatorAgentConfigFinalizer(ctx context.Context, hohOperatorAgentConfig *hohoperatorv1alpha1.AgentConfig) error {
	copiedFinalizers := []string{}
	for _, finalizer := range hohOperatorAgentConfig.Finalizers {
		if finalizer == hohOperatorAgentFinalizer {
			continue
		}
		copiedFinalizers = append(copiedFinalizers, finalizer)
	}

	if len(hohOperatorAgentConfig.Finalizers) != len(copiedFinalizers) {
		hohOperatorAgentConfig.Finalizers = copiedFinalizers
		klog.V(2).Infof("removing finalizer %q from hub-of-hubs-operator agentconfig %q/%q", hohOperatorAgentFinalizer, hohOperatorAgentConfig.GetNamespace(), hohOperatorAgentConfig.GetName())
		_, err := c.hohOperatorClient.HubofhubsV1alpha1().AgentConfigs(hohOperatorAgentConfig.GetNamespace()).Update(ctx, hohOperatorAgentConfig, metav1.UpdateOptions{})
		return err
	}

	return nil
}
