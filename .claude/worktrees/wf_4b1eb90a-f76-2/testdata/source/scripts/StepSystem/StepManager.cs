
using UnityEngine;
using System.Collections.Generic;
using System.Linq;
using System.Collections;
using UnityEngine.SceneManagement;

public class StepManager : MonoBehaviour
{
    // 单例实例
    private static StepManager _instance;
    
    public static StepManager Instance
    {
        get
        {
            if (_instance == null)
            {
                _instance = FindObjectOfType<StepManager>();
                if (_instance == null)
                {
                    GameObject obj = new GameObject();
                    _instance = obj.AddComponent<StepManager>();
                    obj.name = typeof(StepManager).ToString();
                }
            }
            return _instance;
        }
    }

    [Header("开始步骤序号")]
    [SerializeField] public int defaultStep = 1;                // 默认步骤ID
    [Header("默认步骤延迟")]
    [SerializeField] public float defaultStepDelay = 0.1f;      // 默认步骤延迟
    [Header("执行模式")]
    [SerializeField] private StepExecutionMode executionMode;    // 执行模式
    [Header("步骤列表")] 
    [SerializeField] public List<StepUnit> steps;               // 步骤列表
    [Header("进程列表")]
    [SerializeField] public List<GameObject> processes;         // 进程列表
    
    //[HideInInspector]
    public int currentStep = 1;                                 // 当前步骤
    private List<int> _stepId = new List<int>();               // 步骤ID列表
    private Setup _setup;
    private CurrentSceneData _currentSceneData;
    public string _currentSceneName;

    // 组件引用
    [Header("系统组件")]
    [SerializeField] 
    private CameraGlobleManager _cameraManager; 
    [SerializeField]
    private StepManagerDebugger _debugManager;
    [SerializeField] 
    private StepEventSystem _eventSystem;
    [SerializeField] 
    private StepTransitionManager _transitionManager;
    // 修改属性为只读访问器
    public CameraGlobleManager cameraManager => _cameraManager;
    public StepManagerDebugger debugManager => _debugManager;
    public StepEventSystem eventSystem => _eventSystem;
    public StepTransitionManager transitionManager => _transitionManager;

    #region 状态
    private SystemState currentState = SystemState.Uninitialized;
    #endregion

    #region 生命周期
    private void Start()
    {
        _currentSceneName = SceneManager.GetActiveScene().name;
        currentState = SystemState.Ready;
        eventSystem.TriggerSystemReady();
        currentState = SystemState.Running;        
        eventSystem.TriggerSystemStateChanged(currentState);

        if (processes == null)
        {
            processes = new List<GameObject>();
            debugManager.LogWarning("processes列表未初始化，已创建新列表");
        }
    }

    private void Awake()
    {
        if (_instance != null && _instance != this)
        {
            Destroy(gameObject);
            return;
        }
        _instance = this;
        
        VerifyComponents();
        SetupSystem();
        Initialize();        
    }    

    private void OnEnable()
    {
        if (steps == null || steps.Count == 0)
        {
            debugManager.LogError("步骤列表未初始化或为空！");
            return;
        }

        int currentIndex = currentStep - 1;
        StepUnit currentStepUnit = steps[currentIndex];
        currentStepUnit.OnStepComplete += NextStep;             
    }
 
    private void Update()
    {
        if (Input.GetKeyDown(KeyCode.N))
        {
            NextStep();
        }

        if (currentState == SystemState.Running)
        {
            if (Input.GetKeyDown(KeyCode.Alpha1) || Input.GetKeyDown(KeyCode.Keypad1)) JumpToStepById(1);
            else if (Input.GetKeyDown(KeyCode.Alpha2) || Input.GetKeyDown(KeyCode.Keypad2)) JumpToStepById(2);
            else if (Input.GetKeyDown(KeyCode.Alpha3) || Input.GetKeyDown(KeyCode.Keypad3)) JumpToStepById(3);
            else if (Input.GetKeyDown(KeyCode.Alpha4) || Input.GetKeyDown(KeyCode.Keypad4)) JumpToStepById(4);
            else if (Input.GetKeyDown(KeyCode.Alpha5) || Input.GetKeyDown(KeyCode.Keypad5)) JumpToStepById(5);
        }
    }
    #endregion

    #region 初始化
    private void Initialize()
    {
        currentState = SystemState.Initialized;
        eventSystem.TriggerSystemInitialized();
        
        // 1. 加载步骤并设置基本配置
        LoadSteps(); 
        SetLastStep();
        HideAllConditionCollider();

        // 2. 设置第一步的状态
        SetFirstStep(); 

        // 3. 获取并初始化状态记录器
        // var stateRecorder = GetComponent<InitialStateRecorder>();
        // if (stateRecorder != null)
        // {
        //     stateRecorder.RecordInitialStates(steps);
        // }
        // else
        // {
        //     debugManager.LogError("未找到InitialStateRecorder组件！");
        // }
    }

    private IEnumerator DelayedStepActivation()
    {
        yield return new WaitForEndOfFrame();
        
        foreach(var step in steps)
        {
            bool shouldActive = step.stepId == 1;
            step.gameObject.SetActive(shouldActive);
            
            if(shouldActive)
            {
                step.ResetStepState();
                step.gameObject.SetActive(false);
                step.gameObject.SetActive(true);
            }
        }
    }
    #endregion

    #region 步骤方法
    public StepUnit CurrentStepUnit
    {
        get
        {
            if (steps != null && steps.Count > 0 && currentStep > 0 && currentStep <= steps.Count)
            {
                return steps[currentStep - 1];
            }
            debugManager.LogWarning("无法获取当前步骤，可能是步骤列表为空或currentStep超出范围");
            return null;
        }
    }

    public void LoadSteps()
    {
        StepUnit[] allSteps = GetComponentsInChildren<StepUnit>(true);        
        steps = allSteps.Where(step => step.transform != transform).ToList();
        foreach (StepUnit step in steps)
        {
            step.stepId = _stepId.Count + 1;
            _stepId.Add(step.stepId);
        }        
    }

    /// <summary>
    /// 进入下一个步骤（普通切换）
    /// </summary>
    public void NextStep()
    {
        if (currentState != SystemState.Running)
        {
            debugManager.LogWarning($"系统当前状态 [{currentState}] 不允许执行步骤切换");
            return;
        }

        if (steps == null || steps.Count == 0)
        {
            debugManager.LogError("步骤列表未初始化或为空！");
            return;
        }

        int currentIndex = currentStep - 1;
        
        if (currentIndex >= steps.Count - 1)
        {
            debugManager.LogWarning("已完成所有步骤");
            FinishAllStep();
            return;
        }

        StepUnit currentStepUnit = steps[currentIndex];
        currentStepUnit.OnStepComplete -= NextStep;
        
        currentStep++;
        // 使用普通切换逻辑
        transitionManager.HandleStepTransition(currentStep, currentStepUnit);
        debugManager.LogWarning($"执行普通步骤切换到：{currentStep}");
    }

    /// <summary>
    /// 跳转到指定步骤（用于ProcessingBtnController）
    /// </summary>
    public void JumpToStepById(int targetStepId)
    {
        if (currentState != SystemState.Running)
        {
            debugManager.LogWarning($"系统当前状态 [{currentState}] 不允许跳转步骤");
            return;
        }

        // 使用跳转逻辑
        transitionManager.HandleStepJump(targetStepId);
        debugManager.LogWarning($"执行步骤跳转到：{targetStepId}");
    }

    public void FinishAllStep()
    {
        currentState = SystemState.Terminated;
        eventSystem.TriggerSystemStateChanged(currentState);
        foreach (StepUnit step in steps)
        {
            step.gameObject.SetActive(false);
        }
        eventSystem.TriggerAllStepsCompleted();
        FinishStepUI();
    }

    public void FinishStepUI()
    {
        GameObject stepFinishTipsParent = Resources.Load<GameObject>("UI/StepFinishTipsParent");
        if (stepFinishTipsParent == null)
        {
            debugManager.LogError("未找到UI/StepFinishTipsParent资源");
            return;
        }
        GameObject stepFinishTips = Instantiate(stepFinishTipsParent);
        stepFinishTips.transform.SetParent(GameObject.Find("MainCanvas").transform, false);
        stepFinishTips.transform.SetAsLastSibling();
    }

    public void SetFirstStep()
    {
        int targetStep = ValidateDefaultStep();
        
        // 分步骤执行初始化
        SetFirstStepPhase(targetStep);
        SetFirstStepAnimation(targetStep);
        SetFirstStepObjActive(targetStep);
        SetFirstStepCondition(targetStep);
        
        StartCoroutine(DelayedStepActivation(targetStep));
    }
   

    public void SetLastStep()
    {
        if (steps != null && steps.Count > 0)
        {
            StepUnit lastStepUnit = steps[steps.Count - 1];
            lastStepUnit.isLastStep = true;
        }
        else
        {
            debugManager.LogWarning("步骤列表为空，无法设置最后一个步骤");
        }
    }

    public void ResetSystem()
    {
        currentStep = 1;
        foreach (var step in steps)
        {
            step.ResetStepState();
            step.gameObject.SetActive(step.stepId == 1);
        }

        currentState = SystemState.Uninitialized;
        eventSystem.TriggerSystemInitialized();
        currentState = SystemState.Ready;
        eventSystem.TriggerSystemReady();
        currentState = SystemState.Running;
        eventSystem.TriggerSystemStateChanged(currentState);

        transitionManager.HandleStepJump(1);

        if (steps.Count > 0 && steps[0].ThisStepCamera != null)
        {
            cameraManager.ForceSetCamera(steps[0].ThisStepCamera);
        }
    }
    #endregion

    #region 私有方法
    private void HideAllConditionCollider()
    {
        ConditionUnit[] allConditions = GetComponentsInChildren<ConditionUnit>(true);
        foreach (ConditionUnit condition in allConditions)
        {
            if (condition.conditionType == ConditionType.ColliderClick ||
                condition.conditionType == ConditionType.DragToCollider ||
                condition.conditionType == ConditionType.ColliderContact)
            {
                condition.targetCollider.gameObject.SetActive(false);
            }
        }
    } 

    private void SetupSystem()
    {
        GameObject setupObj = GameObject.Find("Setup");
        if (setupObj == null)
        {
            GameObject setupPrefab = Resources.Load<GameObject>("SystemPrefabs/Setup");
        
            if (setupPrefab != null)
            {
                setupObj = Instantiate(setupPrefab);
                setupObj.name = "Setup";
                DontDestroyOnLoad(setupObj);
            }
            else
            {
                debugManager.LogError("未找到Setup预制体，路径：Resources/SystemPrefabs/Setup");
                setupObj = new GameObject("Setup");
                DontDestroyOnLoad(setupObj);
            }
        }

        _setup = setupObj.GetComponent<Setup>();
        if (_setup == null)
        {
            _setup = setupObj.AddComponent<Setup>();
        }
    }

    private void CreateStepGroupManager()
    {
        StepGroupManager existingManager = FindObjectOfType<StepGroupManager>();
        if (existingManager != null)
        {
            DontDestroyOnLoad(existingManager.gameObject);
            debugManager.LogWarning("使用现有StepGroupManager实例");
            return;
        }

        GameObject stepGroupGO = new GameObject("StepGroupManager");
        StepGroupManager stepGroupManager = stepGroupGO.AddComponent<StepGroupManager>();
        stepGroupGO.transform.SetParent(transform);
        DontDestroyOnLoad(stepGroupGO);
        
        debugManager.LogWarning("StepGroupManager实例已创建");
    }

    private IEnumerator DelayedStepActivation(int targetStepId)
    {
        yield return new WaitForEndOfFrame();
        
        foreach(var step in steps)
        {
            bool shouldActive = step.stepId == targetStepId;
            step.gameObject.SetActive(shouldActive);
            
            if(shouldActive)
            {
                step.ResetStepState();
                step.gameObject.SetActive(false);
                step.gameObject.SetActive(true);
            }
        }
    }

    private int ValidateDefaultStep()
    {
        // 获取 CurrentSceneData
        var currentSceneData = FindObjectOfType<CurrentSceneData>();
        if (currentSceneData != null && currentSceneData.FirstStepIndex > 0)
        {
            // 使用 CurrentSceneData 中保存的步骤
            int targetStep = Mathf.Clamp(currentSceneData.FirstStepIndex, 1, steps.Count);
            currentStep = targetStep;
            debugManager.LogWarning($"使用CurrentSceneData中的步骤: {targetStep}");
            return targetStep;
        }
        else
        {
            // 使用默认步骤
            int targetStep = Mathf.Clamp(defaultStep, 1, steps.Count);
            currentStep = targetStep;
            debugManager.LogWarning($"使用默认步骤: {targetStep}");
            return targetStep;
        }
    }

    private void SetFirstStepPhase(int targetStep)
    {
        foreach (StepUnit step in steps)
        {
            bool isSkippedStep = step.stepId < targetStep;
            
            // 设置PhaseController状态
            DefaultPhaseController[] phaseControllers = step.GetComponentsInChildren<DefaultPhaseController>(true);
            foreach (var controller in phaseControllers)
            {
                controller.isCompletingStep = isSkippedStep;
            }
        }
    }

    private void SetFirstStepAnimation(int targetStep)
    {
        foreach (StepUnit step in steps)
        {
            bool isSkippedStep = step.stepId < targetStep;
            
            // 设置AnimationPlayer状态
            AnimationPlayer[] animationPlayers = step.GetComponentsInChildren<AnimationPlayer>(true);
            foreach (var player in animationPlayers)
            {
                player.currentState = isSkippedStep ? AnimationState.Completed : AnimationState.Ready;
            }
        }
    }

    private void SetFirstStepObjActive(int targetStep)
    {
        foreach (StepUnit step in steps)
        {
            bool isTargetStep = step.stepId == targetStep;
            step.gameObject.SetActive(isTargetStep);
        }
    }

    private void SetFirstStepCondition(int targetStep)
    {
        foreach (StepUnit step in steps)
        {
            if (step.stepId == targetStep)
            {
                var conditions = step.GetComponentsInChildren<ConditionUnit>(true);
                foreach(var condition in conditions)
                {
                    condition.gameObject.SetActive(true);
                    condition.InitCondition();
                }
            }
        }
    }

    private void VerifyComponents()
    {
        // 使用字段而不是属性
        if (_cameraManager == null) 
            _cameraManager = GetComponent<CameraGlobleManager>();
        if (_debugManager == null)
            _debugManager = GetComponent<StepManagerDebugger>();
        if (_eventSystem == null)
            _eventSystem = GetComponent<StepEventSystem>();
        if (_transitionManager == null)
            _transitionManager = GetComponent<StepTransitionManager>();
    }
    #endregion
}

#region 枚举类型
public enum SystemState
{
    Uninitialized,
    Initialized,
    Ready,
    Running,
    Terminated
}

public enum StepExecutionMode
{
    Normal,
    Debug,
    AutoPlay
}
#endregion