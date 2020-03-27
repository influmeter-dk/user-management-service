FROM phev8/go-dep-builder:latest as builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
RUN mkdir -p /go/src/github.com/influenzanet/user-management-service/
ADD . /go/src/github.com/influenzanet/user-management-service/
WORKDIR /go/src/github.com/influenzanet/user-management-service
RUN dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o app .
FROM alpine
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/influenzanet/user-management-service/app /app/
WORKDIR /app
EXPOSE 3200:3200
CMD ["./app"]