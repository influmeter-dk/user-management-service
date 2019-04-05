FROM golang:alpine as builder
RUN mkdir -p /go/src/github.com/influenzanet/authentication-service
ADD . /go/src/github.com/influenzanet/authentication-service/
WORKDIR /go/src/github.com/influenzanet/authentication-service
RUN apk add --no-cache git && echo "installing go packages.." && while read line; do echo "$line" && go get "$line"; done < packages.txt
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .
FROM scratch
COPY --from=builder /go/src/github.com/influenzanet/user-management-service/main /app/
COPY ./configs.yaml /app/
COPY ./secret /app/secret
WORKDIR /app
EXPOSE 3200:3200
CMD ["./main"]
