# OpenVPNAdvanced (中文文档)

> 一个基于规则的 OpenVPN 流量分流器，支持 DoH DNS 代理、规则订阅、动态路由注入、DNS 缓存等功能。

---
[English Documentation](https://github.com/iaaaannn0/openvpnadvanced/blob/main/README.md)

## 📚 目录

- [项目概述](#项目概述)
- [功能特性](#功能特性)
- [快速开始](#快速开始)
  - [环境要求](#环境要求)
  - [构建安装](#构建安装)
  - [启动服务](#启动服务)
  - [配置本地 DNS](#配置本地-dns)
- [配置指南](#配置指南)
- [工作原理](#工作原理)
- [系统架构](#系统架构)
- [模块说明](#模块说明)
- [常见问题](#常见问题)
- [性能优化](#性能优化)
- [安全与隐私](#安全与隐私)
- [如何验证 VPN 路由](#如何验证-vpn-路由)
- [开发者指南](#开发者指南)
- [许可证](#许可证)

---

## 项目概述

本项目旨在为 OpenVPN 用户提供一个高性能、灵活的基于规则的流量分流器。它可以防止所有流量都通过 VPN，并支持规则订阅、DNS 缓存、CNAME 解析和 DNS 污染防护。

### 主要优势
- **智能流量路由**：基于规则自动路由流量
- **增强隐私保护**：支持 DoH (DNS over HTTPS) 安全 DNS 查询
- **提升性能**：DNS 缓存和优化的路由
- **易于管理**：简单的配置和规则管理
- **实时监控**：全面的日志和状态跟踪

---

## 功能特性

### 核心功能
- ✅ 本地 DNS 代理（支持 DoH / TCP / UDP）
- ✅ 自定义规则和远程订阅（自动去重和合并）
- ✅ 精确路由（通过 utunX 添加静态路由）
- ✅ 自动 VPN 接口检测（如 utun0 / utun8）
- ✅ 修复默认 macOS 网关到直连网络接口
- ✅ 支持递归 CNAME 解析
- ✅ 通过缓存实现超快速响应
- ✅ 一键启动，无需复杂设置

### 高级功能
- 🔍 域名追踪工具 (`trace.go`)
  - 详细的网络信息显示
  - 路由路径分析
  - 自动路由修复
  - CNAME 链可视化
- 📊 交互式控制台 (`ovpnctl`)
  - 实时日志查看
  - 路由测试
  - 接口管理
  - 配置重载

---

## 快速开始

### 环境要求

- Go 1.18+
- macOS（支持 `route`、`scutil` 等命令）
- 已连接的 OpenVPN 客户端（如 Tunnelblick）

### 构建安装

```bash
# 克隆仓库
git clone https://github.com/iaaaannn0/openvpnadvanced.git
cd openvpnadvanced

# 构建项目
go build -o openvpnadvanced ./cmd
```

### 启动服务

```bash
# 启动服务
sudo ./openvpnadvanced
```

### 交互式控制台

该工具提供了一个交互式命令控制台 (ovpnctl) 用于运行时控制。

#### 启动控制台

```bash
sudo ./openvpnadvanced --start
```

#### 可用命令

| 命令 | 描述 | 示例 |
|------|------|------|
| `start` | 在后台启动核心逻辑 | `start` |
| `startv` | 启动并显示实时日志 | `startv` |
| `status` | 检查服务状态 | `status` |
| `view-log` | 使用过滤器查看日志 | `view-log info` |
| `test` | 测试域名规则匹配 | `test example.com` |
| `rtest` | 测试域名解析 | `rtest example.com` |
| `show-iface` | 显示接口信息 | `show-iface` |
| `reload-config` | 重载配置 | `reload-config` |
| `clear` | 清空控制台 | `clear` |

### 域名追踪工具

`trace.go` 工具提供有关域名解析和路由的详细信息：

```bash
# 运行追踪工具
go run tools/trace.go example.com
```

#### 输出信息
- 网络信息
  - 域名解析
  - IP 地址
  - 匹配规则
  - CNAME 链
- 路由信息
  - 当前接口
  - VPN 接口
  - 默认网关
  - 路由状态

---

## 配置指南

### DNS 配置
1. 将本地 DNS 设置为 127.0.0.1
2. 在 `config.ini` 中配置 DNS 代理设置
3. 添加自定义规则或订阅规则列表

### 规则管理
- 本地规则：`assets/rule.list`
- 远程订阅：在 `config.ini` 中添加 URL
- 自动更新：在 `config.ini` 中配置

---

## 工作原理

1. **DNS 解析**
   - 本地 DNS 代理处理查询
   - 支持 DoH 安全查询
   - 缓存响应以提高性能

2. **流量路由**
   - 分析域名规则
   - 通过 VPN 或直连路由流量
   - 维护最优路由路径

3. **接口管理**
   - 检测 VPN 接口
   - 管理网络路由
   - 处理接口变更

---

## 系统架构

```
├── cmd/                 # 命令行接口
├── dnsmasq/            # DNS 代理实现
├── vpn/                # VPN 路由管理
├── tools/              # 工具集
│   └── trace.go        # 域名追踪工具
├── assets/             # 配置和规则
└── config.ini          # 主配置文件
```

---

## 模块说明

### DNS 代理 (`dnsmasq/`)
- 处理 DNS 查询
- 实现缓存
- 支持 DoH
- 管理规则

### VPN 路由 (`vpn/`)
- 管理网络接口
- 处理路由注入
- 检测 VPN 状态
- 修复路由问题

### 工具集 (`tools/`)
- 域名追踪
- 路由测试
- 接口检查
- 日志管理

---

## 常见问题

### 常见问题
1. **DNS 不工作**
   - 检查本地 DNS 设置
   - 验证 DNS 代理是否运行
   - 检查规则配置

2. **VPN 路由问题**
   - 验证 VPN 连接
   - 检查接口检测
   - 审查路由规则

3. **性能问题**
   - 清除 DNS 缓存
   - 优化规则
   - 检查网络状况

---

## 性能优化

### DNS 优化
- 实现缓存
- 优化规则匹配
- 使用高效算法

### 路由优化
- 最小化路由变更
- 优化接口检测
- 缓存路由决策

---

## 安全与隐私

### DNS 安全
- 支持 DoH
- DNS 缓存保护
- 规则验证

### 路由安全
- 安全路由注入
- 接口验证
- 访问控制

---

## 如何验证 VPN 路由

1. 使用追踪工具：
```bash
go run tools/trace.go example.com
```

2. 检查路由信息：
```bash
sudo ./openvpnadvanced --start
ovpnctl> rtest example.com
```

---

## 开发者指南

### 构建
```bash
go build -o openvpnadvanced ./cmd
```

### 测试
```bash
go test ./...
```

### 贡献
1. Fork 仓库
2. 创建特性分支
3. 提交 Pull Request

---

## 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

```
MIT License

Copyright (c) 2025

Permission is hereby granted, free of charge, to any person obtaining a copy...
```

---

欢迎 Star⭐、提 Issue、提 PR，一起完善这个强大实用的 VPN 分流工具！
