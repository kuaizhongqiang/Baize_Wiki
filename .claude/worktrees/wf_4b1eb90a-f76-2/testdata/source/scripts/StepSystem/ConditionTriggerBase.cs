using UnityEngine;

/// <summary>
/// 条件触发器基类
/// 作为所有条件触发器的基础类，提供基本的生命周期管理
/// </summary>
public class ConditionTriggerBase
{
    /// <summary>
    /// 是否已经初始化
    /// 用于防止重复初始化和确保在使用前已经初始化
    /// </summary>
    protected bool isInit;

    /// <summary>
    /// 是否处于激活状态
    /// 用于控制条件触发器的启用/禁用状态
    /// </summary>
    protected bool isActive;

    /// <summary>
    /// 初始化方法
    /// 在首次使用触发器之前必须调用此方法进行初始化
    /// 该方法确保触发器只被初始化一次
    /// </summary>
    public virtual void Init()
    {
        if (isInit)
            return;
            
        isInit = true;
    }

    /// <summary>
    /// 激活触发器
    /// 使触发器开始工作，激活后才能响应条件检测
    /// 必须在初始化后才能调用此方法
    /// </summary>
    public virtual void Active()
    {
        if (!isInit)
        {
            Debug.LogError("触发器未初始化，请先调用Init方法！");
            return;
        }

        if (isActive)
            return;

        isActive = true;
    }

    /// <summary>
    /// 取消激活触发器
    /// 使触发器停止工作，不再响应条件检测
    /// 如果触发器未被激活，则此方法不会执行任何操作
    /// </summary>
    public virtual void DeActive()
    {
        if (!isActive)
            return;

        isActive = false;
    }

    /// <summary>
    /// 清理触发器
    /// 重置触发器到初始状态，清除所有状态标记
    /// 通常在触发器不再使用或需要重新初始化时调用
    /// </summary>
    public virtual void Clear()
    {
        isInit = false;
        isActive = false;
    }
}
