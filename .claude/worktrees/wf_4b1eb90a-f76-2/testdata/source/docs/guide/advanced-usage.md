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
