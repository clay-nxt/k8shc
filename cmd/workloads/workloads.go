package workloads

import (
	"context"
	"encoding/json"
	"fmt"

	// appsv1 "k8s.io/api/apps/v1"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Workload struct {
	Kind        string `json:"kind" yaml:"kind"`
	Name        string `json:"name" yaml:"name"`
	Namespace   string `json:"namespace" yaml:"namespace"`
	HelmRelease string `json:"helmRelease,omitempty" yaml:"helmRelease,omitempty"`
	Service     string `json:"service,omitempty" yaml:"service,omitempty"`
	Ready       int32  `json:"ready" yaml:"ready"`
	Desired     int32  `json:"desired" yaml:"desired"`
	Status      string `json:"status" yaml:"status"`
}

func ListSTRUCT(clientset *kubernetes.Clientset, namespace string, includeDeployments, includeStatefulSets, includeDaemonSets bool, outputFormat string) {
	ctx := context.Background()
	services := getServiceMap(clientset, namespace)
	var results []Workload

	if includeDeployments {
		deployments, _ := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		for _, d := range deployments.Items {
			ready := d.Status.ReadyReplicas
			desired := *d.Spec.Replicas
			status := healthStatus(int(ready), int(desired))
			helmRelease := d.Annotations["meta.helm.sh/release-name"]
			svc := findService(services, d.Labels)

			results = append(results, Workload{
				Kind:        "Deployment",
				Name:        d.Name,
				Namespace:   d.Namespace,
				HelmRelease: helmRelease,
				Service:     svc,
				Ready:       ready,
				Desired:     desired,
				Status:      status,
			})
		}
	}

	if includeStatefulSets {
		statefulSets, _ := clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
		for _, s := range statefulSets.Items {
			ready := s.Status.ReadyReplicas
			desired := *s.Spec.Replicas
			status := healthStatus(int(ready), int(desired))
			helmRelease := s.Annotations["meta.helm.sh/release-name"]
			svc := findService(services, s.Labels)

			results = append(results, Workload{
				Kind:        "StatefulSet",
				Name:        s.Name,
				Namespace:   s.Namespace,
				HelmRelease: helmRelease,
				Service:     svc,
				Ready:       ready,
				Desired:     desired,
				Status:      status,
			})
		}
	}

	if includeDaemonSets {
		daemonSets, _ := clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
		for _, ds := range daemonSets.Items {
			ready := ds.Status.NumberReady
			desired := ds.Status.DesiredNumberScheduled
			status := healthStatus(int(ready), int(desired))
			helmRelease := ds.Annotations["meta.helm.sh/release-name"]
			svc := findService(services, ds.Labels)

			results = append(results, Workload{
				Kind:        "DaemonSet",
				Name:        ds.Name,
				Namespace:   ds.Namespace,
				HelmRelease: helmRelease,
				Service:     svc,
				Ready:       ready,
				Desired:     desired,
				Status:      status,
			})
		}
	}

	if len(results) == 0 {
		fmt.Println("No resources found.")
		return
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

// Map service selectors for quick lookup
func getServiceMap(clientset *kubernetes.Clientset, namespace string) map[string]map[string]string {
	ctx := context.Background()
	svcMap := make(map[string]map[string]string)

	services, _ := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	for _, svc := range services.Items {
		if svc.Spec.Selector != nil {
			svcMap[svc.Name] = svc.Spec.Selector
		}
	}
	return svcMap
}

// Basic selector match logic
func findService(services map[string]map[string]string, labels map[string]string) string {
	for svcName, selector := range services {
		matched := true
		for k, v := range selector {
			if labels[k] != v {
				matched = false
				break
			}
		}
		if matched {
			return svcName
		}
	}
	return ""
}

func healthStatus(ready, desired int) string {
	if ready == desired {
		return "Healthy"
	}
	return "Unhealthy"
}
