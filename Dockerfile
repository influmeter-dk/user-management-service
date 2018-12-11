FROM golang:alpine as builder
RUN mkdir -p /go/src/github.com/Influenzanet/user-management-service/
ADD . /go/src/github.com/Influenzanet/user-management-service/
WORKDIR /go/src/github.com/Influenzanet/user-management-service
RUN apk add --no-cache git curl \
  && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh \
  && dep ensure \
  && apk del git curl
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .
FROM scratch
COPY --from=builder /go/src/github.com/Influenzanet/user-management-service/main /app/
COPY ./configs.yaml /app/
COPY ./secret /app/secret
WORKDIR /app
EXPOSE 3200:3200
CMD ["./main"]
