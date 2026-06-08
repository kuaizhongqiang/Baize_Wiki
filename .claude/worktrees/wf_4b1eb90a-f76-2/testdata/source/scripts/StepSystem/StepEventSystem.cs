
using UnityEngine;
using System;
using System.Collections.Generic;

public class StepEventSystem : MonoBehaviour
{
    private static StepEventSystem _instance;
    public static StepEventSystem Instance
    {
        get
        {
            if (_instance == null)
            {
                _instance = FindObjectOfType<StepEventSystem>();
            }
            return _instance;
        }
    }

    // 系统状态事件
    public event Action<SystemState> OnSystemStateChanged;
    public event Action OnSystemInitialized;
    public event Action OnSystemReady;

    // 步骤事件
    public event Action<StepUnit> OnStepComplete;
    public event Action<string> OnGroupSwitch;
    public event Action<StepUnit> OnStepChanged;
    public event Action OnAllStepsCompleted;

    // 相机事件
    public event Action OnCameraSwitch;

    // 条件事件
    public event Action<ConditionState> OnConditionStateChanged;
    public event Action OnConditionSatisfied;

    private void Awake()
    {
        if (_instance != null && _instance != this)
        {
            Destroy(gameObject);
            return;
        }
        _instance = this;
    }

    #region 系统事件触发方法
    public void TriggerSystemStateChanged(SystemState newState)
    {
        OnSystemStateChanged?.Invoke(newState);
    }

    public void TriggerSystemInitialized()
    {
        OnSystemInitialized?.Invoke();
    }

    public void TriggerSystemReady()
    {
        OnSystemReady?.Invoke();
    }
    #endregion

    #region 步骤事件触发方法
    public void TriggerStepComplete(StepUnit step)
    {
        OnStepComplete?.Invoke(step);
    }

    public void TriggerGroupSwitch(string groupId)
    {
        OnGroupSwitch?.Invoke(groupId);
    }

    public void TriggerStepChanged(StepUnit newStep)
    {
        OnStepChanged?.Invoke(newStep);
    }

    public void TriggerAllStepsCompleted()
    {
        OnAllStepsCompleted?.Invoke();
    }
    #endregion

    #region 相机事件触发方法
    public void TriggerCameraSwitch()
    {
        OnCameraSwitch?.Invoke();
    }
    #endregion

    #region 条件事件触发方法
    public void TriggerConditionStateChanged(ConditionState newState)
    {
        OnConditionStateChanged?.Invoke(newState);
    }

    public void TriggerConditionSatisfied()
    {
        OnConditionSatisfied?.Invoke();
    }
    #endregion

    #region 事件清理方法
    public void ClearAllEvents()
    {
        OnSystemStateChanged = null;
        OnSystemInitialized = null;
        OnSystemReady = null;
        OnStepComplete = null;
        OnGroupSwitch = null;
        OnCameraSwitch = null;
        OnConditionStateChanged = null;
        OnConditionSatisfied = null;
    }
    #endregion

    private void OnDestroy()
    {
        ClearAllEvents();
    }
}
