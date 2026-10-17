# hellohttp

A small Go web app, deployed to EKS (eu-north-1) with Terraform and GitHub Actions.

### Create the cluster
```
$ cd terraform
$ echo 'budget_email = "you@example.com"' > terraform.tfvars   # gitignored
$ terraform init
$ terraform apply                                              # ~15 min
$ $(terraform output -raw configure_kubectl)
$ curl $(terraform output -raw app_url)/version
```

### Pause / resume (stop paying)
```
$ terraform apply -var cluster_enabled=false   # removes cluster, nodes, ALB, public IPs
$ terraform apply                              # back on; new app_url
```
VPC, IAM, and ECR images are kept (free or near-free). `terraform destroy` removes everything.

### Deploy
Push to `master`. CI tests, builds an amd64 image, pushes it to ECR as
`:<commit>` and `:latest`, and rolls it out one pod at a time with no
downtime. `/version` returns the deployed commit.

### Run locally
```
$ go run ./src/hello          # listens on :8080 (or $PORT)
$ go test ./...
```

### More
- [`terraform/README.md`](terraform/README.md): what each file creates, manual deploys, destroy caveats
- Terraform state is local (`terraform/terraform.tfstate`); don't lose it
- A $10/month AWS Budget emails `budget_email`
- `deployments/`, `services/`, `ingress/`, `mysql/`: legacy GKE manifests, not used
