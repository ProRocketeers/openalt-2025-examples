package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	APIVersion  = "yoke.examples/v1"
	KindBackend = "Backend"
)

type Backend struct {
	metav1.TypeMeta
	metav1.ObjectMeta `json:"metadata"`
	Spec              BackendSpec `json:"spec"`
}

type BackendSpec struct {
	Image    string `json:"image"`
	Replicas int32  `json:"replicas"`
	Port     int32  `json:"port"`
}
