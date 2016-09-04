FROM ubuntu:xenial

COPY bin/arkd /usr/local/bin/arkd

RUN apt-get update \
  && apt-get install -y nginx \
  && ln -sf /dev/stdout /var/log/nginx/access.log \
  && ln -sf /dev/stderr /var/log/nginx/error.log \
  && apt-get clean

EXPOSE 80 443 6660

CMD ["/usr/local/bin/arkd"]
