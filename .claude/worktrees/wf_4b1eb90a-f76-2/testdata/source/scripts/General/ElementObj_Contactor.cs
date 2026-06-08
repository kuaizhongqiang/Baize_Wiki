using System.Collections;
using System.Collections.Generic;
using UnityEngine;

namespace FrameworkCircuit
{
    public class ElementObj_Contactor : ElementObjBase
    {
        [Header("元件名称"), SerializeField] string elementName = "接触器";
        [Header("元件类型"), SerializeField] ElectricalElmentType elementType = ElectricalElmentType.Contactor;
        [Header("点数组"), SerializeField] List<ElementObj_Point> elementObj_Points = new List<ElementObj_Point>();
        [Header("是否接通"), SerializeField] bool isConnected = false;

        #region 接口实现
        public override string ElementName { get { return elementName; } set { elementName = value; } }
        public override ElectricalElmentType ElmentType { get { return elementType; } set { elementType = value; } }
        public override List<ElementObj_Point> Points { get { return elementObj_Points; } set { elementObj_Points = value; } }
        public override bool IsConnected {get {return isConnected;} set {isConnected = value;}}
        
        public override void InitializeElement()
        {
            base.InitializeElement();
        }

        #endregion
    }
}