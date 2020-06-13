# hellohttp

This repository stores a simple "hello world" golang web app and the configuration needed to dockerize it.


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
