FROM ubuntu:20.04

RUN apt-get update && apt-get install -y tzdata ca-certificates

EXPOSE 4000
VOLUME ["/etc/monetr"]
ENTRYPOINT ["/usr/bin/monetr"]
CMD ["serve"]

ARG REVISION
LABEL org.opencontainers.image.url=https://github.com/monetr/monetr
LABEL org.opencontainers.image.source=https://github.com/monetr/monetr
LABEL org.opencontainers.image.authors=elliot.courant@monetr.app
LABEL org.opencontainers.image.revision=$REVISION
LABEL org.opencontainers.image.vendor="monetr"
LABEL org.opencontainers.image.licenses="BSL-1.1"
LABEL org.opencontainers.image.title="monetr"
LABEL org.opencontainers.image.description="monetr's budgeting application"

COPY ./build/monetr /usr/bin/monetr
