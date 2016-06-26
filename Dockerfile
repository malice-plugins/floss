FROM gliderlabs/alpine

MAINTAINER blacktop, https://github.com/blacktop

ADD https://s3.amazonaws.com/build-artifacts.floss.flare.fireeye.com/travis/linux/dist/floss /usr/bin/floss
COPY . /go/src/github.com/maliceio/malice-floss
RUN apk-install -t build-deps go git mercurial \
  && set -x \
  && echo "Building scan Go binary..." \
  && cd /go/src/github.com/maliceio/malice-floss \
  && export GOPATH=/go \
  && go version \
  && go get \
  && go build -ldflags "-X main.Version=$(cat VERSION) -X main.BuildTime=$(date -u +%Y%m%d)" -o /bin/scan \
  && rm -rf /go \
  && apk del --purge build-deps

WORKDIR /malware

ENTRYPOINT ["/bin/scan"]

CMD ["--help"]
