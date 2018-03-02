FROM alpine:3.7
RUN apk --update add curl ca-certificates tzdata mailcap && update-ca-certificates
WORKDIR /usr/src/app
ENTRYPOINT  ["/usr/src/app/goapp"]
