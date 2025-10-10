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
		ObjectMeta: metav1.ObjectMeta{
			Name: "backends.yoke.examples",
		},
		Spec: v1alpha1.AirwaySpec{
			Mode: v1alpha1.AirwayModeStandard,
			WasmURLs: v1alpha1.WasmURLs{
				Flight: "oci://ghcr.io/prorocketeers/openalt-2025-examples/backend:latest",
			},
			Template: apiextv1.CustomResourceDefinitionSpec{
				Group: "yoke.examples",
				Names: apiextv1.CustomResourceDefinitionNames{
					Plural:     "backends",
					Singular:   "backend",
					ShortNames: []string{"be"},
					Kind:       "Backend",
				},
				Scope: apiextv1.NamespaceScoped,
				Versions: []apiextv1.CustomResourceDefinitionVersion{
					{
						Name:    "v1",
						Served:  true,
						Storage: true,
						Schema: &apiextv1.CustomResourceValidation{
							OpenAPIV3Schema: openapi.SchemaFrom(reflect.TypeFor[v1.Backend]()),
						},
					},
				},
			},
		},
	})
}
