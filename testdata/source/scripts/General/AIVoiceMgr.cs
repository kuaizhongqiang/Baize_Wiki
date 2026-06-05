using System;
using System.Collections;
using System.Text;
using UnityEngine;
using UnityEngine.Networking;
using Newtonsoft.Json;
using System.IO;
using FrameworkAI_Doubao;

namespace FrameworkSingleton
{
    public class AIVoiceMgr : MonoBehaviour
    {        
        [Header("火山引擎配置")]
        public string appID = "1430566545";                                 // 对应Python的appID
        public string accessKey = "8F9KEYU6tFwLURpPCHB51y4yhtOkhugp";       // 对应Python的accessKey（关键！之前遗漏）
        public string resourceID = "volc.service_type.10029";               // 对应Python的resourceID
        public string appKey = "NSv2_mo-ZR5Xr-FIZmyymLC1JqNWNpWc";          // 对应Python的X-Api-App-Key
        public string speaker = "zh_male_yuanboxiaoshu_moon_bigtts";        // 对应Python的speaker
        [Header("请求签名")]
        public string uid = "TKTestProject";      

        [Header("请求配置")]
        public string requestUrl = "https://openspeech.bytedance.com/api/v3/tts/unidirectional";
        public string textToConvert = "这是一段测试文本，用于测试字节大模型语音合成接口效果。";

        [Header("音频保存配置")]
        // 新增：公开的临时音频保存路径（相对于StreamingAssets目录）
        public string tempAudioSavePath = "Audio/TTS"; // 示例：保存到 StreamingAssets/Audio/TTS 目录
        public string audioFileName = "tts_Generation.mp3"; // 音频文件名

        void Awake()
        {            
            GlobalAIChatMgr.Instance.voiceMgr = this;
        }

        void Start()
        {

        }

        void Update()
        {
        }

        /// <summary>
        /// 开始合成并播放语音
        /// </summary>
        public void StartVoiceSynthesis()
        {
            if (string.IsNullOrEmpty(appID) || string.IsNullOrEmpty(accessKey) || string.IsNullOrEmpty(resourceID))
            {
                Debug.LogError("请填写appID、accessKey、resourceID");
                return;
            }
            StartCoroutine(SendTtsRequest(textToConvert));
        }

        private IEnumerator SendTtsRequest(string text)
        {
            // 1. 构建请求体
            var requestData = new AIVoiceRequest
            {
                user = { uid = uid },
                reqParams = { text = text, speaker = speaker}
            };
            string jsonPayload = JsonConvert.SerializeObject(requestData);

            // 2. 构建请求
            using (var webRequest = new UnityWebRequest(requestUrl, "POST"))
            {
                byte[] bodyRaw = Encoding.UTF8.GetBytes(jsonPayload);
                webRequest.uploadHandler = new UploadHandlerRaw(bodyRaw);
                webRequest.downloadHandler = new DownloadHandlerBuffer();

                // 3. 设置请求头（与Python示例完全一致）
                webRequest.SetRequestHeader("Content-Type", "application/json");
                webRequest.SetRequestHeader("X-Api-App-Id", appID);
                webRequest.SetRequestHeader("X-Api-Access-Key", accessKey); // 关键修正：使用accessKey
                webRequest.SetRequestHeader("X-Api-Resource-Id", resourceID);
                webRequest.SetRequestHeader("X-Api-App-Key", appKey); // 补充遗漏的appKey


                // 4. 发送请求
                yield return webRequest.SendWebRequest();

                // 5. 处理响应
                if (webRequest.result != UnityWebRequest.Result.Success)
                {
                    Debug.LogError($"请求失败: {webRequest.error}，状态码: {webRequest.responseCode}");
                    Debug.LogError($"响应内容: {webRequest.downloadHandler.text}"); // 打印详细错误信息
                    yield break;
                }

                // 6. 解析流式响应（同之前逻辑）
                string responseText = webRequest.downloadHandler.text;
                var responseLines = responseText.Split(new[] { '\n', '\r' }, StringSplitOptions.RemoveEmptyEntries);
                var audioData = new byte[0];

                foreach (var line in responseLines)
                {
                    try
                    {
                        var response = JsonConvert.DeserializeObject<AIVoiceResponse>(line);
                        Debug.Log($"解析行：code={response.code}，data是否存在={!string.IsNullOrEmpty(response.data)}");
                        if (response.code == 0 && !string.IsNullOrEmpty(response.data))
                        {
                            byte[] chunk = Convert.FromBase64String(response.data);
                            Debug.Log($"收到音频片段，长度={chunk.Length}");
                            Array.Resize(ref audioData, audioData.Length + chunk.Length);
                            Array.Copy(chunk, 0, audioData, audioData.Length - chunk.Length, chunk.Length);
                            Debug.Log($"当前总音频长度={audioData.Length}");
                        }
                        else if (response.code == 20000000)
                        {
                            Debug.Log("语音合成完成");
                            Debug.Log($"合成完成时音频总长度={audioData.Length}");
                            break;
                        }
                        else if (response.code > 0)
                        {
                            Debug.LogError($"接口错误: {response.code}，信息: {response.message}");
                            yield break;
                        }
                    }
                    catch (Exception e)
                    {
                        Debug.LogError($"解析响应失败: {e.Message}，行内容: {line}");
                    }
                }

                // 7. 保存或播放音频（根据需要实现）
                if (audioData.Length > 0)
                {
                    // 构建完整路径：StreamingAssets + 自定义子路径 + 文件名
                    string fullSavePath = Path.Combine(
                        Application.streamingAssetsPath,
                        tempAudioSavePath,
                        audioFileName
                    );

                    // 创建目录（如果不存在）
                    string directoryPath = Path.GetDirectoryName(fullSavePath);
                    if (!Directory.Exists(directoryPath))
                    {
                        Directory.CreateDirectory(directoryPath);
                    }

                    // 保存文件
                    File.WriteAllBytes(fullSavePath, audioData);
                    Debug.Log($"音频已保存至: {fullSavePath}");
                }

                GlobalAudioMgr.Instance.OnPlayAudioByType.Invoke(FrameworkProjectSettings.AudioType.TTS, audioFileName);
            }
        }
    }
}