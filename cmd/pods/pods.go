package pods

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nexthink/k8shc/cmd/ecr_parser"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListUnhealthy(clientset *kubernetes.Clientset, namespace string) {
	ctx := context.Background()
	pods, _ := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})

	for _, pod := range pods.Items {
		if isUnhealthy(pod) {
			fmt.Printf("%s:\n", pod.Name)
			fmt.Printf("   Namespace: %s\n", pod.Namespace)

			// Pod-level conditions
			for _, cond := range pod.Status.Conditions {
				if cond.Type == "Ready" && cond.Status != "True" {
					fmt.Printf("   Pod Not Ready - Reason: %s | Message: %s\n", cond.Reason, cond.Message)
				}
			}

			for _, container := range pod.Spec.Containers {
				repo, name, tag := ecr_parser.ParseImage(container.Image)
				fmt.Printf("   Image: %s\n", container.Image)
				fmt.Printf("   Repo:  %s\n", repo)
				fmt.Printf("   Name:  %s\n", name)
				fmt.Printf("   Tag:   %s\n", tag)
			}

			// Container-level statuses
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.State.Waiting != nil {
					fmt.Printf("   Container: %s | State: Waiting | Reason: %s | Message: %s\n",
						cs.Name, cs.State.Waiting.Reason, cs.State.Waiting.Message)
				}
				if cs.State.Terminated != nil {
					fmt.Printf("   Container: %s | State: Terminated | Reason: %s | ExitCode: %d\n",
						cs.Name, cs.State.Terminated.Reason, cs.State.Terminated.ExitCode)
				}
			}

			fmt.Println()
		}
	}
}

type UnhealthyPod struct {
	PodName          string `json:"podName" yaml:"podName"`
	Namespace        string `json:"namespace" yaml:"namespace"`
	Image            string `json:"image" yaml:"image"`
	Repo             string `json:"repo" yaml:"repo"`
	Name             string `json:"name" yaml:"name"`
	Tag              string `json:"tag" yaml:"tag"`
	PodReason        string `json:"podReason,omitempty" yaml:"podReason,omitempty"`
	PodMessage       string `json:"podMessage,omitempty" yaml:"podMessage,omitempty"`
	ContainerReason  string `json:"containerReason,omitempty" yaml:"containerReason,omitempty"`
	ContainerMessage string `json:"containerMessage,omitempty" yaml:"containerMessage,omitempty"`
	ContainerState   string `json:"containerState,omitempty" yaml:"containerState,omitempty"`
}

func ListUnhealthySTRUCT(clientset *kubernetes.Clientset, namespace string, outputFormat string) {
	ctx := context.Background()
	var results []UnhealthyPod

	pods, _ := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})

	for _, pod := range pods.Items {
		if isUnhealthy(pod) {
			podReason := ""
			podMessage := ""

			for _, cond := range pod.Status.Conditions {
				if cond.Type == "Ready" && cond.Status != "True" {
					podReason = cond.Reason
					podMessage = cond.Message
				}
			}

			for _, container := range pod.Spec.Containers {
				repo, name, tag := ecr_parser.ParseImage(container.Image)

				containerReason := ""
				containerMessage := ""
				containerState := ""

				for _, cs := range pod.Status.ContainerStatuses {
					if cs.Name != container.Name {
						continue
					}
					if cs.State.Waiting != nil {
						containerState = "Waiting"
						containerReason = cs.State.Waiting.Reason
						containerMessage = cs.State.Waiting.Message
					}
					if cs.State.Terminated != nil {
						containerState = "Terminated"
						containerReason = cs.State.Terminated.Reason
						containerMessage = fmt.Sprintf("ExitCode: %d", cs.State.Terminated.ExitCode)
					}
				}

				results = append(results, UnhealthyPod{
					PodName:          pod.Name,
					Namespace:        pod.Namespace,
					Image:            container.Image,
					Repo:             repo,
					Name:             name,
					Tag:              tag,
					PodReason:        podReason,
					PodMessage:       podMessage,
					ContainerReason:  containerReason,
					ContainerMessage: containerMessage,
					ContainerState:   containerState,
				})
			}
		}
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

func isUnhealthy(pod corev1.Pod) bool {
	if pod.Status.Phase == corev1.PodSucceeded {
		return false
	}
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady && cond.Status != corev1.ConditionTrue {
			return true
		}
	}
	return false
}
