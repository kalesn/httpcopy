# httpcopy - 流量复制工具


# 项目简介
使用Go语言开发的轻量级流量复制工具, 基于7层http协议进行录制。 完全兼容gor (goreplay)，可以使用gor 进行流量回放，流量过滤等。

主要为了解决gor使用4层录制流量丢失问题


## 功能特性
* http层流量复制


### 支持平台
> Windows、Linux、Mac OS


## 安装

### 源码安装

- 安装Go 1.11+
- `go get -d github.com/kalesn/httpcopy`
- `export GO111MODULE=on`
- 编译 `cd httpcopy/cmd/httpcopy/; go build `
-  使用方式: \
  `./httpcopy --input-http :9797 --output-file dir/xxx.file ` \
  `./httpcopy --input-http :9797 --output-http] http[s]://domain` \
  `./httpcopy --input-file dir/xxx.file --output-http http[s]://domain`
- 注：流量回放 "--output-http" 可以使用gor进行回放


