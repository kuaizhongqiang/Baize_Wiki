
using System.Collections.Generic;
using Complex = System.Numerics.Complex;

namespace ACConnection.CircuitSystem
{
    /// <summary>
    /// 电路电流计算器
    /// </summary>
    public class CircuitCurrentCalculator
    {
        /// <summary>
        /// 计算所有元件电流
        /// </summary>
        public void CalculateCurrents(CircuitSolution solution, List<ThreePhaseElementBase> elements)
        {
            if (solution.PhaseVoltages != null)
            {
                // 三相电流计算
                for (int phase = 0; phase < 3; phase++)
                {
                    CalculatePhaseCurrents(solution.PhaseVoltages[phase], solution.PhaseCurrents[phase], elements, phase);
                }
            }
            else
            {
                // 单相电流计算
                CalculateSinglePhaseCurrents(solution.Voltages, solution.Currents, elements);
            }
        }

        private void CalculatePhaseCurrents(Complex[] voltages, Complex[] currents, 
            List<ThreePhaseElementBase> elements, int phase)
        {
            for (int i = 0; i < elements.Count; i++)
            {
                var element = elements[i];
                if (element.ElementData._pointsData == null || element.ElementData._pointsData.Length <= phase)
                    continue;

                var terminal = element.ElementData._pointsData[phase];
                if (terminal.isConnected)
                {
                    // 目前使用欧姆定律简单计算
                    // 更复杂的计算需要考虑互感耦合等因素
                    var voltage = voltages[terminal.id];  // 节点电压
                    var impedance = new Complex(element.ElementData._resistance, 0);  // 元件阻抗
                    var current = voltage / impedance;  // 计算电流
                    
                    // 更新元件数据
                    terminal.pointCurrent = (float)current.Magnitude;
                    currents[i] = current;
                }
            }
        }

        private void CalculateSinglePhaseCurrents(Complex[] voltages, Complex[] currents, 
            List<ThreePhaseElementBase> elements)
        {
            for (int i = 0; i < elements.Count; i++)
            {
                var element = elements[i];
                if (element.ElementData._pointsData == null || element.ElementData._pointsData.Length == 0)
                    continue;

                var terminal = element.ElementData._pointsData[0]; // 第一个端子
                if (terminal.isConnected)
                {
                    var voltage = voltages[terminal.id];  // 节点电压
                    var impedance = new Complex(element.ElementData._resistance, 0);  // 元件阻抗
                    var current = voltage / impedance;  // 计算电流
                    
                    // 更新元件数据
                    terminal.pointCurrent = (float)current.Magnitude;
                    element.ElementData._current = (float)current.Magnitude;
                    currents[i] = current;
                }
            }
        }
    }
}
