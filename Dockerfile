####################################################
# GOLANG BUILDER
####################################################
FROM golang:1.11-alpine as go_builder

RUN apk add git mercurial
COPY . /go/src/github.com/malice-plugins/floss
WORKDIR /go/src/github.com/malice-plugins/floss
RUN go get -u github.com/golang/dep/cmd/dep && dep ensure
RUN go build -ldflags "-s -w -X main.Version=v$(cat VERSION) -X main.BuildTime=$(date -u +%Y%m%d)" -o /bin/flscan

####################################################
# PLUGIN BUILDER
####################################################
FROM malice/alpine

LABEL maintainer "https://github.com/blacktop"

LABEL malice.plugin.repository = "https://github.com/malice-plugins/floss.git"
LABEL malice.plugin.category="exe"
LABEL malice.plugin.mime="application/x-dosexec"
LABEL malice.plugin.docker.engine="*"

RUN apk --update add --no-cache python py-setuptools ca-certificates
RUN apk --update add --no-cache -t .build-deps \
  python-dev \
  build-base \
  musl-dev \
  openssl \
  py-pip \
  && echo "===> Install FLOSS..." \
  && export PIP_NO_CACHE_DIR=off \
  && export PIP_DISABLE_PIP_VERSION_CHECK=on \
  && pip install https://github.com/williballenthin/vivisect/zipball/master \
  && pip install https://github.com/fireeye/flare-floss/zipball/master \
  && rm -rf /tmp/* \
  && apk del --purge .build-deps

COPY --from=go_builder /bin/flscan /bin/flscan

WORKDIR /malware

ENTRYPOINT ["su-exec","malice","flscan"]
CMD ["--help"]
