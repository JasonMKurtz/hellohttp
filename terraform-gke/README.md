# terraform-gke

Runs hellohttp on GKE in GCP project `hellohttp-a`, sized to fit GCP's free
tier: a Standard zonal cluster in us-central1-a with one e2-micro node, and
no load balancer. A separate stack (own state) from the EKS one in
`../terraform`, built the same way so the two can be compared file by file:

| AWS (`../terraform`) | GCP (here) | |
|---|---|---|
| `vpc.tf` | `network.tf` | VPC + one regional subnet, firewall rule for the NodePort |
| `iam.tf` | `iam.tf` | Node service account (no cluster role: GKE uses its own) |
| `eks.tf` | `gke.tf` | Standard cluster + one-node pool (toggle with `cluster_enabled`) |
| `ecr.tf` | `registry.tf` | Artifact Registry `hellohttp` repo |
| `budget.tf` | `budget.tf` | $10/month budget on the billing account |
| `k8s.tf` | `k8s.tf` | Deployment, NodePort Service, optional Ingress |
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

Terraform ignores the Deployment's image after creating it, so applies don't
roll back what you deployed. A recreated Deployment (e.g. after a pause)
starts from `:latest`.

`app_url` is the node's public IP and the NodePort (30080). GKE replaces the
node on upgrades and repairs, which changes the IP; run `terraform refresh`
to update the output.

### Cost
Free, as long as this is the only thing using the free tier on the billing
account:
- Cluster management fee ($0.10/hour): covered by GKE's monthly free-tier
  credit, for one zonal or Autopilot cluster.
- The e2-micro node and its 30 GB standard disk: Compute Engine's free tier,
  one VM per billing account, us-central1/us-west1/us-east1 only.
- Artifact Registry (0.5 GB) and Cloud Build (2,500 build-minutes/month):
  free tiers.

Not free, and small: the node's public IPv4 address may be billed (about
$3.65/month), and outbound traffic over 1 GB/month.

If pods stay `Pending` with "Insufficient memory", the e2-micro is too small
for GKE's system pods plus the app. Switch to a spot e2-small (~$4/month):
```
$ terraform apply -var node_machine_type=e2-small -var node_spot=true
```

### Try the load balancer
The Ingress (an external Application Load Balancer, ~$0.025/hour) is off by
default. Turn it on to try container-native load balancing and
zero-downtime rollouts through it, then back off:
```
$ terraform apply -var ingress_enabled=true    # app_url becomes the LB's IP
$ terraform apply                              # off again
```
It takes a few minutes after the IP shows up before requests stop returning
404/502.

### Pause
```
$ terraform apply -var cluster_enabled=false   # off
$ terraform apply                              # back on (~10 min)
```
Removes the cluster and node (and the load balancer, if on). The network,
service account, and images are kept. Mostly useful if you need the free
tier for something else.

### Destroy
```
$ terraform destroy
```

Anything you create with `kubectl` instead of Terraform (LoadBalancer
Services, Ingresses, PVCs) still creates GCP resources Terraform doesn't
know about, and none of them are free. Delete those yourself before pausing
or destroying, or they keep billing and block the network from being
deleted.

State is local (`terraform.tfstate`, gitignored). Don't delete it while
infrastructure exists, or Terraform loses track of what it created.
