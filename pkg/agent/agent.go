package agent

import (
	"context"
	"time"

	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/spf13/cobra"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"open-cluster-management.io/addon-framework/pkg/lease"
	"open-cluster-management.io/addon-framework/pkg/version"

	hohoperatorclientset "github.com/stolostron/hub-of-hubs-operator/apis/client/clientset/versioned"
	hohoperatorinformer "github.com/stolostron/hub-of-hubs-operator/apis/client/informers/externalversions"
	hohoperatoragentcontroller "github.com/stolostron/hub-of-hubs-operator/pkg/agent/controller"
)

func NewAgentCommand(addonName string) *cobra.Command {
	o := NewAgentOptions(addonName)
	cmd := controllercmd.
		NewControllerCommandConfig("hub-of-hubs-operator-agent", version.Get(), o.RunAgent).
		NewCommand()
	cmd.Use = "agent"
	cmd.Short = "Start the hub-of-hubs-operator agent"

	o.AddFlags(cmd)
	return cmd
}

// AgentOptions defines the flags for workload agent
type AgentOptions struct {
	HubKubeconfigFile string
	SpokeClusterName  string
	AddonName         string
	AddonNamespace    string
}

// NewWorkloadAgentOptions returns the flags with default value set
func NewAgentOptions(addonName string) *AgentOptions {
	return &AgentOptions{AddonName: addonName}
}

func (o *AgentOptions) AddFlags(cmd *cobra.Command) {
	flags := cmd.Flags()
	// This command only supports reading from config
	flags.StringVar(&o.HubKubeconfigFile, "hub-kubeconfig", o.HubKubeconfigFile, "Location of kubeconfig file to connect to hub cluster.")
	flags.StringVar(&o.SpokeClusterName, "cluster-name", o.SpokeClusterName, "Name of spoke cluster.")
	flags.StringVar(&o.AddonNamespace, "addon-namespace", o.AddonNamespace, "Installation namespace of addon.")
}

// RunAgent starts the controllers on agent to process work from hub.
func (o *AgentOptions) RunAgent(ctx context.Context, controllerContext *controllercmd.ControllerContext) error {
	// build kubeconfig of hub cluster
	hubRestConfig, err := clientcmd.BuildConfigFromFlags("", o.HubKubeconfigFile)
	if err != nil {
		return err
	}

	// build kube client of hub cluster
	hubKubeClient, err := kubernetes.NewForConfig(hubRestConfig)

	// build hub kube informer factory
	hubKubeInformerFactory := informers.NewSharedInformerFactoryWithOptions(hubKubeClient, 10*time.Minute, informers.WithNamespace(o.SpokeClusterName))

	// build kubeclient of managed cluster
	spokeKubeClient, err := kubernetes.NewForConfig(controllerContext.KubeConfig)
	if err != nil {
		return err
	}

	// build spoke kube informer factory
	spokeKubeInformerFactory := informers.NewSharedInformerFactory(spokeKubeClient, 10*time.Minute)

	// build hohoperatorclientset of hub cluster
	hubHoHOperatorClient, err := hohoperatorclientset.NewForConfig(hubRestConfig)
	if err != nil {
		return err
	}

	// build hohoperator informer factory
	hubHoHOperatorInformerFactory := hohoperatorinformer.NewSharedInformerFactory(hubHoHOperatorClient, 10*time.Minute)

	// create an hohoperatoragentcontroller controller
	hohOperatorAgentController := hohoperatoragentcontroller.NewHoHOperatorAgentController(
		o.SpokeClusterName,
		o.AddonNamespace,
		hubKubeClient,
		spokeKubeClient,
		hubHoHOperatorClient,
		hubHoHOperatorInformerFactory.Hubofhubs().V1alpha1().Configs(),
		controllerContext.EventRecorder,
	)

	// create a lease updater
	leaseUpdater := lease.NewLeaseUpdater(
		spokeKubeClient,
		o.AddonName,
		o.AddonNamespace,
	)

	go hubKubeInformerFactory.Start(ctx.Done())
	go hubHoHOperatorInformerFactory.Start(ctx.Done())
	go spokeKubeInformerFactory.Start(ctx.Done())
	go hohOperatorAgentController.Run(ctx, 1)
	go leaseUpdater.Start(ctx)

	<-ctx.Done()
	return nil
}
