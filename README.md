English | [简体中文](./README-zh_CN.md)
<h1 align="center">acr</h1>

<div align="center">

![Build](https://github.com/Bpazy/acr/workflows/Build/badge.svg)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=Bpazy_acr&metric=alert_status)](https://sonarcloud.io/dashboard?id=Bpazy_acr)
[![Go Report Card](https://goreportcard.com/badge/github.com/Bpazy/acr)](https://goreportcard.com/report/github.com/Bpazy/acr)

`acr` Means add clash rule
</div>

## Install
You have many options.

### Download stable release
Download the latest version from [release page](https://github.com/Bpazy/acr/releases). And put it under the `$PATH`.

Linux example:
```shell
wget -O /usr/local/bin/acr https://github.com/Bpazy/acr/releases/latest/download/acr-linux-amd64
chmod +x /usr/local/bin/acr
```

### Or install the latest
Golang version above 1.17
```shell
$ go install github.com/Bpazy/acr@latest
```

## Usage
1. Type urls: `acr add https://www.google.com https://www.youtube.com`
2. Then the following contents will be added to the rule-provider's file:
```
  - DOMAIN-SUFFIX,google.com
  - DOMAIN-SUFFIX,youtube.com
```
3. CFW's clash core API will be called to reload rule-provider's file.
