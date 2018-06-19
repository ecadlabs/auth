FROM node:alpine as build
RUN npm install -g api2html speccy
WORKDIR /spec
COPY openapi.yaml logo.png /spec/
RUN speccy lint openapi.yaml
RUN api2html -o openapi.html -c logo.png openapi.yaml 

# TODO Should use something lighter than nginx
FROM nginx
WORKDIR /usr/share/nginx/html
RUN rm -rf ./*
COPY --from=build /spec .
ENTRYPOINT /bin/bash -c "nginx -g 'daemon off;'"
