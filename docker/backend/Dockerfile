FROM golang:1 as builder
WORKDIR /tmp/backend
RUN mkdir -p /tmp/backend/src /tmp/backend/lib
COPY src/hellohttp-backend/*.go /tmp/backend/src
COPY src/lib /tmp/backend/lib
RUN go build -o /tmp/backend/backend /tmp/backend/src/*.go

FROM ubuntu:18.04
COPY --from=builder /tmp/backend/backend /backend
RUN chmod +x ./backend
EXPOSE 8080
CMD ["/backend"]
