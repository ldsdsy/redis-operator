/*
Copyright 2022 ldsdsy.

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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisSentinelSpec defines the desired state of RedisSentinel
type RedisSentinelSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name     string       `json:"name"`
	Redis    RedisSpec    `json:"redis,omitempty"`
	Sentinel SentinelSpec `json:"sentinel,omitempty"`
}
type RedisSpec struct {
	Replicas        int32                       `json:"replicas"`
	Image           string                      `json:"image"`
	ImagePullPolicy corev1.PullPolicy           `json:"imagePullPolicy,omitempty"`
	Resources       corev1.ResourceRequirements `json:"resources,omitempty"`
	Storage         Storage                     `json:"storage,omitempty"`
	Configuration   map[string]string           `json:"configuration,omitempty"`
}
type SentinelSpec struct {
	Replicas        int32                       `json:"replicas"`
	Image           string                      `json:"image"`
	ImagePullPolicy corev1.PullPolicy           `json:"imagePullPolicy,omitempty"`
	Resources       corev1.ResourceRequirements `json:"resources,omitempty"`
	Storage         Storage                     `json:"storage,omitempty"`
	Configuration   map[string]string           `json:"configuration,omitempty"`
}

// RedisSentinelStatus defines the observed state of RedisSentinel
type RedisSentinelStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status Status `json:"standaloneStatus"`
	Reason string `json:"reason"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RedisSentinel is the Schema for the redissentinels API
type RedisSentinel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisSentinelSpec   `json:"spec,omitempty"`
	Status RedisSentinelStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedisSentinelList contains a list of RedisSentinel
type RedisSentinelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisSentinel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisSentinel{}, &RedisSentinelList{})
}
