package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	APIVersion  = "yoke.examples/v1"
	KindBackend = "BackendDynamic"
)

type BackendDynamic struct {
	metav1.TypeMeta
	metav1.ObjectMeta `json:"metadata"`
	Spec              BackendDynamicSpec   `json:"spec"`
	Status            BackendDynamicStatus `json:"status"`
}

type BackendDynamicSpec struct {
	ConfigMapName string `json:"configMapName"`
	// will have one field "config", which will hold a json string
	// JSON deserializes into object {[name]: { replicas: number }}
}

type BackendDynamicStatus struct {
	Message string `json:"message"`
}
