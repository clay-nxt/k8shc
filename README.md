# k8shc: Kubernetes Health Check

## Getting started
1. Download the binary
2. Make sure it is executable
3. Execute `aws-vault exec <profile>`
4. Update kube profile
5. Execute your binary

## Usage:
- `--getDeployments` gets all deployments
- `--getStatefulSets` gets all stateful sets
- `--getDaemonSets` gets all daemon sets
- `--getUnhealthyPods` gets all unhealthy pods
- `--getFlux` gets flux configuration and status
- `--getCronJobs` list cron data
### Optional
- `--formatOutput` outputs as json or yaml. yaml is the default
- `--namespace` select single namespace. The default is all
### ToDo

### Features ToDo
- ECR parsing and verification

## Useful 1-liners:
`k8shc --getUnhealthyPods --outputFormat=json | jq '.[] | select(.containerReason == "ImagePullBackOff")'`

## Troubleshooting:
- Check `echo $KUBECONFIG` if it is empty try `export KUBECONFIG=~/.kube/config`