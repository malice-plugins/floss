REPO=malice-plugins/floss
ORG=malice
NAME=floss
VERSION=$(shell cat VERSION)

all: build size test

build:
	docker build -t $(ORG)/$(NAME):$(VERSION) .

size:
	sed -i.bu 's/docker%20image-.*-blue/docker%20image-$(shell docker images --format "{{.Size}}" $(ORG)/$(NAME):$(VERSION)| cut -d' ' -f1)%20MB-blue/' README.md

test:
	docker run --rm $(ORG)/$(NAME):$(VERSION) --help
	docker run --rm -v $(PWD):/malware $(ORG)/$(NAME):$(VERSION) -V befb88b89c2eb401900a68e9f5b78764203f2b48264fcc3f7121bf04a57fd408 > results.json
	cat results.json | jq .
	cat results.json | jq -r .$(NAME).markdown

circle:
	http https://circleci.com/api/v1.1/project/github/${REPO} | jq '.[0].build_num' > .circleci/build_num
	http "$(shell http https://circleci.com/api/v1.1/project/github/${REPO}/$(shell cat .circleci/build_num)/artifacts${CIRCLE_TOKEN} | jq '.[].url')" > .circleci/SIZE
	sed -i.bu 's/docker%20image-.*-blue/docker%20image-$(shell cat .circleci/SIZE)-blue/' README.md
	sed -i.bu '/latest/ s/[0-9.]\{3,5\}MB/$(shell cat .circleci/SIZE)/' README.md
	sed -i.bu '/$(BUILD)/ s/[0-9.]\{3,5\}MB/$(shell cat .circleci/SIZE)/' README.

.PHONY: build size test circle
