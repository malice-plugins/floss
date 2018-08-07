![FLOSS-logo](https://raw.githubusercontent.com/malice-plugins/floss/master/logo.png)

# malice-floss

[![Circle CI](https://circleci.com/gh/malice-plugins/floss.png?style=shield)](https://circleci.com/gh/malice-plugins/floss) [![License](http://img.shields.io/:license-mit-blue.svg)](http://doge.mit-license.org) [![Docker Stars](https://img.shields.io/docker/stars/malice/floss.svg)](https://hub.docker.com/r/malice/floss/) [![Docker Pulls](https://img.shields.io/docker/pulls/malice/floss.svg)](https://hub.docker.com/r/malice/floss/) [![Docker Image](https://img.shields.io/badge/docker%20image-109MB-blue.svg)](https://hub.docker.com/r/malice/floss/)

Malice FLOSS Plugin

> This repository contains a **Dockerfile** of the [FLOSS](https://github.com/fireeye/flare-floss) malice plugin **malice/floss**.

---

### Dependencies

- [malice/alpine](https://hub.docker.com/r/malice/alpine/)

### Installation

1.  Install [Docker](https://www.docker.io/).
2.  Download [trusted build](https://hub.docker.com/r/malice/floss/) from public [DockerHub](https://hub.docker.com): `docker pull malice/floss`

### Usage

```bash
docker run --rm -v /path/to/file:/malware:ro malice/floss FILE

Usage: floss [OPTIONS] COMMAND [arg...]

Malice FLOSS Plugin

Version: v0.1.0, BuildTime: 20170130

Author:
  blacktop - <https://github.com/blacktop>

Options:
  --verbose, -V		    verbose output
  --timeout value       malice plugin timeout (in seconds) (default: 60) [$MALICE_TIMEOUT]
  --elasitcsearch value elasitcsearch address for Malice to store results [$MALICE_ELASTICSEARCH]
  --callback, -c	    POST results to Malice webhook [$MALICE_ENDPOINT]
  --proxy, -x		    proxy settings for Malice webhook endpoint [$MALICE_PROXY]
  --table, -t		    output as Markdown table
  --all, -a		        output ascii/utf-16 strings
  --help, -h		    show help
  --version, -v		    print the version

Commands:
  web   Create a FLOSS scan web service
  help	Shows a list of commands or help for one command

Run 'floss COMMAND --help' for more information on a command.
```

This will output to stdout and POST to malice results API webhook endpoint.

## Sample Output

### [JSON](https://github.com/malice-plugins/floss/blob/master/docs/results.json)

```json
{
  "floss": {
    "ascii": null,
    "utf-16": null,
    "decoded": [
      {
        "location": "0x401059",
        "strings": [
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
        "location": "0x404DDE",
        "strings": [
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
      },
      {
        "location": "0x401047",
        "strings": ["Ie_nkokbpAtep", "+^]g*dpi", "Ie_nkokbpD]ra=_g"]
      }
    ],
    "stack": ["cmd.exe"]
  }
}
```

### [Markdown](https://github.com/malice-plugins/floss/blob/master/docs/SAMPLE.md)

---

#### Floss

##### Decoded Strings

Location: `0x401059`

- `*lecnaC*`
- `Software\Microsoft\CurrentNetInf`
- `SYSTEM\CurrentControlSet\Control\Lsa`
- `Software\Microsoft\Windows\CurrentVersion\Policies\Explorer\Run`
- `MicrosoftZj`
- `LhbqnrnesDwhs`
- `MicrosoftHaveExit`
- `LhbqnrnesG`ud@bj\`
- `IEXPLORE.EXE`
- `/ver.htm`
- `/exe.htm`
- `/app.htm`
- `/myapp.htm`
- `/hostlist.htm`
- `.a`j-gsl\`
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

Location: `0x401047`

- `Ie_nkokbpAtep`
- `+^]g*dpi`
- `Ie_nkokbpD]ra=_g`

##### Stack Strings

- `cmd.exe`

---

## Documentation

- [To write results to ElasticSearch](https://github.com/malice-plugins/floss/blob/master/docs/elasticsearch.md)
- [To create a FLOSS scan micro-service](https://github.com/malice-plugins/floss/blob/master/docs/web.md)
- [To post results to a webhook](https://github.com/malice-plugins/floss/blob/master/docs/callback.md)

### Issues

Find a bug? Want more features? Find something missing in the documentation? Let me know! Please don't hesitate to [file an issue](https://github.com/malice-plugins/floss/issues/new)

### CHANGELOG

See [`CHANGELOG.md`](https://github.com/malice-plugins/floss/blob/master/CHANGELOG.md)

### Contributing

[See all contributors on GitHub](https://github.com/malice-plugins/floss/graphs/contributors).

Please update the [CHANGELOG.md](https://github.com/malice-plugins/floss/blob/master/CHANGELOG.md) and submit a [Pull Request on GitHub](https://help.github.com/articles/using-pull-requests/).

### TODO

- [ ] https://bitbucket.org/cse-assemblyline/alsvc_frankenstrings
- [ ] prevent URLs from being rendered as links in MarkDown :warning:

### License

MIT Copyright (c) 2016 **blacktop**
