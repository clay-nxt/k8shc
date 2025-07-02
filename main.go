package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/nexthink/k8shc/cmd/cron"
	"github.com/nexthink/k8shc/cmd/flux"
	"github.com/nexthink/k8shc/cmd/kubeclient"
	"github.com/nexthink/k8shc/cmd/pods"
	"github.com/nexthink/k8shc/cmd/workloads"
)

var (
	version = "dev"
)

func main() {
	verFlag := flag.Bool("version", false, "Show version")
	namespace := flag.String("namespace", "", "Target namespace (default: all namespaces)")
	includeDeployments := flag.Bool("getDeployments", false, "Check Deployments (default: false)")
	includeStatefulSets := flag.Bool("getStatefulSets", false, "Check StatefulSets (default: false)")
	includeDaemonSets := flag.Bool("getDaemonSets", false, "Check DaemonSets (default: false)")
	checkUnhealthySTRUCT := flag.Bool("getUnhealthyPods", false, "Check for unhealthy pods and output json (default: false)")
	fluxCheckAndList := flag.Bool("getFlux", false, "Check Flux resources (default: false)")
	getCronJobs := flag.Bool("getCronJobs", false, "Check CronJobs (default: false)")
	outputFormat := flag.String("outputFormat", "yaml", "Outpout format: json or yaml(default)")

	flag.Parse()

	if *verFlag {
		fmt.Println("k8shc version", version)
		os.Exit(0)
	}
	config := kubeclient.GetRestConfig()
	client := kubeclient.Connect()

	if *fluxCheckAndList {
		flux.ListKustomizationsSTRUCT(config, *namespace, *outputFormat)
	}

	if *includeDeployments || *includeStatefulSets || *includeDaemonSets {
		workloads.ListSTRUCT(client, *namespace, *includeDeployments, *includeStatefulSets, *includeDaemonSets, *outputFormat)

	}
	if *checkUnhealthySTRUCT {
		pods.ListUnhealthySTRUCT(client, *namespace, *outputFormat)
	}
	if *getCronJobs {
		cron.ListCronJobs(client, *namespace, *outputFormat)
	}

	if !*getCronJobs && !*includeDeployments && !*includeStatefulSets && !*includeDaemonSets && !*checkUnhealthySTRUCT && !*fluxCheckAndList {
		fmt.Println("No output selected. Use --getDeployments, --getStatefulSets, --getDaemonSets, --getUnhealthyPods, or --getFlux.")
		fmt.Println("Optional: --formatOutput=json|yaml - yaml is default")
		fmt.Println("Optional: --namespace=<namespace> - all namespaces is default")
		return
	}

}
