using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UIElements;
public class Server : MonoBehaviour
{
    public struct Status
    {
        public Vector2 pos;
        public bool isRight ;
    }
    public int tickrate = 5;

    private float tickrateTimer = 0;
    List<OnlineObject> onlineObjects = new List<OnlineObject>();
    List<Status> worldStatus = new List<Status>();
    // Start is called before the first frame update
    void Start()
    {
        onlineObjects.AddRange(FindObjectsOfType<OtherPlayer>());
        // craete a new status for each online object
        foreach (var onlineObject in onlineObjects)
        { 
            //TODO collect all online objects initial status
            worldStatus.Add(new Status());
        }
    }

    // Update is called once per frame
    void Update()
    {
        // every seconds update tickrate times
        tickrateTimer += Time.deltaTime;
        if (tickrateTimer >= 1f / tickrate)
        {
            tickrateTimer -= 1f / tickrate;
            UpdateServer();
        }
    }

    private void UpdateServer()
    {
        float timeDelta = 1f / tickrate;
        for (int i = 0; i < onlineObjects.Count; i++)
        {
            OnlineObject onlineObject = onlineObjects[i];
            // update the world status
            // based on the input 
            // and the previous status
            Status previousStatus = worldStatus[i];
            Status newStatus = previousStatus;
            var x= onlineObject.NewInput.x;
            var y= onlineObject.NewInput.y;
            if (x==1)
            {
                newStatus.isRight = true;
            }
            else if (x==-1)
            {
                newStatus.isRight = false;
            }
            var speedVector = new Vector2(x, y).normalized;
            newStatus.pos = previousStatus.pos + speedVector * timeDelta;
            worldStatus[i] = newStatus;
            // update the online object status
            onlineObject.UpdateStatus(newStatus);
        }
    }
}
