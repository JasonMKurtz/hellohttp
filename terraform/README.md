# terraform

Creates everything needed to run hellohttp on AWS (eu-north-1):

- `vpc.tf`: VPC with 3 public subnets (no NAT gateway)
- `iam.tf`: EKS cluster role and Auto Mode node role
- `eks.tf`: EKS Auto Mode cluster, plus cluster-admin access for `admin_principal_arns`
  (toggle with `cluster_enabled`)
- `ecr.tf`: `hellohttp` image repository

### Create
```
$ cd terraform
$ terraform init
$ terraform apply          # ~10-15 min, mostly the EKS control plane
$ $(terraform output -raw configure_kubectl)
$ kubectl get nodes        # empty until a pod needs one; Auto Mode adds nodes on demand
```

### Pause (stop paying)
Nearly all the cost is the cluster: the EKS control plane (~$73/month),
its nodes, and their public IPs. Turn it off while keeping the free parts
(VPC, IAM roles, ECR images):
```
$ terraform apply -var cluster_enabled=false   # off
$ terraform apply                              # back on (~10-15 min)
```
Clean up LoadBalancer Services, Ingresses, and PVCs first, as described in
Destroy below.

### Destroy
First delete Kubernetes objects that created AWS resources (LoadBalancer
Services, Ingresses, PVCs). Otherwise their load balancers and volumes are
left behind and block the VPC from being deleted.
```
$ kubectl delete ingress --all -A
$ kubectl get svc -A | grep LoadBalancer   # delete each of these
$ kubectl delete pvc --all -A
$ terraform destroy
```

State is local (`terraform.tfstate`, gitignored). Don't delete it while
infrastructure exists, or Terraform loses track of what it created.
