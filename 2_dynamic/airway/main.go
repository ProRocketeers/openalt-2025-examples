package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	v1 "yoke-playground/backend/api/v1"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/yokecd/yoke/pkg/apis/airway/v1alpha1"
	"github.com/yokecd/yoke/pkg/openapi"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	return json.NewEncoder(os.Stdout).Encode(v1alpha1.Airway{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.AirwayGVR().GroupVersion().Identifier(),
			Kind:       v1alpha1.KindAirway,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "backendsdynamic.yoke.examples",
		},
		Spec: v1alpha1.AirwaySpec{
			Mode: v1alpha1.AirwayModeDynamic,
			WasmURLs: v1alpha1.WasmURLs{
				Flight: "oci://ghcr.io/prorocketeers/openalt-2025-examples/backend_dynamic:latest",
			},
			// this property allows the Flight to access the cluster
			ClusterAccess: true,
			// and this property specifies, what exactly can it access
			// in this case, it's any `ConfigMap` in any namespace (which are in `core` API group, and therefore omitted)
			ResourceAccessMatchers: []string{"ConfigMap"},
			Template: apiextv1.CustomResourceDefinitionSpec{
				Group: "yoke.examples",
				Names: apiextv1.CustomResourceDefinitionNames{
					Plural:     "backendsdynamic",
					Singular:   "backenddynamic",
					ShortNames: []string{"bed"},
					Kind:       "BackendDynamic",
				},
				Scope: apiextv1.NamespaceScoped,
				Versions: []apiextv1.CustomResourceDefinitionVersion{
					{
						Name:    "v1",
						Served:  true,
						Storage: true,
						Schema: &apiextv1.CustomResourceValidation{
							OpenAPIV3Schema: openapi.SchemaFrom(reflect.TypeFor[v1.BackendDynamic]()),
						},
						// since we're working with a status resource on the CR, we need to specify that in the CRD, so it exposes Kube API for interacting with it
						Subresources: &apiextv1.CustomResourceSubresources{
							Status: &apiextv1.CustomResourceSubresourceStatus{},
						},
					},
				},
			},
		},
	})
}
