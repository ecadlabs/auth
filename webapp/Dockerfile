FROM node:alpine as build-env
ARG GIT_COMMIT
ENV GIT_COMMIT ${GIT_COMMIT}
LABEL git-commit=$GIT_COMMIT

WORKDIR /app
COPY . .
RUN npm ci
RUN npm run-script build-prod

FROM nginx:alpine
COPY nginx_default.conf /etc/nginx/conf.d/default.conf
COPY --from=build-env /app/dist /usr/share/nginx/html

