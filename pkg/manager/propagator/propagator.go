package deployment

import (
	"context"

	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	clusterclientset "open-cluster-management.io/api/client/cluster/clientset/versioned"

	hohoperatorclientset "github.com/stolostron/hub-of-hubs-operator/apis/client/clientset/versioned"
	hohoperatorv1alpha1informer "github.com/stolostron/hub-of-hubs-operator/apis/client/informers/externalversions/hubofhubs/v1alpha1"
	hohoperatorv1alpha1lister "github.com/stolostron/hub-of-hubs-operator/apis/client/listers/hubofhubs/v1alpha1"
	hohoperatorv1alpha1 "github.com/stolostron/hub-of-hubs-operator/apis/hubofhubs/v1alpha1"
	hohoperatorresourceapply "github.com/stolostron/hub-of-hubs-operator/pkg/resourceapply"
)

const hohOperatorPropagatorFinalizer = "hubofhubs.open-cluster-management.io/operator-propagator-resources-cleanup"

type hohOperatorPropagatorController struct {
	clusterClient           clusterclientset.Interface
	hohOperatorClient       hohoperatorclientset.Interface
	hohOperatorConfigLister hohoperatorv1alpha1lister.ConfigLister
	recorder                events.Recorder
}

func NewHohOperatorPropagatorController(
	clusterClient clusterclientset.Interface,
	hohOperatorClient hohoperatorclientset.Interface,
	hohOperatorConfigInformer hohoperatorv1alpha1informer.ConfigInformer,
	recorder events.Recorder,
) factory.Controller {
	c := &hohOperatorPropagatorController{
		clusterClient:           clusterClient,
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
		WithSync(c.sync).ToController("hub-of-hubs-operator-propagator", recorder)
}

func (c *hohOperatorPropagatorController) sync(ctx context.Context, syncCtx factory.SyncContext) error {
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
			if hohOperatorConfig.Finalizers[i] == hohOperatorPropagatorFinalizer {
				hasFinalizer = true
				break
			}
		}
		if !hasFinalizer {
			hohOperatorConfig.Finalizers = append(hohOperatorConfig.Finalizers, hohOperatorPropagatorFinalizer)
			klog.V(2).Infof("adding finalizer %q to hub-of-hubs-operator config %q/%q", hohOperatorPropagatorFinalizer, namespace, name)
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

	// create the AgentConfig
	hohOperatorAgentConfig := &hohoperatorv1alpha1.AgentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"hub-of-hubs.open-cluster-management.io/managed-by": "hub-of-hubs-operator-manager",
			},
		},
		Spec: hohoperatorv1alpha1.AgentConfigSpec{
			Global: &hohoperatorv1alpha1.AgentGlobalConfig{
				HeartbeatInterval: &hohoperatorv1alpha1.AgentHeartbeatIntervalConfig{
					LeafHub: hohOperatorConfig.Spec.Global.HeartbeatInterval.LeafHub,
				},
			},
			Components: &hohoperatorv1alpha1.AgentComponentsConfig{
				Core: &hohoperatorv1alpha1.AgentCoreConfig{
					LeafHub: hohOperatorConfig.Spec.Components.Core.LeafHub,
				},
				Transport: &hohoperatorv1alpha1.AgentTransportConfig{
					Provider:    hohOperatorConfig.Spec.Components.Transport.Provider,
					SyncService: hohOperatorConfig.Spec.Components.Transport.SyncService,
				},
			},
		},
	}

	// list all the openshift managedclusters
	managedClusterList, err := c.clusterClient.ClusterV1().ManagedClusters().List(ctx, metav1.ListOptions{LabelSelector: "vendor=OpenShift"})
	if err != nil {
		return err
	}

	// propagate the AgentConfig to the openshift managedclusters
	for _, mc := range managedClusterList.Items {
		hohOperatorAgentConfig.SetNamespace(mc.GetName())
		if _, _, err := hohoperatorresourceapply.ApplyAgentConfig(ctx, c.hohOperatorClient.HubofhubsV1alpha1(), c.recorder, hohOperatorAgentConfig); err != nil {
			return err
		}
	}

	return nil
}

func (c *hohOperatorPropagatorController) removeHoHOperatorConfigResources(ctx context.Context, hohOperatorConfig *hohoperatorv1alpha1.Config) error {
	// list all the openshift managedclusters
	managedClusterList, err := c.clusterClient.ClusterV1().ManagedClusters().List(ctx, metav1.ListOptions{LabelSelector: "vendor=OpenShift"})
	if err != nil {
		return err
	}

	// remove the AgentConfig from the openshift managedclusters
	for _, mc := range managedClusterList.Items {
		if err := c.hohOperatorClient.HubofhubsV1alpha1().AgentConfigs(mc.GetName()).Delete(ctx, hohOperatorConfig.GetName(), metav1.DeleteOptions{}); err != nil {
			return err
		}
	}

	return nil
}

func (c *hohOperatorPropagatorController) removeHoHOperatorConfigFinalizer(ctx context.Context, hohOperatorConfig *hohoperatorv1alpha1.Config) error {
	copiedFinalizers := []string{}
	for _, finalizer := range hohOperatorConfig.Finalizers {
		if finalizer == hohOperatorPropagatorFinalizer {
			continue
		}
		copiedFinalizers = append(copiedFinalizers, finalizer)
	}

	if len(hohOperatorConfig.Finalizers) != len(copiedFinalizers) {
		hohOperatorConfig.Finalizers = copiedFinalizers
		klog.V(2).Infof("removing finalizer %q from hub-of-hubs-operator config %q/%q", hohOperatorPropagatorFinalizer, hohOperatorConfig.GetNamespace(), hohOperatorConfig.GetName())
		_, err := c.hohOperatorClient.HubofhubsV1alpha1().Configs(hohOperatorConfig.GetNamespace()).Update(ctx, hohOperatorConfig, metav1.UpdateOptions{})
		return err
	}

	return nil
}
