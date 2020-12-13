/*


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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VirtualRobotSpec defines the desired state of VirtualRobot
type VirtualRobotSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of VirtualRobot. Edit VirtualRobot_types.go to remove/update
	RobotName string `json:"robotName"`
	BaseURL   string `json:"baseUrl"`
}

// VirtualRobotStatus defines the observed state of VirtualRobot
type VirtualRobotStatus struct {
	URL string `json:"url"`
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Robot",type=string,JSONPath=`.spec.robotName`
// +kubebuilder:printcolumn:name="URL",type=integer,JSONPath=`.status.url`

// VirtualRobot is the Schema for the virtualrobots API
type VirtualRobot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualRobotSpec   `json:"spec,omitempty"`
	Status VirtualRobotStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
//

// VirtualRobotList contains a list of VirtualRobot
type VirtualRobotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualRobot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualRobot{}, &VirtualRobotList{})
}
