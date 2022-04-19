module github.com/stolostron/hub-of-hubs-operator

go 1.16

require (
	github.com/openshift/library-go v0.0.0-20220414091252-b28506bc43e5
	github.com/spf13/cobra v1.2.1
	k8s.io/apimachinery v0.23.4
	k8s.io/client-go v0.23.4
	k8s.io/code-generator v0.23.4
	k8s.io/component-base v0.23.0
	k8s.io/klog/v2 v2.30.0
	open-cluster-management.io/addon-framework v0.3.0
	open-cluster-management.io/api v0.7.0
	sigs.k8s.io/controller-runtime v0.11.1
	sigs.k8s.io/yaml v1.3.0
)
