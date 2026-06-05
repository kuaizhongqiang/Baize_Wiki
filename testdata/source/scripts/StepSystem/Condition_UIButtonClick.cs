using System;
using UnityEngine;
using UnityEngine.UI;

public class Condition_UIButtonClick : MonoBehaviour
{
    private Button targetButton;
    private ConditionUnit conditionUnit;
    private string _targetTag;
    public string UIType;
    public string UIID;
    private InstUIType instUIType;
    private string UIPath;
    private InstantiatedStepUI currentStepUI;

    public enum ConfigPath
    {
        [InspectorName("技术要求")] TechRequire = 0,
        [InspectorName("单选题")] SingleChoice = 1,
        [InspectorName("多选题")] MultipleChoice = 2
    }

    [SerializeField] 
    private ConfigPath _uiType;
    public bool isCorrect = false;    

    private void Awake()
    {
        //Debug.Log($"[按钮点击条件] {gameObject.name} 正在初始化");
        conditionUnit = GetComponent<ConditionUnit>();
        if (conditionUnit == null)
        {
            Debug.LogError($"[按钮点击条件] {gameObject.name} 找不到ConditionUnit组件");
            return;
        }
        SplitTargetTag();
        //Debug.Log($"[按钮点击条件] {gameObject.name} 初始化完成");
    }

    private void OnEnable()
    {
        Debug.Log($"[{Time.frameCount}] {gameObject.name} Condition_UIButtonClick被启用");
        
        // 重置状态
        isCorrect = false;
        
        // 确保ConditionUnit处于未检查状态
        if (conditionUnit != null)
        {
            conditionUnit.ResetCondition();
        }
        
        // 延迟初始化UI，确保其他组件都准备就绪
        StartCoroutine(DelayedUIInitialization());
    }

    private System.Collections.IEnumerator DelayedUIInitialization()
    {
        yield return null; // 等待一帧

        // 初始化UI
        if (currentStepUI == null)
        {
            InstUI();
            Debug.Log($"[{Time.frameCount}] {gameObject.name} 完成UI实例化，isCorrect={isCorrect}");
            
            // 根据UI类型设置初始状态
            InitializeIsCorrect();
        }
        else
        {
            Debug.Log($"[{Time.frameCount}] {gameObject.name} UI已存在，跳过实例化");
        }
    }

    private void OnDisable()
    {
        //Debug.Log($"[按钮点击条件] {gameObject.name} 被禁用");
        if (currentStepUI != null)
        {
            DestroyUI();
            //Debug.Log($"[按钮点击条件] {gameObject.name} UI已销毁");
        }
    }

    private void OnDestroy()
    {
        if (targetButton != null)
        {
            //Debug.Log($"[按钮点击条件] {gameObject.name} 销毁时移除按钮监听");
            targetButton.onClick.RemoveListener(OnButtonClickRight);            
        }
        isCorrect = false;
    }

    // 点击按钮直接触发条件
    public void OnButtonClickRight()
    {
        Debug.Log($"[{Time.frameCount}] {gameObject.name} 按钮被点击\n" +
                  $"- UI类型: {UIType}\n" +
                  $"- 当前isCorrect: {isCorrect}");

        // 1. 检查组件状态
        if (conditionUnit == null)
        {
            Debug.LogError($"[{Time.frameCount}] {gameObject.name} 的conditionUnit为空");
            return;
        }

        // 2. 技术要求类型，点击即为正确
        if (UIType == "0")
        {
            Debug.Log($"[{Time.frameCount}] {gameObject.name} 技术要求类型，点击即满足条件");
            isCorrect = true;
        }

        // 3. 检查答案是否正确
        if (!isCorrect)
        {
            Debug.Log($"[{Time.frameCount}] {gameObject.name} 答案不正确，不满足条件");
            return;
        }

        Debug.Log($"[{Time.frameCount}] {gameObject.name} 答案正确，准备更新条件状态");

        // 4. 先触发条件满足
        conditionUnit.DirectTriggerCondition();
        
        // 5. 延迟销毁UI
        StartCoroutine(DelayedUIDestroy());
    }

    private System.Collections.IEnumerator DelayedUIDestroy()
    {
        // 等待几帧确保状态更新完成
        yield return new WaitForSeconds(0.1f);
        
        if (currentStepUI != null)
        {
            DestroyUI();
            Debug.Log($"[{Time.frameCount}] {gameObject.name} UI已销毁");
        }
        
        isCorrect = false;
    }
    
    /// <summary>
    /// 自动查找并设置NextStepBtn按钮的点击事件
    /// </summary>
    public void AutoSetupNextButton()
    {
        if (currentStepUI != null)
        {
            //Debug.Log($"[按钮点击条件] {gameObject.name} 正在从实例化步骤UI获取按钮");
            Button button = currentStepUI.GetButton();
            if (button != null)
            {
                SetTargetButton(button);
            }
            else
            {
                Debug.LogError($"[按钮点击条件] {gameObject.name} 无法从实例化步骤UI获取按钮");
            }
        }
        else
        {
            Debug.LogError($"[按钮点击条件] {gameObject.name} 未找到实例化步骤UI引用");
        }
    }

    /// <summary>
    /// 手动设置目标按钮
    /// </summary>
    /// <param name="button">要监听的按钮</param>
    public void SetTargetButton(Button button)
    {
        if (targetButton != null)
        {
            //Debug.Log($"[按钮点击条件] {gameObject.name} 正在移除旧按钮监听");
            targetButton.onClick.RemoveListener(OnButtonClickRight);            
        }

        targetButton = button;
        if (targetButton != null)
        {
            //Debug.Log($"[按钮点击条件] {gameObject.name} 正在为 {targetButton.gameObject.name} 设置新的按钮监听");
            targetButton.onClick.AddListener(OnButtonClickRight);            
        }
    }

    // 获取ConditionUnit的targetTag,并赋值给_targetTag
    public void GetTargetTag()
    {        
        _targetTag = conditionUnit.targetTag;
        //Debug.Log($"[按钮点击条件] {gameObject.name} 成功获取目标标签: {_targetTag}");
    }

    // 对_targetTag的值进行拆解，以"_"符号为界限，之前的值赋值给UIType，之后的值赋值给UIID
    public void SplitTargetTag()
    {
        GetTargetTag();
        string[] strArray = _targetTag.Split('_');
        UIType = strArray[0];
        UIID = strArray[1];
        //Debug.Log($"[按钮点击条件] {gameObject.name} 拆分标签 - UI类型: {UIType}, UIID: {UIID}");
    }

    /// <summary>
    /// 实例化预制体并赋值
    /// </summary>
    private void InstUI()
    {
        // 获取路径配置
        InstUIPath();
        
        // 加载预制体资源
        GameObject uiPrefab = Resources.Load<GameObject>(UIPath);
        if (uiPrefab == null)
        {
            Debug.LogError($"[按钮点击条件] {gameObject.name} 无法加载UI预制体，路径: {UIPath}");
            return;
        }

        // 查找MainCanvas
        GameObject mainCanvas = GameObject.FindWithTag("MainCanvas");
        if (mainCanvas == null)
        {
            Debug.LogError($"[按钮点击条件] {gameObject.name} 无法找到标签为'MainCanvas'的画布");
            return;
        }

        // 实例化并设置父对象
        GameObject instantiatedUI = Instantiate(uiPrefab, mainCanvas.transform, false);
        //Debug.Log($"[按钮点击条件] {gameObject.name} 已实例化UI预制体: {instantiatedUI.name}");
        
        // 获取InstantiatedStepUI组件
        currentStepUI = instantiatedUI.GetComponent<InstantiatedStepUI>();
        if (currentStepUI != null)
        {
            //Debug.Log($"[按钮点击条件] {gameObject.name} 已找到实例化步骤UI组件");
            // 设置UI类型
            currentStepUI.SetUIType(UIType);
            
            // 获取或添加ChoiceUIManager组件
            var choiceManager = instantiatedUI.GetComponent<ChoiceUIManager>();
            if (choiceManager == null && (UIType == "1" || UIType == "2"))
            {
                //Debug.Log($"[按钮点击条件] {gameObject.name} 添加ChoiceUIManager组件");
                choiceManager = instantiatedUI.AddComponent<ChoiceUIManager>();
            }

            if (choiceManager != null)
            {
                //Debug.Log($"[按钮点击条件] {gameObject.name} 设置ChoiceUIManager的buttonClickCondition引用");
                choiceManager.SetButtonClickCondition(this);
                
                // 设置ChoiceUIInstance引用
                choiceManager.ChoiceUIInstance = currentStepUI;
                //Debug.Log($"[按钮点击条件] {gameObject.name} 设置ChoiceUIManager的ChoiceUIInstance引用");

                // 根据UIType设置正确的ChoiceUIType
                if (UIType == "1")
                {
                    choiceManager.SetUIType(ChoiceUIType.SingleChoice);
                    //Debug.Log($"[按钮点击条件] {gameObject.name} 设置为单选题类型");
                }
                else if (UIType == "2")
                {
                    choiceManager.SetUIType(ChoiceUIType.MultipleChoice);
                    //Debug.Log($"[按钮点击条件] {gameObject.name} 设置为多选题类型");
                }
            }
            
            // 自动设置按钮事件
            AutoSetupNextButton();
        }
        else
        {
            Debug.LogError($"[按钮点击条件] {gameObject.name} 在实例化的预制体上未找到实例化步骤UI组件");
        }
    }
    
    // 确定实例化物体的路径
    private void InstUIPath()
    {
        // 设置UI路径
        if (UIType == "0")
        {
            instUIType = InstUIType.TechRequire;
            UIPath = "UI/InstUIParent";
            //Debug.Log($"[按钮点击条件] {gameObject.name} 设置为技术要求UI");
        }
        else if (UIType == "1")
        {
            instUIType = InstUIType.SingleChoice;
            UIPath = "UI/SingleChoiceParent";
            //Debug.Log($"[按钮点击条件] {gameObject.name} 设置为单选题UI");
        }
        else if (UIType == "2")
        {
            instUIType = InstUIType.MultipleChoice;
            UIPath = "UI/SingleChoiceParent";
            //Debug.Log($"[按钮点击条件] {gameObject.name} 设置为多选题UI");
        }
        else
        {
            Debug.LogError($"[按钮点击条件] {gameObject.name} 收到无效的UI类型: {UIType}");
        }
        
        //Debug.Log($"[按钮点击条件] {gameObject.name} 最终UI路径为: {UIPath}");
    }

    private void InitializeIsCorrect()
    {
        Debug.Log($"[{Time.frameCount}] {gameObject.name} 初始化isCorrect状态\n" +
                  $"- UI类型: {UIType}");
        
        // 默认设置为false
        isCorrect = false;
        
        // 只有在技术要求类型且UI已经实例化的情况下才设置为true
        if (UIType == "0" && currentStepUI != null) 
        {
            Debug.Log($"[{Time.frameCount}] {gameObject.name} 技术要求类型，等待用户确认");
            // 技术要求类型需要用户点击确认按钮
            isCorrect = false;
        }
        
        Debug.Log($"[{Time.frameCount}] {gameObject.name} isCorrect初始化完成: {isCorrect}");
    }

    // 销毁被实例化的预制体，并放置内存泄漏
    private void DestroyUI()
    {
        DestroyImmediate(currentStepUI.gameObject);
    }

    
}

public enum InstUIType
{
    TechRequire,    //技术要求，对应UIType = 0
    SingleChoice,   //单选题，对应UIType = 1
    MultipleChoice  //多选题，对应UIType = 2
}
