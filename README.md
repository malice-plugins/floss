![FLOSS-logo](https://raw.githubusercontent.com/maliceio/malice-floss/master/logo.png)

malice-floss
============

[![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org) [![Docker Stars](https://img.shields.io/docker/stars/malice/floss.svg)](https://hub.docker.com/r/malice/floss/) [![Docker Pulls](https://img.shields.io/docker/pulls/malice/floss.svg)](https://hub.docker.com/r/malice/floss/)

Malice FLOSS Plugin

This repository contains a **Dockerfile** of the [FLOSS](https://github.com/fireeye/flare-floss) malice plugin **malice/floss**.

### Dependencies

-	[gliderlabs/alpine](https://hub.docker.com/_/gliderlabs/alpine/)

### Installation

1.	Install [Docker](https://www.docker.io/).
2.	Download [trusted build](https://hub.docker.com/r/malice/floss/) from public [DockerHub](https://hub.docker.com): `docker pull malice/floss`

### Usage

```
docker run --rm -v /path/to/file:/malware:ro malice/floss FILE
```

```bash
Usage: floss [OPTIONS] COMMAND [arg...]

Malice FLOSS Plugin

Version: v0.1.0, BuildTime: 20160627

Author:
  blacktop - <https://github.com/blacktop>

Options:
  --verbose, -V      verbose output
  --rethinkdb value  rethinkdb address for Malice to store results [$MALICE_RETHINKDB]
  --post, -p         POST results to Malice webhook [$MALICE_ENDPOINT]
  --proxy, -x        proxy settings for Malice webhook endpoint [$MALICE_PROXY]
  --table, -t        output as Markdown table
  --help, -h         show help
  --version, -v      print the version

Commands:
  help  Shows a list of commands or help for one command

Run 'floss COMMAND --help' for more information on a command.
```

This will output to stdout and POST to malice results API webhook endpoint.

### Sample Output JSON:

```json
{
  "floss": {
    "decoded": [
      {
        "Location": "0x401047",
        "Strings": [
          "Ie_nkokbpAtep",
          "+^]g*dpi",
          "Ie_nkokbpD]ra=_g"
        ]
      },
      {
        "Location": "0x401059",
        "Strings": [
          "*lecnaC*",
          "Software\\Microsoft\\CurrentNetInf",
          "SYSTEM\\CurrentControlSet\\Control\\Lsa",
          "Software\\Microsoft\\Windows\\CurrentVersion\\Policies\\Explorer\\Run",
          "MicrosoftZj",
          "LhbqnrnesDwhs",
          "MicrosoftHaveExit",
          "LhbqnrnesG`ud@bj",
          "IEXPLORE.EXE",
          "/ver.htm",
          "/exe.htm",
          "/app.htm",
          "/myapp.htm",
          "/hostlist.htm",
          ".a`j-gsl",
          "/SomeUpList.htm",
          "/SomeUpVer.htm",
          "www.flyeagles.com",
          "www.km-nyc.com",
          "/restore",
          "/dizhi.gif",
          "/connect.gif",
          "\\$NtUninstallKB900727$",
          "\\netsvc.exe",
          "\\netscv.exe",
          "\\netsvcs.exe",
          "System Idle Process",
          "Program Files",
          "\\Internet Exp1orer",
          "forceguest",
          "AudioPort",
          "AudioPort.sys",
          "SYSTEM\\CurrentControlSet\\Services",
          "SYSTEM\\ControlSet001\\Services",
          "SYSTEM\\ControlSet002\\Services",
          "\\drivers\\",
          "\\DriverNum.dat"
        ]
      },
      {
        "Location": "0x40511A",
        "Strings": [
          "\\A|{@"
        ]
      },
      {
        "Location": "0x403DDA",
        "Strings": [
          ",pVA."
        ]
      },
      {
        "Location": "0x404DDE",
        "Strings": [
          "SMBs",
          "NTLMSSP",
          "Windows 2000 2195",
          "Windows 2000 5.0",
          "SMBr",
          "PC NETWORK PROGRAM 1.0",
          "LANMAN1.0",
          "Windows for Workgroups 3.1a",
          "LM1.2X002",
          "LANMAN2.1",
          "NT LM 0.12"
        ]
      }
    ],
    "stack": [
      "\\A|{@",
      "CAAA\\",
      "cmd.exe"
    ]
  }
}
```

### Sample Output STDOUT (Markdown Table):

---

#### Floss

##### Decoded Strings

Location: `0x401047`
 - `Ie_nkokbpAtep`
 - `+^]g*dpi`
 - `Ie_nkokbpD]ra=_g`

Location: `0x401059`
 - `*lecnaC*`
 - `Software\Microsoft\CurrentNetInf`
 - `SYSTEM\CurrentControlSet\Control\Lsa`
 - `Software\Microsoft\Windows\CurrentVersion\Policies\Explorer\Run`
 - `MicrosoftZj`
 - `LhbqnrnesDwhs`
 - `MicrosoftHaveExit`
 - `LhbqnrnesG`ud@bj`
 - `IEXPLORE.EXE`
 - `/ver.htm`
 - `/exe.htm`
 - `/app.htm`
 - `/myapp.htm`
 - `/hostlist.htm`
 - `.a`j-gsl`
 - `/SomeUpList.htm`
 - `/SomeUpVer.htm`
 - `www.flyeagles.com`
 - `www.km-nyc.com`
 - `/restore`
 - `/dizhi.gif`
 - `/connect.gif`
 - `\$NtUninstallKB900727$`
 - `\netsvc.exe`
 - `\netscv.exe`
 - `\netsvcs.exe`
 - `System Idle Process`
 - `Program Files`
 - `\Internet Exp1orer`
 - `forceguest`
 - `AudioPort`
 - `AudioPort.sys`
 - `SYSTEM\CurrentControlSet\Services`
 - `SYSTEM\ControlSet001\Services`
 - `SYSTEM\ControlSet002\Services`
 - `\drivers\`
 - `\DriverNum.dat`

Location: `0x40511A`
 - `\A|{@`

Location: `0x403DDA`
 - `,pVA.`

Location: `0x404DDE`
 - `SMBs`
 - `NTLMSSP`
 - `Windows 2000 2195`
 - `Windows 2000 5.0`
 - `SMBr`
 - `PC NETWORK PROGRAM 1.0`
 - `LANMAN1.0`
 - `Windows for Workgroups 3.1a`
 - `LM1.2X002`
 - `LANMAN2.1`
 - `NT LM 0.12`

##### Stack Strings
 - `\A|{@`
 - `CAAA\`
 - `cmd.exe`

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

### TODO

-	[ ] Cover all possible string type outputs (ASCII/UTF-16) ?

### License

MIT Copyright (c) 2016 **blacktop**
