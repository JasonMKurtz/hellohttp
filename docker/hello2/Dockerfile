FROM golang:1 as builder
WORKDIR /tmp/hello
RUN mkdir -p src lib
COPY src/hello/*.go src
COPY src/lib lib
RUN go build -o hello src/*.go

FROM ubuntu:18.04
COPY --from=builder /tmp/hello/hello /hello
RUN chmod +x ./hello
EXPOSE 8080
CMD ["/hello"]
