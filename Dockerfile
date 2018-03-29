# build stage
FROM golang:alpine AS build-env
WORKDIR  /go/src/git.ecadlabs.com/ecad/auth
COPY . .
RUN CGO_ENABLED=0 go build -o auth

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /go/src/git.ecadlabs.com/ecad/auth /app/

RUN mkdir /data
# RUN /app/auth -gen_secret 256 > /data/secret.bin
# 
# VOLUME /data

CMD ["/app/auth", "-c", "/app/config.json"]
