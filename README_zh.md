> **语言**：本文档为简体中文翻译，[点此查看英文原版 README.md](README.md)。

# LF

[文档](doc.md)
| [Wiki](https://github.com/gokcehan/lf/wiki)
| [#lf:matrix.org](https://matrix.to/#/#lf:matrix.org)（带 IRC 桥接）

[![Go Build](https://github.com/gokcehan/lf/actions/workflows/go.yml/badge.svg)](https://github.com/gokcehan/lf/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gokcehan/lf)](https://goreportcard.com/report/github.com/gokcehan/lf)

`lf`（即 "list files"）是一款用 Go 编写的终端文件管理器，很大程度上借鉴了 [`ranger`](https://github.com/ranger/ranger) 文件管理器。
更多信息参见 [faq](https://github.com/gokcehan/lf/wiki/FAQ)，配套的入门教程（含录屏）参见 [tutorial](https://github.com/gokcehan/lf/wiki/Tutorial)。

![icons-and-border](https://github.com/user-attachments/assets/d5623462-05ab-4921-aeb3-d377d4732f9e)
![image-preview](https://github.com/user-attachments/assets/9ff42a21-19dd-42fa-8407-5d055aa9d561)
![new-features](https://github.com/user-attachments/assets/2a9a32aa-a764-438f-a810-e687e979dcee)

## 特性

- 跨平台（Linux、macOS、BSDs、Windows）
- 单一二进制文件，无任何运行时依赖
- 启动迅速、内存占用低，得益于原生代码与静态二进制
- 异步 I/O 操作，避免界面被阻塞
- 服务器/客户端架构与远程命令，可管理多个实例
- 可通过 shell 命令扩展与配置
- 可自定义的键位绑定（默认沿用 vi 与 readline）
- 数量可观的其他特性（参见[文档](doc.md)）

## 非特性

- 标签页或窗口（交由窗口管理器或终端复用器处理更合适）
- 内置分页器/编辑器（交由你所选择的分页器/编辑器处理更合适）
- 内置的文件操作命令（交由底层 shell 工具处理更合适，包括但不限于 `mkdir`、`touch`、`chmod`、`chown`、`chgrp` 和 `ln`）

## 安装

社区维护的软件包参见 [packages](https://github.com/gokcehan/lf/wiki/Packages)。

预构建的二进制文件参见 [releases](https://github.com/gokcehan/lf/releases)。

从源码构建需要 [Go](https://go.dev/)。

在 Unix 上：

```bash
env CGO_ENABLED=0 go install -trimpath -ldflags="-s -w" github.com/gokcehan/lf@latest
```

在 Windows `cmd` 上：

```cmd
set CGO_ENABLED=0
go install -trimpath -ldflags="-s -w" github.com/gokcehan/lf@latest
```

在 Windows `PowerShell` 上：

```powershell
$env:CGO_ENABLED = '0'
go install -trimpath -ldflags="-s -w" github.com/gokcehan/lf@latest
```

## 用法

安装完成后，`lf` 命令应在当前目录下启动应用。

运行 `lf -help` 查看[命令行选项](doc.md#options)。

运行 `lf -doc` 查看[文档](doc.md)。

参见 [etc](etc) 目录，将 `lf` 集成到你的 shell 和/或编辑器中。
该目录下还能找到示例配置文件，以及示例配色（colors）和图标（icons）文件。

参见 [integrations](https://github.com/gokcehan/lf/wiki/Integrations)，将 `lf` 集成到其他工具中。

更多示例参见 [tips](https://github.com/gokcehan/lf/wiki/Tips)。

## 贡献

相关准则参见 [贡献指南](CONTRIBUTING.md)。
