/*
Copyright 2022.

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

package hubofhubs

import (
	"context"
	"embed"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	hubofhubsv1alpha1 "github.com/stolostron/hub-of-hubs-operator/apis/hubofhubs/v1alpha1"
	"github.com/stolostron/hub-of-hubs-operator/pkg/deployer"
	"github.com/stolostron/hub-of-hubs-operator/pkg/renderer"
)

//go:embed manifests
//go:embed manifests/agent
//go:embed manifests/database
//go:embed manifests/manager
var fs embed.FS

type managerConfig struct {
	Registry      string
	ImageTag      string
	TransportType string
}

// ConfigReconciler reconciles a Config object
type ConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=hubofhubs.open-cluster-management.io,resources=configs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hubofhubs.open-cluster-management.io,resources=configs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hubofhubs.open-cluster-management.io,resources=configs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Config object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// Fetch the hub-of-hubs config instance
	hohConfig := &hubofhubsv1alpha1.Config{}
	err := r.Get(ctx, req.NamespacedName, hohConfig)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Config resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Config")
		return ctrl.Result{}, err
	}

	transportType := "kafka"
	if hohConfig.Spec.Components != nil && hohConfig.Spec.Components.Transport != nil {
		if hohConfig.Spec.Components.Transport.Provider == hubofhubsv1alpha1.SyncServiceTransportProvider {
			transportType = "sync-service"
		}
	}

	// create new HoHRenderer and HoHDeployer
	hohRenderer := renderer.NewHoHRenderer(fs)
	hohDeployer := deployer.NewHoHDeployer(r.Client)
	dbObjects, err := hohRenderer.Render("manifests/database", func(component string) (interface{}, error) {
		dbConfig := struct {
			Registry string
			ImageTag string
		}{
			Registry: "quay.io/open-cluster-management-hub-of-hubs",
			ImageTag: "latest",
		}

		return dbConfig, err
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	var dbInitJobObj runtime.Object
	for _, obj := range dbObjects {
		if obj.GetObjectKind().GroupVersionKind().Kind == "Job" {
			dbInitJobObj = obj
			continue
		}

		log.Info("Creating or updating object", "object", obj)
		err := hohDeployer.Deploy(obj)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// create or updating the database initialization job
	log.Info("Creating or updating object", "object", dbInitJobObj)
	if err = hohDeployer.Deploy(dbInitJobObj); err != nil {
		return ctrl.Result{}, err
	}

	managerObjects, err := hohRenderer.Render("manifests/manager", func(component string) (interface{}, error) {
		managerConfig := struct {
			Registry      string
			ImageTag      string
			TransportType string
		}{
			Registry:      "quay.io/open-cluster-management-hub-of-hubs",
			ImageTag:      "latest",
			TransportType: transportType,
		}

		return managerConfig, err
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, obj := range managerObjects {
		log.Info("Creating or updating object", "object", obj)
		err := hohDeployer.Deploy(obj)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hubofhubsv1alpha1.Config{}).
		Complete(r)
}
