FROM golang:1 as builder
WORKDIR /tmp
COPY src/ .
RUN go build -o backend *.go

FROM ubuntu:18.04
COPY --from=builder /tmp/backend /backend
RUN chmod +x ./backend
EXPOSE 9001
CMD ["/backend"]
