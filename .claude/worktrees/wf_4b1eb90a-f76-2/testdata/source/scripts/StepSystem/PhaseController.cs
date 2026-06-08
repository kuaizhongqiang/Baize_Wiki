
using UnityEngine;
using System;
using System.Collections;

[RequireComponent(typeof(ConditionUnit))]
public abstract class PhaseController : MonoBehaviour
{
    // 当前阶段
    //[HideInInspector]
    public PhaseState currentPhase = PhaseState.Initialize;

    // 事件
    public event Action<PhaseState> OnPhaseChanged;
    public event Action OnPhaseCompleted;
    public event Action<float> OnProgressUpdated;

    public ConditionUnit conditionUnit;
    public AnimationPlayer animationPlayer;
    protected bool isConditonMet = false;
    public bool IsConditionMet => isConditonMet;
    //[HideInInspector]
    public bool AnimationCompleted = false;

    protected float delayTime = 0f;
    protected bool isDelayCoroutineRunning;

    #region 生命周期
    protected virtual void Awake()
    {
        //Debug.Log($"[{Time.frameCount}] {gameObject.name} PhaseController.Awake开始执行");

        // 1. 直接初始化组件
        InitializeComponents();
    }

    private void InitializeComponents()
    {
        // 1. 获取ConditionUnit组件（最基础的依赖）
        SetConditionUnit();
        if (conditionUnit == null)
        {
            //Debug.LogError($"[{Time.frameCount}] {gameObject.name} 无法获取ConditionUnit组件");
            return;
        }
        //Debug.Log($"[{Time.frameCount}] {gameObject.name} 成功获取ConditionUnit组件");

        // 2. 获取StepUnit组件和配置
        var stepUnit = GetComponent<StepUnit>();
        if (stepUnit != null)
        {
            delayTime = stepUnit.delay;
            //Debug.Log($"[{Time.frameCount}] {gameObject.name} 成功获取StepUnit组件，延迟时间：{delayTime}");
        }
        else
        {
            //Debug.LogError($"[{Time.frameCount}] {gameObject.name} 无法获取StepUnit组件");
            return;
        }

        // 3. 初始化状态
        isDelayCoroutineRunning = false;
        currentPhase = PhaseState.Initialize;
        isConditonMet = false;
        AnimationCompleted = false;

        // 4. 开始阶段
        StartPhase();

        // 5. 延迟获取AnimationPlayer组件
        StartCoroutine(DelayedGetAnimationPlayer());

        // Debug.Log($"[{Time.frameCount}] {gameObject.name} PhaseController初始化完成\n" +
        //           $"- ConditionUnit: {(conditionUnit != null ? "已获取" : "未获取")}\n" +
        //           $"- DelayTime: {delayTime}");
    }

    private System.Collections.IEnumerator DelayedGetAnimationPlayer()
    {
        yield return null; // 等待一帧

        // 获取AnimationPlayer组件
        animationPlayer = GetComponent<AnimationPlayer>();
        if (animationPlayer == null)
        {
            Debug.LogWarning($"[{Time.frameCount}] {gameObject.name} 未能获取AnimationPlayer组件，将在需要时重试");
            yield break;
        }

        //Debug.Log($"[{Time.frameCount}] {gameObject.name} 成功获取AnimationPlayer组件");
    }
    #endregion
    
    // 阶段控制
    private bool isDestroyed = false;

    protected virtual void Update()
    {
        if (isDestroyed || !enabled || !gameObject.activeInHierarchy)
        {
            return;
        }

        try
        {
            // 检查必要组件是否存在
            if (conditionUnit == null)
            {
                Debug.LogWarning($"[{Time.frameCount}] {gameObject.name} PhaseController的conditionUnit为空，跳过更新");
                return;
            }

            // 只在Validation阶段检查条件
            if (currentPhase == PhaseState.Validation)
            {
                if (CheckConditions())
                {
                    if (!isConditonMet)
                    {
                        Debug.Log($"[{Time.frameCount}] {gameObject.name} 条件满足，准备切换到动画阶段");
                        isConditonMet = true;
                        MoveToNextPhase(); // 切换到Animation阶段
                    }
                }
                else
                {
                    // 如果条件不满足，保持在Validation阶段
                    return;
                }
            }

            // 动画完成检查
            if (currentPhase == PhaseState.Animation && AnimationCompleted)
            {
                Debug.Log($"[{Time.frameCount}] {gameObject.name} 动画完成，准备切换到完成阶段");
                MoveToNextPhase(); // 切换到Completion阶段
                var stepUnit = GetComponent<StepUnit>();
                if (stepUnit != null)
                {
                    stepUnit.ShowEndObjects();
                }
            }

            // 完成阶段处理
            if (currentPhase == PhaseState.Completion && !isDelayCoroutineRunning)
            {
                var stepUnit = GetComponent<StepUnit>();
                if (stepUnit == null)
                {
                    Debug.LogWarning($"[{Time.frameCount}] {gameObject.name} PhaseController无法获取StepUnit，跳过完成阶段");
                    return;
                }

                if (delayTime > Mathf.Epsilon)
                {
                    StartCoroutine(DelayPhaseCoroutine(delayTime));
                }
                else
                {
                    MoveToNextPhase(); // 切换到Delay阶段
                    CompletePhase();
                    CompleteStepUnit();
                }
            }
        }
        catch (MissingReferenceException e)
        {
            Debug.LogWarning($"[{Time.frameCount}] {gameObject.name} PhaseController检测到组件已被销毁: {e.Message}");
            isDestroyed = true;
            enabled = false;
        }
    }

    protected virtual void OnDestroy()
    {
        isDestroyed = true;
        enabled = false;
    }
    
    public virtual void StartPhase()
    {
        //Debug.Log($"[{Time.frameCount}] {gameObject.name} 开始初始化阶段");
        
        // 1. 设置初始状态
        currentPhase = PhaseState.Initialize;
        isConditonMet = false;
        AnimationCompleted = false;
        
        // 2. 重置条件状态
        if (conditionUnit != null)
        {
            conditionUnit.ResetCondition();
        }
        
        // 3. 重置动画状态
        if (animationPlayer != null)
        {
            animationPlayer.ResetAnimationToFirstFrame();
        }
        
        // 4. 立即切换到验证阶段
        //Debug.Log($"[{Time.frameCount}] {gameObject.name} 初始化完成，准备切换到验证阶段");
        SetPhase(PhaseState.Validation);
    }

    public virtual void MoveToNextPhase()
    {
        Debug.Log($"[{Time.frameCount}] {gameObject.name} 准备切换阶段\n" +
                  $"- 当前阶段: {currentPhase}\n" +
                  $"- 条件是否满足: {isConditonMet}\n" +
                  $"- 动画是否完成: {AnimationCompleted}");
        
        // 检查组件状态
        if (!ValidateComponents())
        {
            Debug.LogError($"[{Time.frameCount}] {gameObject.name} 组件验证失败，无法切换阶段");
            return;
        }

        // 使用协程处理状态转换
        StartCoroutine(HandlePhaseTransition());
    }

    private System.Collections.IEnumerator HandlePhaseTransition()
    {
        switch (currentPhase)
        {
            case PhaseState.Initialize:
                // 等待组件初始化完成
                yield return new WaitUntil(() => conditionUnit != null && animationPlayer != null);
                SetPhase(PhaseState.Validation);
                Debug.Log($"[{Time.frameCount}] {gameObject.name} 进入验证阶段");
                break;

            case PhaseState.Validation:
                // 等待条件满足
                if (!isConditonMet)
                {
                    Debug.Log($"[{Time.frameCount}] {gameObject.name} 等待条件满足");
                    yield return new WaitUntil(() => isConditonMet);
                }
                SetPhase(PhaseState.Animation);
                Debug.Log($"[{Time.frameCount}] {gameObject.name} 进入动画阶段");
                break;

            case PhaseState.Animation:
                // 等待动画完成
                if (!AnimationCompleted)
                {
                    Debug.Log($"[{Time.frameCount}] {gameObject.name} 等待动画完成");
                    yield return new WaitUntil(() => AnimationCompleted);
                }
                SetPhase(PhaseState.Completion);
                Debug.Log($"[{Time.frameCount}] {gameObject.name} 进入完成阶段");
                
                // 如果没有延迟，直接进入延迟阶段
                if (delayTime <= Mathf.Epsilon)
                {
                    yield return null; // 等待一帧确保状态更新
                    SetPhase(PhaseState.Delay);
                    CompletePhase();
                    CompleteStepUnit();
                }
                break;

            case PhaseState.Completion:
                SetPhase(PhaseState.Delay);
                Debug.Log($"[{Time.frameCount}] {gameObject.name} 进入延迟阶段");
                CompletePhase();
                CompleteStepUnit();
                break;

            case PhaseState.Delay:
                // Delay阶段是最终状态，不需要进一步处理
                Debug.Log($"[{Time.frameCount}] {gameObject.name} 已经在延迟阶段，完成所有处理");
                break;
        }
    }

    private bool ValidateComponents()
    {
        bool isValid = true;

        // 检查ConditionUnit
        if (conditionUnit == null)
        {
            Debug.LogError($"[{Time.frameCount}] {gameObject.name} ConditionUnit组件缺失");
            isValid = false;
        }

        // 检查AnimationPlayer，如果缺失则尝试添加
        if (animationPlayer == null)
        {
            Debug.Log($"[{Time.frameCount}] {gameObject.name} 尝试重新获取或添加AnimationPlayer组件");
            animationPlayer = GetComponent<AnimationPlayer>();
            if (animationPlayer == null)
            {
                animationPlayer = gameObject.AddComponent<AnimationPlayer>();
                Debug.Log($"[{Time.frameCount}] {gameObject.name} 已添加新的AnimationPlayer组件");
            }
        }

        return isValid;
    }

    public virtual void SetPhase(PhaseState newPhase)
    {
        currentPhase = newPhase;
        RaisePhaseChangedEvent(currentPhase);
    }

    // 添加受保护的方法用于触发事件
    protected virtual void RaisePhaseChangedEvent(PhaseState phase)
    {
        OnPhaseChanged?.Invoke(phase);
    }

    // 条件检查
    protected virtual bool CheckConditions()
    {       
        if (conditionUnit == null)
        {
            Debug.LogError($"[{Time.frameCount}] {gameObject.name} PhaseController的conditionUnit为空");
            return false;
        }

        // 只在Validation阶段检查条件
        if (currentPhase != PhaseState.Validation)
        {
            return false;
        }

        // 如果已经满足条件，直接返回true
        if (isConditonMet)
        {
            return true;
        }

        // 检查条件状态
        if (conditionUnit.currentState == ConditionState.Satisfied)
        {
            Debug.Log($"[{Time.frameCount}] {gameObject.name} 检测到条件已满足\n" +
                     $"- 条件类型: {conditionUnit.conditionType}");
            
            // 确保动画状态正确
            if (animationPlayer != null)
            {
                animationPlayer.ResetAnimationToFirstFrame();
            }
            
            isConditonMet = true;
            return true;
        }
        
        // 如果条件未满足且不在检查中，触发检查
        if (conditionUnit.currentState != ConditionState.Checking)
        {
            conditionUnit.CheckCondition();
        }

        return false;
    }

    protected virtual bool CheckAnimationPhase()
    {
        if (animationPlayer == null || 
            (animationPlayer.Animations == null || animationPlayer.Animations.Count == 0))
        {
            // 如果没有动画组件或没有动画，直接完成此阶段
            Debug.Log($"[{Time.frameCount}] {gameObject.name} 没有动画需要播放，跳过动画阶段");
            return true;
        }
        
        return animationPlayer.currentState == AnimationState.Completed;
    }

    protected virtual void OnEnable()
    {
        if (conditionUnit != null)
        {
            // 监听条件状态变化
            conditionUnit.OnConditionStateChanged += HandleConditionStateChanged;
        }
    }

    protected virtual void OnDisable()
    {
        if (conditionUnit != null)
        {
            // 取消监听
            conditionUnit.OnConditionStateChanged -= HandleConditionStateChanged;
        }
    }

    private void HandleConditionStateChanged(ConditionState newState)
    {
        // Debug.Log($"[{Time.frameCount}] {gameObject.name} 收到条件状态变化通知\n" +
        //           $"- 新状态: {newState}\n" +
        //           $"- 当前阶段: {currentPhase}\n" +
        //           $"- 条件是否满足: {isConditonMet}");

        // 如果不在验证阶段，先切换到验证阶段
        if (currentPhase != PhaseState.Validation)
        {
            //Debug.Log($"[{Time.frameCount}] {gameObject.name} 当前不在验证阶段，切换到验证阶段");
            SetPhase(PhaseState.Validation);
        }

        // 检查是否是首次满足条件
        if (newState == ConditionState.Satisfied)
        {
            //Debug.Log($"[{Time.frameCount}] {gameObject.name} 检测到条件满足");
            
            // 确保动画状态正确
            if (animationPlayer != null)
            {
                animationPlayer.ResetAnimationToFirstFrame();
                //Debug.Log($"[{Time.frameCount}] {gameObject.name} 重置动画到第一帧");
            }
            
            // 更新状态并触发下一阶段
            isConditonMet = true;
            
            // 确保在下一帧执行状态转换
            StartCoroutine(DelayedMoveToNextPhase());
        }
    }

    private System.Collections.IEnumerator DelayedMoveToNextPhase()
    {
        yield return null; // 等待一帧
        
        Debug.Log($"[{Time.frameCount}] {gameObject.name} 延迟执行状态转换");
        MoveToNextPhase();
    }

    // 根据StepUnit的delay值，协程延迟，协程完成后执行MoveToNextPhase
    protected virtual IEnumerator DelayPhaseCoroutine(float delay)
    {
        if (isDelayCoroutineRunning) yield break;
        isDelayCoroutineRunning = true;

        StepUnit stepUnit = GetComponent<StepUnit>();
        bool isLastStep = stepUnit != null && stepUnit.isLastStep;

        if (currentPhase == PhaseState.Completion)
        {
            if (!isLastStep && !gameObject.activeInHierarchy)
            {
                isDelayCoroutineRunning = false;
                yield break;
            }

            // 先完成所有状态切换
            MoveToNextPhase();
            CompletePhase();
            CompleteStepUnit();

            // 然后等待指定的停留时间
            if (delay > Mathf.Epsilon)
            {
                yield return new WaitForSeconds(delay);
            }
        }

        isDelayCoroutineRunning = false;
    }

    // 完成处理
    protected virtual void CompletePhase()
    {
        OnPhaseCompleted?.Invoke();
    }

    // 进度更新
    public virtual void UpdateProgress(float progress)
    {
        OnProgressUpdated?.Invoke(progress);
    }

    public virtual void SetConditionUnit()
    {
         conditionUnit = GetComponent<ConditionUnit>();
    
        // 添加错误检查
        if (conditionUnit == null)
        {
            Debug.LogError($"PhaseController 需要 ConditionUnit 组件！对象: {gameObject.name}", gameObject);
        }
    } 

    // 切换将StepUnit的currentState改变为StepState.Completed
    public bool isCompletingStep = false;
    
    protected virtual void CompleteStepUnit()
    {
        // 防止重复调用
        if (isCompletingStep) return;
        
        try
        {
            isCompletingStep = true;
            
            StepUnit stepUnit = GetComponent<StepUnit>();
            if (stepUnit == null)
            {
                Debug.LogError($"[{Time.frameCount}] {gameObject.name} 找不到StepUnit组件");
                return;
            }

            Debug.Log($"[{Time.frameCount}] {gameObject.name} 准备完成步骤\n" +
                     $"- 当前阶段: {currentPhase}\n" +
                     $"- 动画状态: {(AnimationCompleted ? "已完成" : "未完成")}");

            // 显示结束物体
            stepUnit.ShowEndObjects();

            // 只在Delay阶段且动画完成时触发步骤完成
            if (currentPhase == PhaseState.Delay && AnimationCompleted)
            {
                Debug.Log($"[{Time.frameCount}] {gameObject.name} 所有条件满足，触发步骤完成");
                stepUnit.OnStepCompleteHandler();
                Debug.Log($"[{Time.frameCount}] {gameObject.name} 步骤完成处理完毕");
            }
        }
        catch (Exception e)
        {
            Debug.LogError($"[{Time.frameCount}] {gameObject.name} 完成步骤时发生错误: {e.Message}\n{e.StackTrace}");
        }
        finally
        {
            isCompletingStep = false;
        }
    }
}

// 阶段状态枚举
public enum PhaseState
{
    Initialize,     // 初始化阶段：对应StepState.Ready
    Validation,     // 验证阶段：对应StepState.Running（条件检查）
    Animation,      // 动画阶段：对应StepState.Running（动画播放）
    Completion,     // 完成阶段：对应StepState.Completed
    Delay          // 延迟阶段：对应StepState.Delay
}
