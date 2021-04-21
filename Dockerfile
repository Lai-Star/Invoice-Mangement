FROM nginx:1.20.0
LABEL org.opencontainers.image.source=https://github.com/monetrapp/web-ui
EXPOSE 80
COPY ./build /usr/share/nginx/html
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
