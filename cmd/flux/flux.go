package flux

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/source-controller/api/v1beta2"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KustomizationStatus struct {
	Name                string `json:"name" yaml:"name"`
	Namespace           string `json:"namespace" yaml:"namespace"`
	HelmRelease         string `json:"helmRelease,omitempty" yaml:"helmRelease,omitempty"`
	ReadyStatus         string `json:"readyStatus" yaml:"readyStatus"`
	LastAppliedRevision string `json:"lastAppliedRevision" yaml:"lastAppliedRevision"`
	Message             string `json:"message,omitempty" yaml:"message,omitempty"`
	Suspended           bool   `json:"suspended" yaml:"suspended"`
}

func ListKustomizationsSTRUCT(config *rest.Config, namespace, outputFormat string) {
	_ = v1.AddToScheme(scheme.Scheme)
	_ = v1beta2.AddToScheme(scheme.Scheme)

	k8sClient, err := client.New(config, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	var kustomizations v1.KustomizationList

	opts := []client.ListOption{}
	if namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}

	if err := k8sClient.List(ctx, &kustomizations, opts...); err != nil {
		panic(err)
	}

	var results []KustomizationStatus

	for _, k := range kustomizations.Items {
		readyStatus := "Unknown"
		message := ""

		for _, cond := range k.Status.Conditions {
			if cond.Type == "Ready" {
				readyStatus = string(cond.Status)
				message = cond.Message
				break
			}
		}

		helmRelease := k.Annotations["meta.helm.sh/release-name"]

		results = append(results, KustomizationStatus{
			Name:                k.Name,
			Namespace:           k.Namespace,
			HelmRelease:         helmRelease,
			ReadyStatus:         readyStatus,
			LastAppliedRevision: k.Status.LastAppliedRevision,
			Message:             message,
			Suspended:           k.Spec.Suspend,
		})
	}

	switch outputFormat {
	case "json":
		out, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(out))
	case "yaml":
		out, _ := yaml.Marshal(results)
		fmt.Println(string(out))
	default:
		fmt.Println("Unsupported format. Use 'json' or 'yaml'.")
	}
}
