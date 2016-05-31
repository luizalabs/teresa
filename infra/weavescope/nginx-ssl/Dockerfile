FROM nginx
MAINTAINER E.S.B - Luizalabs
COPY default.conf /etc/nginx/conf.d/default.conf
CMD ["nginx", "-g", "daemon off;"]
RUN mkdir -p /etc/nginx/ssl
COPY app_name-wildcard.* /etc/nginx/ssl/
COPY .htpasswd /etc/nginx/.htpasswd
RUN service nginx restart
