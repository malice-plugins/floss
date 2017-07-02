FROM malice/alpine

LABEL maintainer "https://github.com/blacktop"

COPY . /go/src/github.com/maliceio/malice-floss
RUN apk --update add --no-cache python py-setuptools
RUN apk --update add --no-cache -t .build-deps \
                                    python-dev \
                                    build-base \
                                    mercurial \
                                    musl-dev \
                                    openssl \
                                    py-pip \
                                    bash \
                                    wget \
                                    git \
                                    gcc \
                                    go \
  && echo "Install FLOSS..." \
  && export PIP_NO_CACHE_DIR=off \
  && export PIP_DISABLE_PIP_VERSION_CHECK=on \
  && pip install https://github.com/williballenthin/vivisect/zipball/master \
  && pip install https://github.com/fireeye/flare-floss/zipball/master \
  && echo "Building scan Go binary..." \
  && cd /go/src/github.com/maliceio/malice-floss \
  && export GOPATH=/go \
  && go version \
  && go get -v \
  && go build -ldflags "-X main.Version=$(cat VERSION) -X main.BuildTime=$(date -u +%Y%m%d)" -o /bin/scan \
  && rm -rf /go /usr/local/go /usr/lib/go /tmp/* \
  && apk del --purge .build-deps

WORKDIR /malware

ENTRYPOINT ["su-exec","malice","/sbin/tini","--","scan"]
CMD ["--help"]
