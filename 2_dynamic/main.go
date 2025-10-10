package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	v1 "yoke-playground/backend/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/yokecd/yoke/pkg/flight"
	"github.com/yokecd/yoke/pkg/flight/wasi/k8s"
)

// Flight entrypoint
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// this example is a fabricated, not really realistic example but one nonetheless
	// BackendDynamic is going to reference a ConfigMap, which can deploy multiple applications in a certain amount of replicas
	// just to show how Yoke can dynamically react to live changes in cluster
	var backend v1.BackendDynamic
	if err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&backend); err != nil && err != io.EOF {
		return err
	}
	encoder := json.NewEncoder(os.Stdout)

	// when working with Dynamic Flights, you have to return the CR itself as well
	// `flight.Resources` is just a convenience type over an array of any Kubernetes resource
	resources := flight.Resources{&backend}

	// first try to fetch the configmap
	// Yoke automatically tracks changes to any resource you look up, so it can react to what's happening
	cm, err := k8s.Lookup[corev1.ConfigMap](k8s.ResourceIdentifier{
		Name:       backend.Spec.ConfigMapName,
		Namespace:  backend.Namespace,
		ApiVersion: "v1",
		Kind:       "ConfigMap",
	})
	if err != nil {
		if k8s.IsErrNotFound(err) {
			// when complicated logic is used in reconciling the CR, it's useful to expose at least *some* information to the user as to what's happening
			backend.Status.Message = "ConfigMap not found"
			return encoder.Encode(resources)
		}
		return fmt.Errorf("Failed to fetch the ConfigMap: %w", err)
	}
	configStr, ok := cm.Data["config"]
	if !ok || configStr == "" {
		backend.Status.Message = "ConfigMap does not exist, or the 'config' field does not exist"
		return encoder.Encode(resources)
	}

	type configItem struct {
		Replicas int32 `json:"replicas"`
	}

	var config map[string]configItem

	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		backend.Status.Message = fmt.Sprintf("Config `%v` is not a valid JSON", configStr)
		return encoder.Encode(resources)
	}

	// create a Deployment + Service pair for every config item in the ConfigMap
	for name, value := range config {
		resources = append(
			resources,
			createDeployment(fmt.Sprintf("%v-%v", backend.Name, name), backend.Namespace, value.Replicas),
			createService(fmt.Sprintf("%v-%v", backend.Name, name), backend.Namespace),
		)
	}
	backend.Status.Message = fmt.Sprintf("Deployed %v deployments", len(config))
	return encoder.Encode(resources)
}

func commonLabels(name string) map[string]string {
	return map[string]string{
		"app": name,
	}
}

func createDeployment(name, namespace string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.Identifier(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: commonLabels(name),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: commonLabels(name),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{Name: "http", Protocol: corev1.ProtocolTCP, ContainerPort: int32(80)},
							},
						},
					},
				},
			},
		},
	}
}

func createService(name, namespace string) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.Identifier(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: commonLabels(name),
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromString("http"),
				},
			},
		},
	}
}
