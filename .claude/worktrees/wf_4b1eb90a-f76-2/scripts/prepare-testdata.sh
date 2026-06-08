#!/bin/bash
# 准备 Baize Wiki 测试数据
# 从 Unity 项目抽取 .cs 文件 + 创建配套测试文件

set -e

BASE="F:/Project"
TESTDATA="F:/Project/Baize_Wiki/testdata"
SOURCE="$TESTDATA/source"

echo "=== 准备测试数据 ==="

# ─── 1. .cs 文件：从不同项目精选少量代表性文件 ───

echo "[1/4] 抽取 .cs 文件..."

# GeneralFrameWorkModule → scripts/General/
for f in \
  "GeneralFrameWorkModule/.plastic_fileCache_24/Assets/Scripts/Setup.cs" \
  "GeneralFrameWorkModule/.plastic_fileCache_24/Assets/Scripts/SystemData.cs" \
  "GeneralFrameWorkModule/.plastic_fileCache_24/Assets/Scripts/Singleton/GlobalInteractiveMgr.cs" \
  "GeneralFrameWorkModule/.plastic_fileCache_24/Assets/Scripts/InteractiveObj/BaseObj.cs" \
  "GeneralFrameWorkModule/.plastic_fileCache_24/Assets/Scripts/InteractiveObj/IBaseObj.cs"; do
  src="$BASE/$f"
  if [ -f "$src" ]; then
    cp "$src" "$SOURCE/scripts/General/$(basename $f)"
  fi
done

# StepSystemProject → scripts/StepSystem/
for f in \
  "StepSystemProject/Assets/StepSystem/Core/StepManager.cs" \
  "StepSystemProject/Assets/StepSystem/Core/PhaseController.cs" \
  "StepSystemProject/Assets/StepSystem/Core/StepEventSystem.cs" \
  "StepSystemProject/Assets/StepSystem/Conditions/ConditionTriggerBase.cs" \
  "StepSystemProject/Assets/StepSystem/Conditions/Condition_UIButtonClick.cs" \
  "StepSystemProject/Assets/StepSystem/Core/StepObjectController.cs"; do
  src="$BASE/$f"
  if [ -f "$src" ]; then
    cp "$src" "$SOURCE/scripts/StepSystem/$(basename $f)"
  fi
done

# UIFunctionProject → scripts/UI/
for f in \
  "UIFunctionProject/Assets/Scripts/UIManager.cs" \
  "UIFunctionProject/Assets/Scripts/UIMgr.cs" \
  "UIFunctionProject/Assets/Scripts/UIGlobalMgr.cs"; do
  src="$BASE/$f"
  if [ -f "$src" ]; then
    cp "$src" "$SOURCE/scripts/UI/$(basename $f 2>/dev/null || echo $f)"
  fi 2>/dev/null
done

# SensorSimulation → scripts/Sensor/
for f in \
  "SensorSimulationTrainingSoftware_v1.0_2026/.plastic_fileCache_21/Assets/Scripts/Data/ProjectData.cs" \
  "SensorSimulationTrainingSoftware_v1.0_2026/.plastic_fileCache_21/Assets/Scripts/Data/QuestionData.cs" \
  "SensorSimulationTrainingSoftware_v1.0_2026/.plastic_fileCache_21/Assets/Scripts/GlobalManagement/GlobalDataMgr.cs" \
  "SensorSimulationTrainingSoftware_v1.0_2026/.plastic_fileCache_21/Assets/Scripts/UI/Panel/QuestionPanel.cs" \
  "SensorSimulationTrainingSoftware_v1.0_2026/.plastic_fileCache_21/Assets/Scripts/UI/UIBase.cs"; do
  src="$BASE/$f"
  if [ -f "$src" ]; then
    cp "$src" "$SOURCE/scripts/Sensor/$(basename $f)"
  fi
done

# VR Project → scripts/VR/
for f in \
  "RehabilitationProject_TK_VR/Assets/PlayMove.cs" \
  "RehabilitationProject_TK_VR/Assets/Scripts/Answer/ExcelUtility.cs" \
  "RehabilitationProject_TK_VR/Assets/Scripts/Answer/GameUtility.cs" \
  "RehabilitationProject_TK_VR/Assets/Script/Timeline/TimelineCtrl.cs"; do
  src="$BASE/$f"
  if [ -f "$src" ]; then
    cp "$src" "$SOURCE/scripts/VR/$(basename $f)"
  fi
done

echo "  .cs 文件就绪"

# ─── 2. .md 文件（带 frontmatter 和不带） ───

echo "[2/4] 创建 .md 文件..."

cat > "$SOURCE/docs/project-overview.md" << 'MDEOF'
---
title: 项目概览
description: 所有 Unity 项目的总体说明
tags: [unity, overview]
weight: 1
---

# 项目概览

本文档涵盖了多个 Unity 项目的架构设计和使用说明。

## 项目列表

1. **GeneralFrameWorkModule** — 通用框架模块
2. **UIFunctionProject** — UI 功能项目
3. **StepSystemProject** — 步骤系统
4. **SensorSimulationTrainingSoftware** — 传感器仿真培训软件
5. **RehabilitationProject_TK_VR** — VR 康复项目
MDEOF

cat > "$SOURCE/docs/architecture.md" << 'MDEOF'
---
title: 架构设计
description: 系统架构说明
tags: [architecture, design]
weight: 2
---

# 架构设计

## 分层架构

系统采用三层架构：

### 表现层
- UI 管理
- 场景控制
- 用户交互

### 业务逻辑层
- 数据处理
- 状态管理
- 条件判断

### 数据层
- 数据持久化
- 配置管理
MDEOF

cat > "$SOURCE/docs/guide/getting-started.md" << 'MDEOF'
# 快速开始

本文档没有 frontmatter，用于测试解析器的兼容性。

## 环境要求

- Unity 2022.3+
- .NET Standard 2.0

## 安装步骤

1. 克隆仓库
2. 打开 Unity Hub 添加项目
3. 等待导入完成
4. 运行示例场景

## 常见问题

Q: 编译报错怎么办？
A: 检查 Unity 版本是否匹配。
MDEOF

cat > "$SOURCE/docs/guide/advanced-usage.md" << 'MDEOF'
---
title: 高级用法
tags: [guide, advanced]
---

# 高级用法

## 性能优化

### 对象池
使用对象池减少实例化开销。

### LOD 组
设置合适的 LOD 层级。

> 注意：LOD 切换距离需要根据项目实际情况调整。

## 扩展开发

```csharp
public class CustomExtension : MonoBehaviour
{
    void Start()
    {
        Debug.Log("Extension loaded");
    }
}
```
MDEOF

cat > "$SOURCE/docs/api/reference.md" << 'MDEOF'
---
title: API 参考
description: 核心 API 说明文档
tags: [api, reference]
weight: 10
---

# API 参考

## 核心接口

### IBaseObj

所有可交互对象的基础接口。

### IDataManager

数据管理器接口，负责数据的读写。

## 配置说明

配置项通过 JSON 文件加载。
MDEOF

echo "  .md 文件就绪"

# ─── 3. .txt 纯文本文件 ───

echo "[3/4] 创建 .txt 文件..."

cat > "$SOURCE/docs/notes.txt" << 'EOF'
开发笔记
========

2024-01-15: 重构了 UI 管理器，改为单例模式
2024-02-20: 新增 StepSystem 条件系统
2024-03-10: 优化了数据持久化方案

待办事项:
- 完善异常处理
- 添加单元测试
- 优化编辑器工具
EOF

cat > "$SOURCE/docs/changelog.txt" << 'EOF'
Changelog
=========

v2.1.0
  - 新增 VR 交互模块
  - 修复已知崩溃问题

v2.0.0
  - 重构 StepSystem
  - 升级到 Unity 2022.3
EOF

echo "  .txt 文件就绪"

# ─── 4. 二进制文件（用于测试扫描器跳过） ───

echo "[4/4] 创建二进制测试文件..."

python3 -c "
import struct
# 创建一个含 null 字节的"二进制"文件
with open('$SOURCE/binary-test/test.bin', 'wb') as f:
    f.write(b'PNG\r\n\x1a\n')
    f.write(b'\x00\x00\x00\rIHDR\x00\x00')
    f.write(b'\x00\x00\x00\x08\x08\x02\x00\x00\x00')
    f.write(b'Binary content with null bytes: \x00\x01\x02\x03')
" 2>/dev/null || {
  # fallback: 用 echo + printf
  printf 'PNG\r\n\x1a\n' > "$SOURCE/binary-test/test.bin"
  printf '\x00\x00\x00\rIHDR\x00\x00' >> "$SOURCE/binary-test/test.bin"
}

# 创建一个 .DS_Store 模拟文件（系统文件，应被忽略规则跳过）
touch "$SOURCE/binary-test/.DS_Store"
touch "$SOURCE/binary-test/Thumbs.db"

echo "  二进制文件就绪"
echo ""
echo "=== 完成 ==="
echo ""
echo "目录结构:"
find "$SOURCE" -type f | sort | sed 's|F:/Project/Baize_Wiki/testdata/||'
echo ""
echo "文件总数: $(find "$SOURCE" -type f | wc -l)"
