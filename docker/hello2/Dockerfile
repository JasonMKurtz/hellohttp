FROM golang:1 as builder
WORKDIR /tmp/hello
RUN mkdir -p code lib
COPY src/hello code
COPY src/lib lib
RUN go get github.com/go-sql-driver/mysql
RUN go build -o hello code/*.go

FROM ubuntu:18.04
COPY --from=builder /tmp/hello/hello /hello
RUN chmod +x ./hello
EXPOSE 8080
CMD ["/hello"]
