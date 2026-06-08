using System.Collections;
using Sensor.SingletonManagement;
using UnityEngine;

namespace Sensor.Controller.CameraControl
{
    public class CameraBg : MonoBehaviour
    {
        Camera cam;
        float distance = 500f;
        Vector2 screenSize;
        float lastFov;
        float lastAspect;
        float lastLocalZ;
        MeshRenderer meshRenderer;
        Material material;
        string bgName = "";
        Coroutine TexFadeCoroutine;
        const string MatTexture_1 = "_Texture_1";
        const string MatTexture_2 = "_Texture_2";
        const string ChangeTextureBool = "_TexOrColor";         // 这是一个float属性数据 0 = 颜色 1 = 纹理
        const string TexCutFade = "_CutTex";                    // 这是一个float属性数据0-1 平滑切换1 和2 两个纹理
        
        void Awake()
        {
            StartCoroutine(DelayInit());
            screenSize = new Vector2(Screen.width, Screen.height);
            InitMesh();
        }

        IEnumerator DelayInit()
        {
            while(GlobalCameraMgr.Camera == null)
            {
                yield return null;
            }
            cam = GlobalCameraMgr.Camera;
            // 初始化时记录初始值
            if (cam != null)
            {
                lastFov = cam.fieldOfView;
                lastAspect = cam.aspect;
                lastLocalZ = transform.localPosition.z;
                SetScale();
            }
        }

        void Update()
        {
            if (cam == null)
            {
                return;
            }

            // 只在相机参数或位置变化时才更新
            float currentFov = cam.fieldOfView;
            float currentAspect = cam.aspect;
            float currentLocalZ = transform.localPosition.z;

            if (!Mathf.Approximately(currentFov, lastFov) || 
                !Mathf.Approximately(currentAspect, lastAspect) || 
                !Mathf.Approximately(currentLocalZ, lastLocalZ))
            {
                SetScale();
                lastFov = currentFov;
                lastAspect = currentAspect;
                lastLocalZ = currentLocalZ;
            }
        }

        void SetScale()
        {
            float fov = cam.fieldOfView * Mathf.Deg2Rad;
            float aspect = cam.aspect;
            float localZ = transform.localPosition.z;

            float viewHeight = 2f * localZ * Mathf.Tan(fov * 0.5f);
            float viewWidth = viewHeight * aspect;

            transform.localScale = new Vector3(viewWidth, viewHeight, 1f);
        }
    
        void InitMesh()
        {
            meshRenderer = GetComponent<MeshRenderer>();
            material = meshRenderer.sharedMaterial;
        }

        public void SmoothFadeTexture(bool fadeIn)        
        {
            if (material.GetTexture(MatTexture_1) == null || material.GetTexture(MatTexture_2) == null) return;

            material.SetFloat(ChangeTextureBool,1f);
            
            if (TexFadeCoroutine != null)
            {
                StopCoroutine(TexFadeCoroutine);
            }

            TexFadeCoroutine = StartCoroutine(FadeTexture(fadeIn));
        }

        IEnumerator FadeTexture(bool fadeIn)
        {
            float time = 0f;
            float current = material.GetFloat(TexCutFade);
            float duration = 1f;
            float target = fadeIn ? 1f : 0f;

            while (time < duration)
            {
                time += Time.deltaTime;
                current = Mathf.Lerp(current, target, time / duration);
                material.SetFloat(TexCutFade, current);
                yield return null;
            }
            
            material.SetFloat(TexCutFade, target);

            TexFadeCoroutine = null;
        }
    
        public void SetTwoTexture(Texture2D tex1, Texture2D tex2)
        {
            if (tex1 == null || tex2 == null) return;
            
            material.SetTexture(MatTexture_1, tex1);
            material.SetTexture(MatTexture_2, tex2);
        }
    }
}