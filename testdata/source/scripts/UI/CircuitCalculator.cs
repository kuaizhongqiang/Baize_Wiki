
using Complex = System.Numerics.Complex;
using MathNetVector = MathNet.Numerics.LinearAlgebra.Vector<System.Numerics.Complex>;
using MathNetMatrix = MathNet.Numerics.LinearAlgebra.Matrix<System.Numerics.Complex>;

namespace ACConnection.CircuitSystem
{
    public class CircuitCalculator
    {
        /// <summary>
        /// 求解电路
        /// </summary>
        public CircuitSolution Solve(CircuitModel model, bool threePhase)
        {
            var solution = new CircuitSolution();
            
            if (threePhase)
            {
                // 三相电路计算
                solution.PhaseVoltages = new Complex[3][];
                solution.PhaseCurrents = new Complex[3][];
                
                for (int phase = 0; phase < 3; phase++)
                {
                    // 求解每相电路
                    solution.PhaseVoltages[phase] = SolveSinglePhase(model, phase);
                    solution.PhaseCurrents[phase] = CalculatePhaseCurrents(solution.PhaseVoltages[phase], model);
                }
            }
            else
            {
                // 单相电路计算
                solution.Voltages = SolveSinglePhase(model);
                solution.Currents = CalculateCurrents(solution.Voltages, model);
            }

            return solution;
        }

        /// <summary>
        /// 求解单相电路
        /// </summary>
        private Complex[] SolveSinglePhase(CircuitModel model, int phase = 0)
        {
            // 创建导纳矩阵Y和电流向量I
            int nodeCount = model.Nodes.Count;
            var Y = MathNetMatrix.Build.Sparse(nodeCount, nodeCount);
            var I = MathNetVector.Build.Dense(nodeCount);

            // 构建导纳矩阵
            foreach (var branch in model.Branches)
            {
                int from = branch.StartNode;
                int to = branch.EndNode;
                Complex admittance = 1.0 / branch.Impedance;

                // 对角线元素
                Y[from, from] += admittance;
                Y[to, to] += admittance;
                
                // 非对角线元素
                Y[from, to] -= admittance;
                Y[to, from] -= admittance;
            }

            // 构建电流向量(注入电流)
            // 处理电源节点(如有需要)
            // 具体实现取决于CircuitManager中如何定义电源

            // 求解方程 YV = I
            var V = Y.Solve(I);

            return V.ToArray();
        }

        /// <summary>
        /// 计算相电流
        /// </summary>
        private Complex[] CalculatePhaseCurrents(Complex[] voltages, CircuitModel model)
        {
            var currents = new Complex[model.Branches.Count];
            
            for (int i = 0; i < model.Branches.Count; i++)
            {
                var branch = model.Branches[i];
                Complex vFrom = voltages[branch.StartNode];  // 起始节点电压
                Complex vTo = voltages[branch.EndNode];      // 终止节点电压
                Complex deltaV = vFrom - vTo;                // 电压差
                
                currents[i] = deltaV / branch.Impedance;     // 计算电流
            }

            return currents;
        }

        private Complex[] CalculateCurrents(Complex[] voltages, CircuitModel model)
        {
            // Single-phase current calculation
            // Placeholder implementation
            return new Complex[model.Branches.Count];
        }
    }

    /// <summary>
    /// 电路求解结果
    /// </summary>
    public class CircuitSolution
    {
        public Complex[] Voltages { get; set; } // 单相电压 [节点]
        public Complex[] Currents { get; set; } // 单相电流 [支路]
        public Complex[][] PhaseVoltages { get; set; } // 三相电压 [相][节点]
        public Complex[][] PhaseCurrents { get; set; } // 三相电流 [相][支路]
    }
}
