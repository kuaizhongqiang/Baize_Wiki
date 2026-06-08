using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using System.Numerics;

namespace ACConnection.CircuitSystem
{
    [System.Serializable]
    public class CircuitModel
    {
        public List<CircuitNode> Nodes { get; set; }
        public List<ThreePhaseElementBase> Elements { get; set; }
        public List<CircuitBranch> Branches { get; set; } = new List<CircuitBranch>();
    }

    [System.Serializable]
    public class CircuitBranch
    {
        public int StartNode { get; set; }
        public int EndNode { get; set; }
        public Complex Impedance { get; set; }
    }
}

namespace ACConnection.CircuitSystem
{
    [DefaultExecutionOrder(-90)]
    public class CircuitManager : MonoBehaviour
    {
        public static CircuitManager Instance { get; private set; }

        [Header("电路设置")]
        public bool calculateThreePhase = false;
        public float simulationFrequency = 50f;

        [Header("子系统引用")]
        public CircuitCalculator circuitCalculator;
        public CircuitCurrentCalculator currentCalculator;
        public CircuitMatrixCalculator matrixCalculator;
        public CircuitProtectionSystem protectionSystem;

        private List<CircuitNode> nodes = new List<CircuitNode>();
        private List<ThreePhaseElementBase> elements = new List<ThreePhaseElementBase>();
        private List<LineData> lines = new List<LineData>();
        private bool circuitChanged = false;
        private float lastSimulationTime;

        private void Awake()
        {
            if (Instance == null)
            {
                Instance = this;
                DontDestroyOnLoad(this);
            }
            else
            {
                Destroy(this);
                return;
            }

            circuitCalculator = new CircuitCalculator();
            currentCalculator = new CircuitCurrentCalculator();
            matrixCalculator = new CircuitMatrixCalculator();
            protectionSystem = new CircuitProtectionSystem();
        }

        private void Start()
        {
            calculateThreePhase = ACConnectionManager.Instance.ifCulculateThreePhase;
            lastSimulationTime = Time.time;
        }

        private void Update()
        {
            if (circuitChanged || Time.time - lastSimulationTime >= 1f / simulationFrequency)
            {
                RecalculateCircuit();
                circuitChanged = false;
                lastSimulationTime = Time.time;
            }

            ACConnectionDebugger.DebugLog(this, "CircuitManager.Update" + GetCircuitInfo());
        }

        public void RegisterNode(CircuitNode node)
        {
            if (!nodes.Contains(node))
            {
                nodes.Add(node);
                circuitChanged = true;
            }
        }

        public void UnregisterNode(CircuitNode node)
        {
            nodes.Remove(node);
            circuitChanged = true;
        }

        public void RegisterTerminal(ComponentTerminal terminal)
        {
            var node = new CircuitNode {
                Index = nodes.Count,
                TerminalRef = terminal
            };
            terminal.ConnectedNode = node;
            RegisterNode(node);
        }

        private int GetNextNodeIndex()
        {
            return nodes.Count;
        }

        public void RegisterElement(ThreePhaseElementBase element)
        {
            if (!elements.Contains(element))
            {
                elements.Add(element);
                element.ElementDataChanged.AddListener(OnCircuitChanged);
                StartCoroutine(DelayedRecalculate());
            }
        }

        public void UnregisterElement(ThreePhaseElementBase element)
        {
            if (elements.Remove(element))
            {
                RecalculateCircuit();
            }
        }

        private IEnumerator DelayedRecalculate()
        {
            yield return new WaitForEndOfFrame();
            RecalculateCircuit();
        }

        public void RegisterLine(LineData line)
        {
            if (!lines.Contains(line))
            {
                lines.Add(line);
                // 直接标记变化，不依赖LineData事件
                circuitChanged = true;
            }
        }

        private void OnCircuitChanged()
        {
            circuitChanged = true;
        }

        public void RecalculateCircuit()
        {
            // 创建电路模型
            var model = new CircuitModel
            {
                Nodes = nodes,
                Elements = elements
            };

            // 使用CircuitCalculator求解电路
            var solution = circuitCalculator.Solve(model, calculateThreePhase);
            
            // 验证计算结果
            if (solution == null || 
                (!calculateThreePhase && (solution.Voltages == null || solution.Currents == null)) ||
                (calculateThreePhase && (solution.PhaseVoltages == null || solution.PhaseCurrents == null)))
            {
                Debug.LogError("电路计算结果无效");
                return;
            }

            // 更新元件电压电流
            foreach (var element in elements)
            {
                if (calculateThreePhase)
                {
                    // 三相电路更新
                    for (int phase = 0; phase < 3 && phase < solution.PhaseVoltages.Length; phase++)
                    {
                        if (phase < element.ElementData._pointsData.Length)
                        {
                            int nodeIndex = element.ElementData._pointsData[phase].id;
                            if (nodeIndex < solution.PhaseVoltages[phase].Length && 
                                nodeIndex < solution.PhaseCurrents[phase].Length)
                            {
                                element.ElementData._pointsData[phase].pointVoltage =
                                    (float)solution.PhaseVoltages[phase][nodeIndex].Magnitude;
                                element.ElementData._pointsData[phase].pointCurrent =
                                    (float)solution.PhaseCurrents[phase][nodeIndex].Magnitude;
                            }
                        }
                    }
                }
                else
                {
                    // 单相电路更新
                    for (int i = 0; i < element.ElementData._pointsData.Length; i++)
                    {
                        int nodeIndex = element.ElementData._pointsData[i].id;
                        if (nodeIndex < solution.Voltages.Length && 
                            nodeIndex < solution.Currents.Length)
                        {
                            element.ElementData._pointsData[i].pointVoltage =
                                (float)solution.Voltages[nodeIndex].Magnitude;
                            element.ElementData._pointsData[i].pointCurrent =
                                (float)solution.Currents[nodeIndex].Magnitude;
                        }
                        else
                        {
                            Debug.LogWarning($"节点索引越界: {nodeIndex} (Voltages长度: {solution.Voltages.Length}, Currents长度: {solution.Currents.Length})");
                        }
                    }
                }

                // 触发数据变化事件
                element.ElementDataChanged.Invoke();
            }

            // 检查保护条件
            protectionSystem.CheckFaultConditions(elements);
        }

        private string GetCircuitInfo()
        {
            string info = "";
            info += "电路信息：\n";
            info += "节点数：" + nodes.Count + "\n";
            info += "元件数：" + elements.Count + "\n";
            info += "线路数：" + lines.Count + "\n";

            info += "电源信息：\n";
            info += "电源电压：" + ACConnectionManager.Instance.ElementGroupManager.ThreePhasePowerSystem.Voltage + "\n";
            info += "电源节点：" + "\n";
            foreach (var node in ACConnectionManager.Instance.ElementGroupManager.ThreePhasePowerSystem.Points)
            {
                info += "电源节点：" + node.Point.name + "\n";
                info += "电源电压：" + node.pointVoltage.ToString("F2") + "\n";
            }

            info += "节点信息：\n";
            for (int i = 0; i < nodes.Count; i++)
            {
                info += "节点" + i + "：" + nodes[i].Position + "\n";
                info += "节点" + i + "的电压：" + nodes[i].Voltages[0] + " " + nodes[i].Voltages[1] + " " + nodes[i].Voltages[2] + "\n";
            }

            info += "元件信息：\n";
            for (int i = 0; i < elements.Count; i++)
            {
                info += "元件" + i + "：" + elements[i].gameObject.name + "\n";
                info += "元件" + i + "的电流：" + elements[i].ElementData._current + "\n";
                info += "元件" + i + "的电压：" + elements[i].ElementData._voltage + "\n";
                info += "元件" + i + "的电阻：" + elements[i].ElementData._resistance + "\n";
                foreach (var pointData in elements[i].ElementData._pointsData)
                {
                    info += "元件" + i + "的终端" + pointData.Point.name + "的电压：" + pointData.pointVoltage + "\n";
                    info += "元件" + i + "的终端" + "类型1：" +pointData.CircuitType.ToString() + "类型2：" + pointData.TerminalType.ToString() + "\n";
                }
                
            }

            info += "连线信息：\n";
            for (int i = 0; i < lines.Count; i++)
            {
                info += "连线" + i + "：" + lines[i].StartTarget.name + " " + lines[i].EndTarget.name + "\n";
                info += "连线" + i + "的电流：" + lines[i].lineCurrents[0] + "\n";
                info += "连线" + i + "的电压：" + lines[i].lineCurrents[1] + "\n";
            }
            return info;
        }
    }
}
