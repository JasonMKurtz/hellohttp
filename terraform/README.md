# terraform

Creates everything needed to run hellohttp on AWS (eu-north-1):

- `vpc.tf`: VPC with 3 public subnets (no NAT gateway)
- `iam.tf`: EKS cluster role and Auto Mode node role
- `eks.tf`: EKS Auto Mode cluster, plus cluster-admin access for `admin_principal_arns`
  (toggle with `cluster_enabled`)
- `ecr.tf`: `hellohttp` image repository
- `k8s.tf`: hellohttp Deployment, Service, and an Ingress that Auto Mode turns
  into an internet-facing Application Load Balancer

### Create
```
$ cd terraform
$ terraform init
$ terraform apply          # ~10-15 min, mostly the EKS control plane
$ $(terraform output -raw configure_kubectl)
```

### Deploy an image
Nodes are amd64; build for that even on an arm64 Mac.
```
$ REPO=$(terraform output -raw ecr_repository_url)
$ aws ecr get-login-password --region eu-north-1 | docker login --username AWS --password-stdin ${REPO%%/*}
$ docker buildx build --platform linux/amd64 -f docker/hello2/Dockerfile -t $REPO:v1 --push ..
$ terraform apply -var image_tag=v1
$ curl $(terraform output -raw app_url)/version
```

### Pause (stop paying)
Nearly all the cost is the cluster: the EKS control plane (~$73/month),
its nodes and their public IPs, and the load balancer. Turn it off while
keeping the free parts (VPC, IAM roles, ECR images):
```
$ terraform apply -var cluster_enabled=false   # off
$ terraform apply                              # back on (~10-15 min)
```
Terraform deletes the Ingress before the cluster, so Auto Mode removes the
load balancer first.

### Destroy
```
$ terraform destroy
```

Anything you create with `kubectl` instead of Terraform (LoadBalancer
Services, Ingresses, PVCs) still creates AWS resources Terraform doesn't know
about. Delete those yourself before pausing or destroying, or they keep
billing and block the VPC from being deleted.

State is local (`terraform.tfstate`, gitignored). Don't delete it while
infrastructure exists, or Terraform loses track of what it created.
