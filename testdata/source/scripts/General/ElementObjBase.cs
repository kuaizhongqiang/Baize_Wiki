
using System.Collections.Generic;
using FrameworkInteractiveObject;
using UnityEngine;

namespace FrameworkCircuit
{
    public abstract class ElementObjBase : BaseObj,IElementObj
    {
        [Header("已有元件数量，主要用于改名"),SerializeField] protected int objNumber = 0;
        public virtual string ElementName { get; set; }
        public virtual ElectricalElmentType ElmentType { get; set; }
        public virtual List<ElementObj_Point> Points { get; set; }
        public virtual bool IsConnected { get; set; }
        public int ObjNumber
        {
            get { return objNumber; }
            set { objNumber = value; }
        }
        
        protected override void Awake()
        {
            base.Awake();
            gameObject.name = RenameObj();
            ElementName = RenameObj();
            InitializeElement();
        }
        
        #region 接口实现
        public virtual void InitializeElement()
        {
            if (Points == null)
            {
                Points = new List<ElementObj_Point>();
            }
            else
            {
                Points.Clear();
            }
            
            // 获取所有Point
            for (int i = 0; i < transform.childCount; i++)
            {
                ElementObj_Point pt = transform.GetChild(i).GetComponent<ElementObj_Point>();
                if (pt != null)
                {
                    Points.Add(pt);
                }
            }
        }

        public override void OnMouseEnter()
        {
            
        }
        public override void OnMouseLeave()
        {
            
        }
        public override void OnMouseClick(bool isLeft = true)
        {
            
        }
        public override void OnMouseDoubleClick()
        {
            // 不实现
        }
        public override void MouseDown(bool isLeft)
        {
            // 不实现
        }
        public override void MouseUp(bool isLeft)
        {
            // 不实现
        }
        public override void MouseMove(bool isLeft,Vector2 pos)
        {
            // 不实现
        }
        #endregion

        #region 工具方法
        public  virtual string RenameObj()
        {
            string name = ElementRename.GetElectricalRename(ElmentType);

            if (objNumber < 1)
            {                
                return name;
            }
            else
            {
                return name + objNumber;
            }
        }
        #endregion
    }
}
