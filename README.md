# hellohttp

This repository stores a simple "hello world" golang web app and the configuration needed to dockerize it.


## Running the go app
```
$ go build -o out/hello src/hello.go
$ out/hello 
```

## Building a new docker container to run the go app
```
$ docker build -t jmliber/hellohttp:<new version> . 
```

