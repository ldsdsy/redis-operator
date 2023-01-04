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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RedisStandaloneSpec defines the desired state of RedisStandalone
type RedisStandaloneSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name            string                      `json:"name"`
	Image           string                      `json:"image"`
	ImagePullPolicy corev1.PullPolicy           `json:"imagePullPolicy,omitempty"`
	Resources       corev1.ResourceRequirements `json:"resources,omitempty"`
	Storage         RedisStorage                `json:"storage,omitempty"`
	Configuration   map[string]string           `json:"configuration,omitempty"`
}

type RedisStorage struct {
	StorageClass string            `json:"storageClass"`
	Size         resource.Quantity `json:"size"`
	Retain       bool              `json:"retain"`
}

// RedisStandaloneStatus defines the observed state of RedisStandalone
type RedisStandaloneStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status StandaloneStatus `json:"standaloneStatus"`
	Reason string           `json:"reason"`
}
type StandaloneStatus string

const (
	StatusOK StandaloneStatus = "Healthy"
	StatusKO StandaloneStatus = "Failed"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// RedisStandalone is the Schema for the redisstandalones API
type RedisStandalone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisStandaloneSpec   `json:"spec,omitempty"`
	Status RedisStandaloneStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RedisStandaloneList contains a list of RedisStandalone
type RedisStandaloneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisStandalone `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisStandalone{}, &RedisStandaloneList{})
}
