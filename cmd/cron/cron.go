package cron

import (
	"context"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type CronJobStatus struct {
	Name         string `json:"name" yaml:"name"`
	Namespace    string `json:"namespace" yaml:"namespace"`
	Suspended    bool   `json:"suspended" yaml:"suspended"`
	LastSchedule string `json:"lastSchedule,omitempty" yaml:"lastSchedule,omitempty"`
	ActiveJobs   int    `json:"activeJobs" yaml:"activeJobs"`
}

func ListCronJobs(clientset *kubernetes.Clientset, namespace, outputFormat string) {
	ctx := context.Background()
	cronJobs, err := clientset.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	var results []CronJobStatus

	for _, cj := range cronJobs.Items {
		lastSchedule := ""
		if cj.Status.LastScheduleTime != nil {
			lastSchedule = cj.Status.LastScheduleTime.String()
		}

		suspended := false
		if cj.Spec.Suspend != nil {
			suspended = *cj.Spec.Suspend
		}

		results = append(results, CronJobStatus{
			Name:         cj.Name,
			Namespace:    cj.Namespace,
			Suspended:    suspended,
			LastSchedule: lastSchedule,
			ActiveJobs:   len(cj.Status.Active),
		})
	}

	if len(results) == 0 {
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
