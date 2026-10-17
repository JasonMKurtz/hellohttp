# terraform-gke

Runs hellohttp on GKE Autopilot in GCP project `hellohttp-a`,
europe-southwest1 (Madrid). A separate stack (own state) from the EKS one in
`../terraform`, built the same way so the two can be compared file by file:

| AWS (`../terraform`) | GCP (here) | |
|---|---|---|
| `vpc.tf` | `network.tf` | VPC + one regional subnet, private nodes, no NAT |
| `iam.tf` | `iam.tf` | Node service account (no cluster role: GKE uses its own) |
| `eks.tf` | `gke.tf` | Autopilot cluster (toggle with `cluster_enabled`) |
| `ecr.tf` | `registry.tf` | Artifact Registry `hellohttp` repo |
| `budget.tf` | `budget.tf` | $10/month budget on the billing account |
| `k8s.tf` | `k8s.tf` | Deployment, Service, and an Ingress GKE turns into a load balancer |
| | `apis.tf` | Turns on the GCP services used here |
| `github.tf` | | Not yet: CI still deploys to EKS only |

### Setup (once)
```
$ gcloud auth login
$ gcloud auth application-default login     # what Terraform uses
$ gcloud config set project hellohttp-a
$ gcloud components install gke-gcloud-auth-plugin   # for kubectl
```

### Create
```
$ cd terraform-gke
$ terraform init
$ terraform apply          # ~10 min, mostly the cluster
```
The pods sit in `ImagePullBackOff` until you push an image (next step).

### Deploy an image
Cloud Build builds the image in GCP (amd64, like the nodes), so Docker isn't
needed locally. It gets a `git archive` of the committed build inputs rather
than the working tree, which has dangling symlinks (`gcloud`, `bazel-*`) that
`gcloud builds submit` can't upload. From the repo root:
```
$ REPO=$(terraform -chdir=terraform-gke output -raw image_repo)
$ TAG=$(git rev-parse --short HEAD)
$ git archive --format=tar.gz -o /tmp/hellohttp-src.tgz HEAD go.mod go.sum src docker
$ gcloud builds submit --config terraform-gke/cloudbuild.yaml \
    --substitutions _IMAGE=$REPO:$TAG,_VERSION=$TAG /tmp/hellohttp-src.tgz
$ gcloud artifacts docker tags add $REPO:$TAG $REPO:latest
$ $(terraform -chdir=terraform-gke output -raw configure_kubectl)
$ kubectl set image deployment/hellohttp hellohttp=$REPO:$TAG
$ kubectl rollout status deployment/hellohttp
$ curl $(terraform -chdir=terraform-gke output -raw app_url)/version
```
The load balancer needs a few minutes after its first deploy before it stops
returning 404/502.

Terraform ignores the Deployment's image after creating it, so applies don't
roll back what you deployed. A recreated Deployment (e.g. after a pause)
starts from `:latest`.

### Cost
- Cluster management fee ($0.10/hour): covered by GKE's free tier, which
  pays for one Autopilot or zonal cluster per billing account.
- Pods: billed by their requests (50m CPU, 64Mi each). About $5/month for 2.
- Load balancer: about $18/month, the biggest piece. It exists as long as the
  Ingress does.

So roughly $25/month while running, over the $10 budget. Pause when you're
not using it.

### Pause (stop paying)
```
$ terraform apply -var cluster_enabled=false   # off
$ terraform apply                              # back on (~10 min); new app_url
```
Terraform deletes the Ingress before the cluster, so GKE removes the load
balancer first. The network, service account, and images are kept.

### Destroy
```
$ terraform destroy
```

Anything you create with `kubectl` instead of Terraform (LoadBalancer
Services, Ingresses, PVCs) still creates GCP resources Terraform doesn't
know about. Delete those yourself before pausing or destroying, or they keep
billing and block the network from being deleted.

State is local (`terraform.tfstate`, gitignored). Don't delete it while
infrastructure exists, or Terraform loses track of what it created.
