using System;
using System.IO;
using UnityEngine;

namespace FrameworkAI_Doubao
{
    public static class AIChatImageToString
    {
        /// <summary>
        /// 将Texture2D转换为Base64字符串（先等比压缩至最大1024宽/高）
        /// </summary>
        /// <param name="sourceTexture">原始纹理</param>
        /// <returns>压缩后图像的Base64字符串</returns>
        public static string ConvertTextureToBase64(Texture2D sourceTexture)
        {
            if (sourceTexture == null)
            {
                Debug.LogError("ConvertTextureToBase64 failed: 源纹理为空");
                return string.Empty;
            }

            // 步骤1：计算等比压缩后的尺寸（最大边不超过1024）
            int targetWidth = sourceTexture.width;
            int targetHeight = sourceTexture.height;
            float maxSize = 1024f;

            // 判断是否需要压缩
            if (targetWidth > maxSize || targetHeight > maxSize)
            {
                // 计算缩放比例（取最小缩放比例保证最大边不超过1024）
                float scale = Mathf.Min(maxSize / targetWidth, maxSize / targetHeight);
                targetWidth = Mathf.RoundToInt(targetWidth * scale);
                targetHeight = Mathf.RoundToInt(targetHeight * scale);
            }

            // 步骤2：创建压缩后的纹理
            Texture2D compressedTexture = null;
            RenderTexture renderTexture = null;
            try
            {
                // 创建临时渲染纹理用于缩放
                renderTexture = RenderTexture.GetTemporary(
                    targetWidth,
                    targetHeight,
                    0,
                    RenderTextureFormat.Default,
                    RenderTextureReadWrite.Linear
                );

                // 保存当前渲染目标
                RenderTexture previous = RenderTexture.active;

                // 将源纹理绘制到缩放后的渲染纹理
                Graphics.Blit(sourceTexture, renderTexture);
                RenderTexture.active = renderTexture;

                // 从渲染纹理读取像素到新纹理
                compressedTexture = new Texture2D(targetWidth, targetHeight, TextureFormat.RGB24, false);
                compressedTexture.ReadPixels(new Rect(0, 0, targetWidth, targetHeight), 0, 0);
                compressedTexture.Apply();

                // 恢复渲染目标
                RenderTexture.active = previous;

                // 确保StreamingAssets目录存在
                if (!Directory.Exists(Application.streamingAssetsPath))
                {
                    Directory.CreateDirectory(Application.streamingAssetsPath);
                }

                // 保存图片到StreamingAssets路径下，使用唯一文件名
                string fileName = "compressed.png";
                string path = Path.Combine(Application.streamingAssetsPath, fileName);
                File.WriteAllBytes(path, compressedTexture.EncodeToPNG());
                Debug.Log($"图片已保存到: {path}"); 
            }
            catch (Exception e)
            {
                Debug.LogError($"图像压缩失败: {e.Message}");
                return string.Empty;
            }
            finally
            {
                // 释放临时渲染纹理
                if (renderTexture != null)
                    RenderTexture.ReleaseTemporary(renderTexture);
            }

            // 步骤3：转换为Base64
            byte[] imageBytes = compressedTexture.EncodeToPNG(); 
            string base64String = Convert.ToBase64String(imageBytes);

            // 释放资源
            UnityEngine.Object.DestroyImmediate(compressedTexture);

            return base64String;
        }
        /// <summary>
        /// 将Base64字符串转换回Texture2D
        /// </summary>
        /// <param name="base64String">Base64编码字符串</param>
        /// <returns>转换后的纹理</returns>
        public static Texture2D ConvertBase64ToTexture(string base64String)
        {
            if (string.IsNullOrEmpty(base64String))
            {
                Debug.LogError("Base64字符串为空");
                return null;
            }

            try
            {
                byte[] imageBytes = System.Convert.FromBase64String(base64String);
                Texture2D texture = new Texture2D(2, 2);
                if (texture.LoadImage(imageBytes))
                {
                    return texture;
                }
                else
                {
                    Debug.LogError("加载图片数据失败");
                    return null;
                }
            }
            catch (System.Exception e)
            {
                Debug.LogError($"Base64转换失败: {e.Message}");
                return null;
            }
        }

        /// <summary>
        /// 图片格式枚举
        /// </summary>
        public enum ImageFormat
        {
            PNG,
            JPG
        }
    }
}