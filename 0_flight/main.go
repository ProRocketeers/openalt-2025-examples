package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type values struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  int32  `json:"replicas"`
	Image     string `json:"image"`
	Port      int32  `json:"port"`
}

// Flight entrypoint
func main() {
	// obal, který jakýkoliv error pouze zapíše na `stderr`, tím se vrátí ven jako chybová hláška až do ArgoCD
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var values values
	// z `stdin` přečteme JSON reprezentaci našich values
	err := yaml.NewYAMLToJSONDecoder(os.Stdin).Decode(&values)

	if err != nil && err != io.EOF {
		return err
	}

	// vytvoříme dané resourcy a zapíšeme je jako pole do JSONu na `stdout`
	return json.NewEncoder(os.Stdout).Encode([]any{
		createDeployment(values),
		createService(values),
	})
}

func createDeployment(values values) appsv1.Deployment {
	return appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.Identifier(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      values.Name,
			Namespace: values.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &values.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: commonLabels(values),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: commonLabels(values),
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  values.Name,
							Image: values.Image,
							Ports: []corev1.ContainerPort{
								{Name: "http", Protocol: corev1.ProtocolTCP, ContainerPort: values.Port},
							},
						},
					},
				},
			},
		},
	}
}
func commonLabels(values values) map[string]string {
	return map[string]string{
		"app": values.Name,
	}
}
func createService(values values) corev1.Service {
	return corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.Identifier(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      values.Name,
			Namespace: values.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: commonLabels(values),
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
