---
title: 配置说明
weight: 2
category: guide
---

# 配置说明

Baize Wiki 使用 `baize.yaml` 进行配置。

## 基本结构

```yaml
name: "我的 Wiki"
scan:
  paths:
    - ./docs
  exclude: []
  max_size: 10485760
output:
  dir: "./wiki"
  level: 2
```

## Level 说明

| Level | 说明 |
|-------|------|
| 1 | 平面文件，适合快速概览 |
| 2 | 结构化，按主题分组（默认） |
| 3 | 深度目录树 |
