FROM golang:1 as builder
WORKDIR /tmp/hello
COPY src/routes/ .
RUN go build -o hello *.go

FROM ubuntu:18.04
COPY --from=builder /tmp/hello/hello /hello
RUN chmod +x ./hello
EXPOSE 8000
CMD ["/hello"]
