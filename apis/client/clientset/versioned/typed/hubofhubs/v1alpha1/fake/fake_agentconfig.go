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
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/stolostron/hub-of-hubs-operator/apis/hubofhubs/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeAgentConfigs implements AgentConfigInterface
type FakeAgentConfigs struct {
	Fake *FakeHubofhubsV1alpha1
	ns   string
}

var agentconfigsResource = schema.GroupVersionResource{Group: "hubofhubs.open-cluster-management.io", Version: "v1alpha1", Resource: "agentconfigs"}

var agentconfigsKind = schema.GroupVersionKind{Group: "hubofhubs.open-cluster-management.io", Version: "v1alpha1", Kind: "AgentConfig"}

// Get takes name of the agentConfig, and returns the corresponding agentConfig object, and an error if there is any.
func (c *FakeAgentConfigs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.AgentConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(agentconfigsResource, c.ns, name), &v1alpha1.AgentConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AgentConfig), err
}

// List takes label and field selectors, and returns the list of AgentConfigs that match those selectors.
func (c *FakeAgentConfigs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.AgentConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(agentconfigsResource, agentconfigsKind, c.ns, opts), &v1alpha1.AgentConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.AgentConfigList{ListMeta: obj.(*v1alpha1.AgentConfigList).ListMeta}
	for _, item := range obj.(*v1alpha1.AgentConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested agentConfigs.
func (c *FakeAgentConfigs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(agentconfigsResource, c.ns, opts))

}

// Create takes the representation of a agentConfig and creates it.  Returns the server's representation of the agentConfig, and an error, if there is any.
func (c *FakeAgentConfigs) Create(ctx context.Context, agentConfig *v1alpha1.AgentConfig, opts v1.CreateOptions) (result *v1alpha1.AgentConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(agentconfigsResource, c.ns, agentConfig), &v1alpha1.AgentConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AgentConfig), err
}

// Update takes the representation of a agentConfig and updates it. Returns the server's representation of the agentConfig, and an error, if there is any.
func (c *FakeAgentConfigs) Update(ctx context.Context, agentConfig *v1alpha1.AgentConfig, opts v1.UpdateOptions) (result *v1alpha1.AgentConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(agentconfigsResource, c.ns, agentConfig), &v1alpha1.AgentConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AgentConfig), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeAgentConfigs) UpdateStatus(ctx context.Context, agentConfig *v1alpha1.AgentConfig, opts v1.UpdateOptions) (*v1alpha1.AgentConfig, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(agentconfigsResource, "status", c.ns, agentConfig), &v1alpha1.AgentConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AgentConfig), err
}

// Delete takes name of the agentConfig and deletes it. Returns an error if one occurs.
func (c *FakeAgentConfigs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(agentconfigsResource, c.ns, name, opts), &v1alpha1.AgentConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeAgentConfigs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(agentconfigsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.AgentConfigList{})
	return err
}

// Patch applies the patch and returns the patched agentConfig.
func (c *FakeAgentConfigs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.AgentConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(agentconfigsResource, c.ns, name, pt, data, subresources...), &v1alpha1.AgentConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.AgentConfig), err
}
