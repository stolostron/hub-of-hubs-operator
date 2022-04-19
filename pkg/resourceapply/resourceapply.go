package resourceapply

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	"github.com/openshift/library-go/pkg/operator/events"
	openshiftresourceapply "github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/library-go/pkg/operator/resource/resourcehelper"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"

	hubofhubsv1alpha1client "github.com/stolostron/hub-of-hubs-operator/apis/client/clientset/versioned/typed/hubofhubs/v1alpha1"
	hubofhubsv1alpha1 "github.com/stolostron/hub-of-hubs-operator/apis/hubofhubs/v1alpha1"
)

func ApplyAgentConfig(ctx context.Context, client hubofhubsv1alpha1client.AgentConfigsGetter, recorder events.Recorder, required *hubofhubsv1alpha1.AgentConfig) (*hubofhubsv1alpha1.AgentConfig, bool, error) {
	existing, err := client.AgentConfigs(required.Namespace).Get(context.TODO(), required.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		requiredCopy := required.DeepCopy()
		actual, err := client.AgentConfigs(requiredCopy.Namespace).
			Create(context.TODO(), resourcemerge.WithCleanLabelsAndAnnotations(requiredCopy).(*hubofhubsv1alpha1.AgentConfig), metav1.CreateOptions{})
		reportCreateEvent(recorder, requiredCopy, err)
		return actual, true, err
	}
	if err != nil {
		return nil, false, err
	}

	modified := resourcemerge.BoolPtr(false)
	existingCopy := existing.DeepCopy()

	resourcemerge.EnsureObjectMeta(modified, &existingCopy.ObjectMeta, required.ObjectMeta)
	specSame := reflect.DeepEqual(&existingCopy.Spec, required.Spec)
	if !*modified && specSame {
		return existingCopy, false, nil
	}
	if !specSame {
		existingCopy.Spec = required.Spec
	}

	klog.V(2).Infof("AgentConfig %q changes: %v", required.Namespace+"/"+required.Name, openshiftresourceapply.JSONPatchNoError(existing, required))
	actual, err := client.AgentConfigs(required.Namespace).Update(context.TODO(), existingCopy, metav1.UpdateOptions{})
	reportUpdateEvent(recorder, required, err)
	return actual, true, err
}

func reportCreateEvent(recorder events.Recorder, obj runtime.Object, originalErr error) {
	gvk := resourcehelper.GuessObjectGroupVersionKind(obj)
	if originalErr == nil {
		recorder.Eventf(fmt.Sprintf("%sCreated", gvk.Kind), "Created %s because it was missing", resourcehelper.FormatResourceForCLIWithNamespace(obj))
		return
	}
	recorder.Warningf(fmt.Sprintf("%sCreateFailed", gvk.Kind), "Failed to create %s: %v", resourcehelper.FormatResourceForCLIWithNamespace(obj), originalErr)
}

func reportUpdateEvent(recorder events.Recorder, obj runtime.Object, originalErr error, details ...string) {
	gvk := resourcehelper.GuessObjectGroupVersionKind(obj)
	switch {
	case originalErr != nil:
		recorder.Warningf(fmt.Sprintf("%sUpdateFailed", gvk.Kind), "Failed to update %s: %v", resourcehelper.FormatResourceForCLIWithNamespace(obj), originalErr)
	case len(details) == 0:
		recorder.Eventf(fmt.Sprintf("%sUpdated", gvk.Kind), "Updated %s because it changed", resourcehelper.FormatResourceForCLIWithNamespace(obj))
	default:
		recorder.Eventf(fmt.Sprintf("%sUpdated", gvk.Kind), "Updated %s:\n%s", resourcehelper.FormatResourceForCLIWithNamespace(obj), strings.Join(details, "\n"))
	}
}
