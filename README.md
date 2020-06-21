# hellohttp

This repository stores a simple "hello world" golang web app and the configuration needed to dockerize it.

### Build & Deploy 
```
$ vi src/hello.go 
$ git commit -am "Code change"
$ git push 
... Watch circleci pipeline ... 
$ curl http://hellohttp.jkurtz.net # observe code change
```


### Running the go app
```
$ go build -o out/hello src/hello.go
$ out/hello 
```

### Building a new docker container to run the go app
```
$ docker build -t jmliber/hellohttp:<new version> . 
```

### Manually deploying a new version of the app
```
$ vi src/hello.go
$ docker build -t jmliber/hellohttp:<new version> 
$ vi deployment.yml (update old version string to new one)
$ kubectl apply -f deployment.yml
```

### Deploying a new version with dockerhub (CI only)
```
$ vi src/hello.go
$ git tag <new version>
$ git push --tags 
$ # wait for build on dockerhub
$ vi deployment.yml (update old version string to new one)
$ kubectl apply -f deployment.yml
```

### Get latest image version
```
$ go run cli/get-latest.go --image hellohttp
1.2
```

--- GKE (Google K8S Engine) ---
### Setup
We need
- to route internet traffic to our services (ingress.yml)
- our load balancers to know the containers where our app lives (service.yml)
- our apps to know where to deploy new versions (deployment.yml)

```
$ kubectl apply -f ingress/ingress.yml 
$ kubectl apply -f services/hellohttp.yml # load balancer for 'hellohttp' app
$ kubectl apply -f services/hellohttp-foo.yml # load balancer for 'hellohttp-foo' app
$ kubectl apply -f deployments/hellohttp.yml # 'hellohttp' app
$ kubectl apply -f deployments/hellohttp-foo.yml # 'hellohttp-foo' app
```

Let's make sure they all started successfully.

```
$ kubectl get ingress
NAME           HOSTS   ADDRESS          PORTS   AGE
hellohttp-in   *       34.120.140.223   80      178m
```

```
$ kubectl describe ing hellohttp-in 
Name:             hellohttp-in
Namespace:        default
Address:          34.120.140.223
Default backend:  hellohttp:8080 (10.44.0.13:8080,10.44.0.14:8080)
Rules:
  Host        Path  Backends
  ----        ----  --------
  *
              /foo   hellohttp-foo:8080 (10.44.11.2:8080,10.44.11.3:8080)
Annotations:  ingress.kubernetes.io/backends: {"k8s-be-32153--7e366c60b907700f":"HEALTHY","k8s-be-32261--7e366c60b907700f":"HEALTHY"}
              ingress.kubernetes.io/forwarding-rule: k8s-fw-default-hellohttp-in--7e366c60b907700f
              ingress.kubernetes.io/target-proxy: k8s-tp-default-hellohttp-in--7e366c60b907700f
              ingress.kubernetes.io/url-map: k8s-um-default-hellohttp-in--7e366c60b907700f
Events:       <none>
```

``` 
$ kubectl get svc
NAME            TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
hellohttp       LoadBalancer   10.47.255.127   10.150.0.6    8080:32261/TCP   6h34m
hellohttp-foo   LoadBalancer   10.47.244.132   10.150.0.8    8080:32153/TCP   179m
kubernetes      ClusterIP      10.47.240.1     <none>        443/TCP          6h35m
```

```
$ kubectl get deploy
NAME            READY   UP-TO-DATE   AVAILABLE   AGE
hellohttp       2/2     2            2           6h34m
hellohttp-foo   2/2     2            2           3h55m
```

```
$ kubectl get nodes # the hosts where the containers live
NAME                                       STATUS   ROLES    AGE     VERSION
gke-hellohttp-default-pool-541dbf73-ljlv   Ready    <none>   6h36m   v1.14.10-gke.36
gke-hellohttp-default-pool-541dbf73-mt3f   Ready    <none>   4h25m   v1.14.10-gke.36
```

```
$ kubectl get pods # the containers 
NAME                             READY   STATUS    RESTARTS   AGE
hellohttp-88f7ccdbb-sgwl2        1/1     Running   0          6h38m
hellohttp-88f7ccdbb-ssp9b        1/1     Running   0          6h38m
hellohttp-foo-6b455d8f4b-d7b44   1/1     Running   0          3h59m
hellohttp-foo-6b455d8f4b-kqf9c   1/1     Running   0          178m
```

### If encountering auth errors to GKE:
```
$ gcloud container clusters get-credentials <cluster>
```

