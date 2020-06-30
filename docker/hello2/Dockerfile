FROM golang:1 as builder
WORKDIR /tmp/hello
RUN mkdir -p code lib
RUN ls -al
RUN pwd
COPY src/hello /tmp/hello/code
COPY src/lib /tmp/hello/lib
RUN go get github.com/go-sql-driver/mysql
RUN go build -o /tmp/hello/hello /tmp/hello/code/*.go

FROM ubuntu:18.04
COPY --from=builder /tmp/hello/hello /hello
RUN chmod +x ./hello
EXPOSE 8080
CMD ["/hello"]