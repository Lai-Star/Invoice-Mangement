FROM node:16.1.0-buster AS builder

COPY ./ /work
WORKDIR /work
RUN yarn install
RUN yarn build-prod

FROM nginx:1.20.0
LABEL org.opencontainers.image.source=https://github.com/monetrapp/web-ui
EXPOSE 80
COPY --from=builder /work/build /usr/share/nginx/html
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
