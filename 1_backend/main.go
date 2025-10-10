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
)

// Flight entrypoint
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var backend v1.Backend
	if err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&backend); err != nil && err != io.EOF {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode([]any{
		createDeployment(backend),
		createService(backend),
	})
}

func createDeployment(b v1.Backend) appsv1.Deployment {
	return appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.Identifier(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name,
			Namespace: b.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &b.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: commonLabels(b),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: commonLabels(b),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  b.Name,
							Image: b.Spec.Image,
							Ports: []corev1.ContainerPort{
								{Name: "http", Protocol: corev1.ProtocolTCP, ContainerPort: b.Spec.Port},
							},
						},
					},
				},
			},
		},
	}
}

func commonLabels(b v1.Backend) map[string]string {
	return map[string]string{
		"app": b.Name,
	}
}

func createService(b v1.Backend) corev1.Service {
	return corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.Identifier(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.Name,
			Namespace: b.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: commonLabels(b),
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
