FROM golang:1 as builder
# WORKDIR /tmp/hello
COPY src/hello/ src/
COPY src/lib/ lib/
RUN ls -al src/
RUN go get github.com/go-sql-driver/mysql
RUN go build -o /hello src/*.go
RUN chmod +x /hello

# FROM golang:alpine
FROM scratch
WORKDIR /
COPY --from=builder /hello /hello
EXPOSE 8080
ENTRYPOINT ["/hello"]