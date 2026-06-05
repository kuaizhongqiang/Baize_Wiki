
using UnityEngine;
using System.Collections.Generic;
using System.Text;

/// <summary>
/// 步骤物体控制器
/// 负责管理步骤中物体的显示和隐藏
/// </summary>
[RequireComponent(typeof(StepUnit))]
public class StepObjectController : MonoBehaviour
{
    public StepUnit stepUnit;
    
    // 物体列表引用，从StepUnit获取
    private List<GameObject> ShowObjStart;
    private List<GameObject> HideObjStart;
    private List<GameObject> ShowObjEnd;
    private List<GameObject> HideObjEnd;
    
    private void Awake()
    {
        // 初始化列表
        ShowObjStart = new List<GameObject>();
        HideObjStart = new List<GameObject>();
        ShowObjEnd = new List<GameObject>();
        HideObjEnd = new List<GameObject>();

        stepUnit = GetComponent<StepUnit>();
        if (stepUnit == null)
        {
            Debug.LogError($"[{Time.frameCount}] {gameObject.name} StepObjectController 需要 StepUnit 组件！");
            return;
        }
    }

    private void OnEnable()
    {
        // 确保组件引用存在
        if (stepUnit == null)
        {
            stepUnit = GetComponent<StepUnit>();
        }
    }

    /// <summary>
    /// 初始化物体列表
    /// </summary>
    public void InitializeLists(List<GameObject> showStart, List<GameObject> hideStart, 
                              List<GameObject> showEnd, List<GameObject> hideEnd)
    {
        ShowObjStart = showStart;
        HideObjStart = hideStart;
        ShowObjEnd = showEnd;
        HideObjEnd = hideEnd;
    }

    /// <summary>
    /// 执行开始时的物体显隐控制
    /// </summary>
    public void ShowStartObjects(bool debugMode = false)
    {
        if (debugMode)
        {
            LogObjectVisibility("步骤开始");
        }
        // 设置当前步骤的物体状态
        SetObjectsActive(ShowObjStart, true);
        SetObjectsActive(HideObjStart, false);
    }

    /// <summary>
    /// 执行结束时的物体显隐控制
    /// </summary>
    public void ShowEndObjects(bool debugMode = false)
    {
        if (debugMode)
        {
            LogObjectVisibility("步骤结束");
        }

        SetObjectsActive(ShowObjEnd, true);
        SetObjectsActive(HideObjEnd, false);
    }

    /// <summary>
    /// 重置所有物体状态
    /// </summary>
    public void ResetObjectStates()
    {
        SetObjectsActive(ShowObjStart, false);
        SetObjectsActive(HideObjStart, true);
        SetObjectsActive(ShowObjEnd, false);
        SetObjectsActive(HideObjEnd, true);
    }

    /// <summary>
    /// 处理跳转时的物体状态
    /// </summary>
    public void HandleJumpState(bool isTargetStep)
    {
        if (isTargetStep)
        {
            ShowStartObjects();
        }
        else
        {
            ShowStartObjects();
            ShowEndObjects();
        }
    }

    private void SetObjectsActive(List<GameObject> objects, bool state)
    {
        if (objects == null) return;
        
        foreach (var obj in objects)
        {
            if (obj != null) 
            {
                obj.SetActive(state);
            }
        }
    }

    private void LogObjectVisibility(string phase)
    {
        StringBuilder debug = new StringBuilder();
        debug.AppendLine($"\n[StepObjectController] {gameObject.name} - {phase}时物体显隐状态:");
        
        LogObjectList(debug, "显示的物体", ShowObjStart);
        LogObjectList(debug, "隐藏的物体", HideObjStart);
        
        Debug.Log(debug.ToString());
    }

    private void LogObjectList(StringBuilder debug, string title, List<GameObject> objects)
    {
        if (objects != null && objects.Count > 0)
        {
            debug.AppendLine(title + ":");
            foreach (var obj in objects)
            {
                debug.AppendLine($"  └─ {(obj != null ? obj.name : "null")}");
            }
        }
    }
}
