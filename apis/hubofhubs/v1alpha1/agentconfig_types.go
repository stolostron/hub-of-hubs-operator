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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AgentConfigSpec defines the desired state of AgentConfig
type AgentConfigSpec struct {
	Global     *AgentGlobalConfig     `json:"global,omitempty"`
	Components *AgentComponentsConfig `json:"components,omitempty"`
}

// AgentGlobalConfig defines common settings for leaf hub
type AgentGlobalConfig struct {
	HeartbeatInterval *AgentHeartbeatIntervalConfig `json:"heartbeatInterval,omitempty"`
}

// AgentHeartbeatIntervalConfig defines heartbeat interval in seconds leaf hub
type AgentHeartbeatIntervalConfig struct {
	LeafHub uint64 `default:"60" json:"leafHub,omitempty"`
}

// AgentComponentsConfig defines settings for leaf hub components
type AgentComponentsConfig struct {
	Core      *AgentCoreConfig      `json:"core,omitempty"`
	Transport *AgentTransportConfig `json:"transport,omitempty"`
}

// AgentCoreConfig defines settings for leaf hub core controllers
type AgentCoreConfig struct {
	LeafHub *LeafHubConfig `json:"leafHub,omitempty"`
}

// AgentTransportConfig defines settings for transport layer
type AgentTransportConfig struct {
	Provider TransportProvider `json:"provider,omitempty"` // kafka or sync-service
	// Kafka       *KafkaConfig       `json:"kafka,omitempty"`
	SyncService *SyncServiceConfig `json:"syncService,omitempty"`
}

// AgentConfigStatus defines the observed state of AgentConfig
type AgentConfigStatus struct {
}

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AgentConfig is the Schema for the agentConfigs API
type AgentConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentConfigSpec   `json:"spec,omitempty"`
	Status AgentConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AgentConfigList contains a list of AgentConfig
type AgentConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgentConfig{}, &AgentConfigList{})
}
