using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.Playables;

public class TimelineCtrl : MonoBehaviour
{
    public PlayableDirector[] playableDirector;
 
    //public GameObject wordCanvas;
  
    int a;
    public void SetOnFirst()
    {
        playableDirector[0].Pause();
        //wordCanvas.SetActive(true);


    }
}
