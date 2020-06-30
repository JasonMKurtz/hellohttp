FROM golang:1 as builder
COPY src/hello/ src/
COPY src/lib/ lib/
RUN go get github.com/go-sql-driver/mysql
RUN CGO_ENABLED=0 go build -o /hello src/*.go
RUN chmod +x /hello

FROM scratch
COPY --from=builder /hello /
EXPOSE 8080
CMD ["/hello"]