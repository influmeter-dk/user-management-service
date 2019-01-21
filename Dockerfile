FROM phev8/go-dep-builder:latest as builder
RUN mkdir -p /go/src/github.com/Influenzanet/user-management-service/
ADD . /go/src/github.com/Influenzanet/user-management-service/
WORKDIR /go/src/github.com/Influenzanet/user-management-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .
FROM scratch
COPY --from=builder /go/src/github.com/Influenzanet/user-management-service/main /app/
COPY ./configs.yaml /app/
COPY ./secret /app/secret
WORKDIR /app
EXPOSE 3200:3200
CMD ["./main"]
