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

// AggregationLevel specifies the level of aggregation leaf hubs should do before sending the information
// +kubebuilder:validation:Enum=full;minimal
type AggregationLevel string

const (
	// Full is an AggregationLevel
	Full AggregationLevel = "full"

	// Minimal is an AggregationLevel
	Minimal AggregationLevel = "minimal"
)

// MsgCompressType specifies the type of message compress should do before sending the information to transport
// +kubebuilder:validation:Enum=gzip;no-op
type MsgCompressType string

const (
	// GzipMsgCompressType is a MsgCompressType
	GzipMsgCompressType MsgCompressType = "gzip"

	// NoopMsgCompressType is a MsgCompressType
	NoopMsgCompressType MsgCompressType = "no-op"
)

// TransportProvider specifies the provider type of transport layer
// +kubebuilder:validation:Enum=kafka;sync-service
type TransportProvider string

const (
	// KafkaTransportProvider is a TransportProvider
	KafkaTransportProvider TransportProvider = "kafka"

	// KafkaTransportProvider is a TransportProvider
	SyncServiceTransportProvider TransportProvider = "sync-service"
)

// DatabaseProvider specifies the provider type of database
// +kubebuilder:validation:Enum=postgresql
type DatabaseProvider string

const (
	// PostgreSqlDatabaseProvider is a DatabaseProvider
	PostgreSqlDatabaseProvider DatabaseProvider = "postgresql"
)

// ConfigSpec defines the desired state of Config
type ConfigSpec struct {
	Global     *GlobalConfig     `json:"global,omitempty"`
	Components *ComponentsConfig `json:"components,omitempty"`
}

// GlobalConfig defines common settings
type GlobalConfig struct {
	AggregationLevel    AggregationLevel         `json:"aggregationLevel,omitempty"` // full or minimal
	HeartbeatInterval   *HeartbeatIntervalConfig `json:"heartbeatInterval,omitempty"`
	EnableLocalPolicies bool                     `json:"enableLocalPolicies,omitempty"`
}

// HeartbeatIntervalConfig defines heartbeat intervals for HoH and Leaf hub in seconds
type HeartbeatIntervalConfig struct {
	HoHInSeconds     uint64 `default:"60" json:"hohInSeconds,omitempty"`
	LeafHubInSeconds uint64 `default:"60" json:"leafHubInSeconds,omitempty"`
}

// ComponentsConfig defines settings for all components
type ComponentsConfig struct {
	Core      *CoreConfig      `json:"core,omitempty"`
	Transport *TransportConfig `json:"transport,omitempty"`
	Database  *DatabaseConfig  `json:"database,omitempty"`
}

// CoreConfig defines settings for hub-of-hubs core controllers
type CoreConfig struct {
	Hoh     *HohConfig     `json:"hoh,omitempty"`
	LeafHub *LeafHubConfig `json:"leafHub,omitempty"`
}

// HohConfig defines settings for core controllers in hub of hubs cluster
type HohConfig struct {
	Nonk8sAPI             *Nonk8sAPIConfig             `json:"nonk8sAPI,omitempty"`
	RBAC                  *RBACConfig                  `json:"rbac,omitempty"`
	SpecSync              *SpecSyncConfig              `json:"specSync,omitempty"`
	StatusSync            *StatusSyncConfig            `json:"statusSync,omitempty"`
	SpecTransportBridge   *SpecTransportBridgeConfig   `json:"specTransportBridge,omitempty"`
	StatusTransportBridge *StatusTransportBridgeConfig `json:"statusTransportBridge,omitempty"`
}

// Nonk8sAPIConfig defines settings for nonk8s-API
type Nonk8sAPIConfig struct {
	BasePath string `json:"basePath,omitempty"`
}

// RBACConfig defines settings for RBAC
type RBACConfig struct {
}

// SpecSyncConfig defines settings for spec-sync
type SpecSyncConfig struct {
}

// StatusSyncConfig defines settings for status-sync
type StatusSyncConfig struct {
	SyncInterval uint64 `default:"5" json:"syncInterval,omitempty"`
}

// SpecTransportBridgeConfig defines settings for spec-transport-bridge
type SpecTransportBridgeConfig struct {
	SyncInterval    uint64          `default:"5" json:"syncInterval,omitempty"`
	MsgCompressType MsgCompressType `json:"msgCompressType,omitempty"` // gzip or no-op
	MsgSizeLimit    uint64          `default:"940" json:"msgSizeLimit,omitempty"`
}

// StatusTransportBridgeConfig defines settings for status-transport-bridge
type StatusTransportBridgeConfig struct {
	CommitterInterval     uint64 `default:"5" json:"committerInterval,omitempty"`
	StatisticsLogInterval uint64 `default:"5" json:"statisticsLogInterval,omitempty"`
}

// LeafHubConfig defines settings for core controllers in leaf hub cluster
type LeafHubConfig struct {
	SpecSync   *LeafHubSpecSyncConfig   `json:"specSync,omitempty"`
	StatusSync *LeafHubStatusSyncConfig `json:"statusSync,omitempty"`
}

// LeafHubSpecSyncConfig defines settings for leafhub-spec-sync
type LeafHubSpecSyncConfig struct {
	KubeClientPoolSIze uint64 `default:"10" json:"kubeClientPoolSIze,omitempty"`
	EnforceHoHRbac     bool   `json:"enforceHoHRbac,omitempty"`
}

// LeafHubStatusSyncConfig defines settings for leafhub-status-sync
type LeafHubStatusSyncConfig struct {
	SyncInterval               *LeafHubStatusSyncIntervalSettings `json:"syncIntervalConfig,omitempty"`
	DeltaSentCountSwitchFactor uint64                             `default:"100" json:"deltaSentCountSwitchFactor,omitempty"`
	MsgCompressType            MsgCompressType                    `json:"msgCompressType,omitempty"` // gzip or no-op
	MsgSizeLimit               uint64                             `default:"940" json:"msgSizeLimit,omitempty"`
}

// LeafHubStatusSyncIntervalSettings defines snyc interval settings for leahub-status-sync
type LeafHubStatusSyncIntervalSettings struct {
	ManagedClusterSyncInterval uint64 `default:"5" json:"managedClusters,omitempty"`
	PolicySyncInterval         uint64 `default:"5" json:"policies,omitempty"`
	ControlInfoSyncInterval    uint64 `default:"3600" json:"controlInfo,omitempty"`
}

// TransportConfig defines settings for transport layer
type TransportConfig struct {
	Provider    TransportProvider  `json:"provider,omitempty"` // kafka or sync-service
	Kafka       *KafkaConfig       `json:"kafka,omitempty"`
	SyncService *SyncServiceConfig `json:"syncService,omitempty"`
}

// KafkaConfig defines settings for Kafka transport
type KafkaConfig struct {
	Version  string `json:"version,omitempty"`
	Replicas uint64 `default:"3" json:"replicas,omitempty"`
}

// SyncServiceConfig defines settings for Sync-service transport
type SyncServiceConfig struct {
	Version         string `json:"version,omitempty"`
	PollingInterval uint64 `default:"5" json:"pollingInterval,omitempty"`
}

// DatabaseConfig defines settings for database
type DatabaseConfig struct {
	Provider   DatabaseProvider  `json:"provider,omitempty"` // postgresql
	Postgresql *PostgreSqlConfig `json:"postgresql,omitempty"`
}

// PostgreSqlConfig defines settings for PostgreSql
type PostgreSqlConfig struct {
	Version  string `json:"version,omitempty"`
	EnableHA bool   `json:"enableHA,omitempty"`
}

// ConfigStatus defines the observed state of Config
type ConfigStatus struct {
}

// +genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Config is the Schema for the configs API
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigSpec   `json:"spec,omitempty"`
	Status ConfigStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConfigList contains a list of Config
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Config{}, &ConfigList{})
}
