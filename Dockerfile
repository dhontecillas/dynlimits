FROM golang:1.16.0-buster as builder

COPY . /go/src/github.com/dhontecillas/reqstatsrv
WORKDIR /go/src/github.com/dhontecillas/reqstatsrv
RUN go build ./cmd/proxy

FROM bitnami/minideb:stretch

VOLUME /data

ENV DYNLIMITS_FORWARDTO_HOST "127.0.0.1"
ENV DYNLIMITS_FORWARDTO_PORT "8000"
ENV DYNLIMITS_FORWARDTO_SCHEME "http"
ENV DYNLIMITS_REDIS_ADDRESS "localhost:6379"
ENV DYNLIMITS_CATALOG_FILE "/data/catalog.json"
ENV DYNLIMITS_CATALOG_SERVER ""
ENV DYNLIMITS_CATALOG_SERVER_APIKEY ""

EXPOSE 7777

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.description="DynLimits: Rate Limiting Proxy" \
      org.label-schema.name="dynlimits" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.url="http://www.hontecillas.com" \
      org.label-schema.vcs-url="https://github.com/dhontecillas/dynlimits" \
      org.label-schema.vcs-ref=$BUILD_VCS_REF \
      org.label-schema.vendor="David Hontecillas" \
      org.label-schema.version=$BUILD_VERSION
