package deployer

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// Deployer is the interface for the kubernetes resource deployer
type Deployer interface {
	Deploy(obj runtime.Object) error
}
