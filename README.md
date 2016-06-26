![FLOSS-logo](https://raw.githubusercontent.com/maliceio/malice-floss/master/logo.png)

malice-floss
============

[![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org) [![Docker Stars](https://img.shields.io/docker/stars/malice/floss.svg)](https://hub.docker.com/r/malice/floss/) [![Docker Pulls](https://img.shields.io/docker/pulls/malice/floss.svg)](https://hub.docker.com/r/malice/floss/)

Malice FLOSS Plugin

This repository contains a **Dockerfile** of **malice/floss** for [Docker](https://www.docker.io/)'s [trusted build](https://hub.docker.com/r/malice/floss/) published to the public [DockerHub](https://hub.docker.com/).

### Dependencies

-	[gliderlabs/alpine](https://hub.docker.com/_/gliderlabs/alpine/)

### Installation

1.	Install [Docker](https://www.docker.io/).
2.	Download [trusted build](https://hub.docker.com/r/malice/floss/) from public [DockerHub](https://hub.docker.com): `docker pull malice/floss`

### Usage

```
docker run --rm -v /path/to/file:/malware:ro malice/floss FILE
```

#### Or link your own malware folder

```bash
$ docker run -v /path/to/malware:/malware:ro malice/floss FILE

Usage: floss [OPTIONS] COMMAND [arg...]

Malice FLOSS Plugin

Version: v0.1.0, BuildTime: 20160214

Author:
  blacktop - <https://github.com/blacktop>

Options:
  --verbose, -V		verbose output
  --rethinkdb value	rethinkdb address for Malice to store results [$MALICE_RETHINKDB]
  --post, -p		POST results to Malice webhook [$MALICE_ENDPOINT]
  --proxy, -x		proxy settings for Malice webhook endpoint [$MALICE_PROXY]
  --table, -t		output as Markdown table
  --rules value		YARA rules directory (default: "/rules")
  --help, -h		show help
  --version, -v		print the version

Commands:
  help	Shows a list of commands or help for one command

Run 'floss COMMAND --help' for more information on a command.
```

This will output to stdout and POST to malice results API webhook endpoint.

### Sample Output JSON:

```json
{
  "floss": {

  }
}
```

### Sample FILTERED Output JSON:

```bash
$ cat JSON_OUTPUT | jq '.[][][] .Rule'

"_Microsoft_Visual_Cpp_v50v60_MFC_"

```

### Sample Output STDOUT (Markdown Table):

---

#### floss

---

### To write results to [RethinkDB](https://rethinkdb.com)

```bash
$ docker volume create --name malice
$ docker run -d -p 28015:28015 -p 8080:8080 -v malice:/data --name rethink rethinkdb
$ docker run --rm -v /path/to/malware:/malware:ro --link rethink:rethink malice/floss -t FILE
```

### To Run on OSX

-	Install [Homebrew](http://brew.sh)

```bash
$ brew install caskroom/cask/brew-cask
$ brew cask install virtualbox
$ brew install docker
$ brew install docker-machine
$ docker-machine create --driver virtualbox malice
$ eval $(docker-machine env malice)
```

### Documentation

### Issues

Find a bug? Want more features? Find something missing in the documentation? Let me know! Please don't hesitate to [file an issue](https://github.com/maliceio/malice-floss/issues/new) and I'll get right on it.

### Credits

### License

MIT Copyright (c) 2016 **blacktop**
