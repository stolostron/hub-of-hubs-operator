package renderer

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// GetConfigValuesFunc is function type that returns the configuration values for the given component
type GetConfigValuesFunc func(component string) (interface{}, error)

// GetClusterConfigValuesFunc is function type that returns the configuration values for the given component of given cluster
type GetClusterConfigValuesFunc func(cluster, component string) (interface{}, error)

// Renderer is the interface for the template renderer
type Renderer interface {
	Render(component string, getConfigValuesFunc GetConfigValuesFunc) ([]runtime.Object, error)
	RenderWithFilter(component, filter string, getConfigValuesFunc GetConfigValuesFunc) ([]runtime.Object, error)
	RenderForCluster(cluster, component string, getClusterConfigValuesFunc GetClusterConfigValuesFunc) ([]runtime.Object, error)
	RenderForClusterWithFilter(cluster, component, filter string, getClusterConfigValuesFunc GetClusterConfigValuesFunc) ([]runtime.Object, error)
}
