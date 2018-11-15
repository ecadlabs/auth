# build stage
FROM golang:alpine AS build-env
WORKDIR  /go/src/github.com/ecadlabs/auth
COPY . .
RUN CGO_ENABLED=0 go build -o auth

# final stage
FROM alpine
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=build-env /go/src/github.com/ecadlabs/auth /app/

RUN mkdir /data
# RUN /app/auth -gen_secret 256 > /data/secret.bin
#
# VOLUME /data

ENTRYPOINT ["/app/auth"]
CMD ["-c", "/app/config.yaml", "-r", "/app/rbac.yaml"]
