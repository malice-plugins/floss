FROM malice/alpine:tini

MAINTAINER blacktop, https://github.com/blacktop

ENV GOLANG_VERSION 1.7.3
ENV GOLANG_SRC_URL https://golang.org/dl/go$GOLANG_VERSION.src.tar.gz
ENV GOLANG_SRC_SHA256 79430a0027a09b0b3ad57e214c4c1acfdd7af290961dd08d322818895af1ef44
COPY no-pic.patch /

COPY . /go/src/github.com/maliceio/malice-floss
RUN apk-install python py-setuptools
RUN apk-install -t .build-deps \
                    python-dev \
                    build-base \
                    mercurial \
                    musl-dev \
                    openssl \
                    py-pip \
                    bash \
                    git \
                    gcc \
                    go \
  && set -x \
  && echo "Install FLOSS..." \
  && pip install https://github.com/williballenthin/vivisect/zipball/master \
  && pip install https://github.com/fireeye/flare-floss/zipball/master \
  && echo "Upgrade to Golang $GOLANG_VERSION..." \
	&& export GOROOT_BOOTSTRAP="$(go env GOROOT)" \
	&& wget -q "$GOLANG_SRC_URL" -O golang.tar.gz \
	&& echo "$GOLANG_SRC_SHA256  golang.tar.gz" | sha256sum -c - \
	&& tar -C /usr/local -xzf golang.tar.gz \
	&& rm golang.tar.gz \
	&& cd /usr/local/go/src \
	&& patch -p2 -i /no-pic.patch \
	&& ./make.bash \
	&& rm -rf /*.patch \
  && echo "Building scan Go binary..." \
  && cd /go/src/github.com/maliceio/malice-floss \
  && export GOPATH=/go \
  && go version \
  && go get -v \
  && go build -ldflags "-X main.Version=$(cat VERSION) -X main.BuildTime=$(date -u +%Y%m%d)" -o /bin/scan \
  && rm -rf /go /usr/local/go /usr/lib/go /tmp/* \
  && apk del --purge .build-deps

WORKDIR /malware

ENTRYPOINT ["gosu","malice","/sbin/tini","--","scan"]

CMD ["--help"]
