FROM golang:1 as builder
WORKDIR /tmp/hello
COPY hello.go .
RUN go build -o out/hello src/hello.go

FROM ubuntu:18.04
COPY --from=builder /tmp/hello/hello /hello
RUN chmod +x ./hello
EXPOSE 8080
CMD ["/hello"]
