package manager

import (
	"context"
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/openshift/library-go/pkg/assets"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilrand "k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"open-cluster-management.io/addon-framework/pkg/addonfactory"
	"open-cluster-management.io/addon-framework/pkg/addonmanager"
	"open-cluster-management.io/addon-framework/pkg/agent"
	"open-cluster-management.io/addon-framework/pkg/utils"
	"open-cluster-management.io/addon-framework/pkg/version"
	addonapiv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	clusterclientset "open-cluster-management.io/api/client/cluster/clientset/versioned"
	clusterv1 "open-cluster-management.io/api/cluster/v1"

	hohoperatorclientset "github.com/stolostron/hub-of-hubs-operator/apis/client/clientset/versioned"
	hohoperatorinformer "github.com/stolostron/hub-of-hubs-operator/apis/client/informers/externalversions"
	constants "github.com/stolostron/hub-of-hubs-operator/pkg/constants"
	hohoperatorpropagator "github.com/stolostron/hub-of-hubs-operator/pkg/manager/propagator"
)

var (
	genericScheme = runtime.NewScheme()
	genericCodecs = serializer.NewCodecFactory(genericScheme)
	genericCodec  = genericCodecs.UniversalDeserializer()
)

//go:embed manifests
//go:embed manifests/agent
var fs embed.FS

var agentRBACFiles = []string{
	// role with RBAC rules to access resources on hub
	"manifests/rbac/role.yaml",
	// rolebinding to bind the above role to a certain user group
	"manifests/rbac/rolebinding.yaml",
}

func NewControllerCommand() *cobra.Command {
	cmd := controllercmd.
		NewControllerCommandConfig("hub-of-hubs-operator-manager", version.Get(), runManager).
		NewCommand()
	cmd.Use = "manager"
	cmd.Short = "Start the hub-of-hubs-operator manager"

	return cmd
}

func runManager(ctx context.Context, controllerContext *controllercmd.ControllerContext) error {
	mgr, err := addonmanager.New(controllerContext.KubeConfig)
	if err != nil {
		return err
	}
	registrationOption := newRegistrationOption(
		controllerContext.KubeConfig,
		controllerContext.EventRecorder,
		utilrand.String(5))

	agentAddon, err := addonfactory.NewAgentAddonFactory(constants.HoHOperatorName, fs, "manifests/agent").
		WithGetValuesFuncs(getValues, addonfactory.GetValuesFromAddonAnnotation).
		WithAgentRegistrationOption(registrationOption).
		WithInstallStrategy(agent.InstallAllStrategy(constants.HoHAgentNamespace)).
		WithInstallStrategy(agent.InstallByLabelStrategy(constants.HoHAgentNamespace, metav1.LabelSelector{MatchLabels: map[string]string{"vendor": "OpenShift"}})).
		BuildTemplateAgentAddon()
	if err != nil {
		klog.Errorf("failed to build agent %v", err)
		return err
	}

	err = mgr.AddAgent(agentAddon)
	if err != nil {
		klog.Fatal(err)
	}

	// build kube client
	kubeClient, err := kubernetes.NewForConfig(controllerContext.KubeConfig)
	if err != nil {
		return err
	}
	// build kube informer factory
	kubeInformerFactory := informers.NewSharedInformerFactory(kubeClient, 10*time.Minute)

	// build managedcluster kubeclient
	clusterClient, err := clusterclientset.NewForConfig(controllerContext.KubeConfig)
	if err != nil {
		return err
	}

	// build hohoperator kubeclient
	hohOperatorClient, err := hohoperatorclientset.NewForConfig(controllerContext.KubeConfig)
	if err != nil {
		return err
	}

	// build hohoperator informer factory
	hohOperatorInformerFactory := hohoperatorinformer.NewSharedInformerFactory(hohOperatorClient, 10*time.Minute)

	// create an instance of hohOperatorPropagatorController
	hohOperatorPropagatorController := hohoperatorpropagator.NewHohOperatorPropagatorController(
		clusterClient,
		hohOperatorClient,
		hohOperatorInformerFactory.Hubofhubs().V1alpha1().Configs(),
		controllerContext.EventRecorder,
	)

	err = mgr.Start(ctx)
	if err != nil {
		klog.Fatal(err)
	}

	go kubeInformerFactory.Start(ctx.Done())
	go hohOperatorInformerFactory.Start(ctx.Done())
	go hohOperatorPropagatorController.Run(ctx, 1)
	<-ctx.Done()

	return nil
}

func newRegistrationOption(kubeConfig *rest.Config, recorder events.Recorder, agentName string) *agent.RegistrationOption {
	return &agent.RegistrationOption{
		CSRConfigurations: agent.KubeClientSignerConfigurations(constants.HoHOperatorName, agentName),
		CSRApproveCheck:   utils.DefaultCSRApprover(agentName),
		PermissionConfig: func(cluster *clusterv1.ManagedCluster, addon *addonapiv1alpha1.ManagedClusterAddOn) error {
			kubeclient, err := kubernetes.NewForConfig(kubeConfig)
			if err != nil {
				return err
			}

			for _, file := range agentRBACFiles {
				if err := applyManifestFromFile(file, cluster.Name, addon.Name, kubeclient, recorder); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func applyManifestFromFile(file, clusterName, addonName string, kubeclient *kubernetes.Clientset, recorder events.Recorder) error {
	groups := agent.DefaultGroups(clusterName, addonName)
	config := struct {
		ClusterName string
		Group       string
	}{
		ClusterName: clusterName,
		Group:       groups[0],
	}

	results := resourceapply.ApplyDirectly(context.Background(),
		resourceapply.NewKubeClientHolder(kubeclient),
		recorder,
		resourceapply.NewResourceCache(),
		func(name string) ([]byte, error) {
			template, err := fs.ReadFile(file)
			if err != nil {
				return nil, err
			}
			return assets.MustCreateAssetFromTemplate(name, template, config).Data, nil
		},
		file,
	)

	for _, result := range results {
		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}

func getValues(cluster *clusterv1.ManagedCluster, addon *addonapiv1alpha1.ManagedClusterAddOn) (addonfactory.Values, error) {
	installNamespace := addon.Spec.InstallNamespace
	if len(installNamespace) == 0 {
		installNamespace = constants.HoHAgentNamespace
	}

	image := os.Getenv("HUB_OF_HUBS_OPERATOR_IMAGE")
	if len(image) == 0 {
		image = constants.DefaultHoHOperatorImage
	}

	manifestConfig := struct {
		KubeConfigSecret      string
		ClusterName           string
		AddonInstallNamespace string
		Image                 string
	}{
		KubeConfigSecret:      fmt.Sprintf("%s-hub-kubeconfig", addon.Name),
		AddonInstallNamespace: installNamespace,
		ClusterName:           cluster.Name,
		Image:                 image,
	}

	return addonfactory.StructToValues(manifestConfig), nil
}
