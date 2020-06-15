FROM golang:1 as builder
WORKDIR /tmp/hello
COPY hellohttp-foo.go .
RUN go build -o hellohttp-foo hellohttp-foo.go

FROM ubuntu:18.04
COPY --from=builder /tmp/hello/hellohttp-foo /hellohttp-foo
RUN chmod +x ./hellohttp-foo
EXPOSE 8080
CMD ["/hellohttp-foo"]
