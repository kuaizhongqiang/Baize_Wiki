using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.SceneManagement;
using Valve.VR;

public class PlayMove : MonoBehaviour
{
    public SteamVR_Action_Boolean upper; //上轮盘

    public SteamVR_Action_Boolean lower;  //下轮盘

    Transform play;
    private void Awake()
    {

        play = GameObject.Find("Player").transform;
        upper.onState += Action;
        lower.onState += Action;
    }

    public void Action(SteamVR_Action_Boolean fromAction, SteamVR_Input_Sources fromSource)
    {
        if (SceneManager.GetActiveScene().name.StartsWith("3_4_0"))
        {
            if (fromAction == upper && fromAction.state)
            {
                // 上部动作被按下，向前移动
                //Transform child = transform.Find("Player").transform;
                //Vector3 v = child.forward;

                //transform.Translate(v * 1f * Time.deltaTime);

                if (play != null)
                {
                    play.Translate(Vector3.forward * 1f * Time.deltaTime);
                }
                else
                {
                    play = GameObject.Find("Player").transform;
                }
            }
            if (fromAction == lower && fromAction.state)
            {
                //Transform child = transform.Find("Player").transform;
                //Vector3 v = child.forward;
                //// 下部动作被按下，向后移动
                //transform.Translate(-v * 1f * Time.deltaTime);


                if (play != null)
                {
                    play.Translate(-Vector3.forward * 1f * Time.deltaTime);
                }
                else
                {
                    play = GameObject.Find("Player").transform;
                }
            }
        }
    }

    private void Update()
    {
        if (upper.activeBinding) // 检查动作是否激活  
        {
            //  Debug.Log(222222222222);
            //  Vector2 moveInput = moveAction.GetAxis(SteamVR_Input_Sources.LeftHand);
            //  float forwardBackward = moveInput.y;
            ////  float leftRight = moveInput.x;
            //  if (forwardBackward > 0f)
            //  {
            //      Debug.Log(111111111111111111);
            //      transform.Translate(Vector3.forward * 0.1f * Time.deltaTime);
            //      // 向前移动
            //      // 这里添加您的向前移动逻辑，例如：transform.Translate(Vector3.forward * moveSpeed * Time.deltaTime);
            //  }
            //  else if (forwardBackward < 0f)
            //  {
            //      transform.Translate(-Vector3.forward * 0.1f * Time.deltaTime);
            //      // 向后移动
            //      // 这里添加您的向后移动逻辑，例如：transform.Translate(-Vector3.forward * moveSpeed * Time.deltaTime);
            //  }
        }
    }
}
