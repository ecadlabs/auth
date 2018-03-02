# build stage
FROM golang:alpine AS build-env
WORKDIR  /go/src/git.ecadlabs.com/ecad/auth_radius/auth
ADD . .
RUN CGO_ENABLED=0 go build -o auth

# final stage
FROM alpine
WORKDIR /app
COPY --from=build-env /go/src/git.ecadlabs.com/ecad/auth_radius/auth /app/

# Install a radius client for debug/troubleshooting during development
RUN apk --no-cache add freeradius-radclient

ENTRYPOINT ./auth
