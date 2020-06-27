FROM golang:1 as builder
WORKDIR /tmp/backend
RUN mkdir -p src lib
COPY src/hellohttp-backend/*.go src
COPY src/lib lib
RUN go build -o backend src/*.go

FROM ubuntu:18.04
COPY --from=builder /tmp/backend/backend /backend
RUN chmod +x ./backend
EXPOSE 8080
CMD ["/backend"]
