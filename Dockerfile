FROM alpine:3.5
MAINTAINER Audun Strand <audunstrand@gmail.com>

RUN adduser -u 10001 -D -h /app app

USER app

CMD ["/opt/bin/deployer"]

COPY deployer /opt/bin/deployer
