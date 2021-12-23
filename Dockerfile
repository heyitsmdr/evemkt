# Build the go app.
FROM golang:1.17
WORKDIR /go/src
COPY . /go/src
RUN CGO_ENABLED=0 GOOS=linux go build -o build/evemkt-server /go/src/cmd/evemkt-server/main.go

# Build the container for the built go app.
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/build/evemkt-server ./
CMD ["./evemkt-server"]