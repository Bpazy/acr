[English](./README.md) | 简体中文
# acr

![Build](https://github.com/Bpazy/acr/workflows/Build/badge.svg)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=Bpazy_acr&metric=alert_status)](https://sonarcloud.io/dashboard?id=Bpazy_acr)
[![Go Report Card](https://goreportcard.com/badge/github.com/Bpazy/acr)](https://goreportcard.com/report/github.com/Bpazy/acr)
![LICENSE](https://img.shields.io/github/license/Bpazy/acr)

## 介绍
`acr` 的作用是添加 Clash 规则，并使该规则立刻生效。

## 使用教程
1. 输入命令: `acr add https://www.google.com https://www.youtube.com`
2. 然后下面的这些内容会被添加到 rule-provider 指向的文件中:
```
  - DOMAIN-SUFFIX,google.com
  - DOMAIN-SUFFIX,youtube.com
```
3. 接着会调用 CFW 内置的 clash 核心的 API 重载，使上面的规则生效。

## 安装
你有很多种选择：

### 下载稳定的 Release 版本
从这里下载最新的版本 [release page](https://github.com/Bpazy/acr/releases). And put it under the `$PATH`.

以 Linux 举例:
```shell
wget -O /usr/local/bin/acr https://github.com/Bpazy/acr/releases/latest/download/acr-linux-amd64
chmod +x /usr/local/bin/acr
```

### 或者使用Go来安装
Golang 的版本需要大于等于 1.19
```shell
$ go install github.com/Bpazy/acr@latest
```
