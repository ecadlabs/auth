FROM node:alpine as build
WORKDIR /spec
RUN npm install api2html speccy
COPY openapi.yaml logo.png /spec/
RUN $(npm bin)/speccy lint openapi.yaml
RUN $(npm bin)/api2html -o index.html -c logo.png openapi.yaml

# TODO Should use something lighter than nginx
FROM nginx
WORKDIR /usr/share/nginx/html
RUN rm -rf ./*
COPY --from=build /spec .
ENTRYPOINT /bin/bash -c "nginx -g 'daemon off;'"
