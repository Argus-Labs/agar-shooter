using System;
using UnityEngine;

public class RemotePlayer : MonoBehaviour
{
    public Vector2 newPos, prevPos;
    public float t = 0f;
    public int serverTickRate = 5;
    // lerp between prevPos and newPos

    private void Start()
    {
        newPos = new Vector2(-1, -1);
    }

    private void Update()
    {
        float TOLERANCE=0.1f;
        if (Math.Abs(newPos.x - (-1f)) < TOLERANCE)
        {
            // invalid newpos 
            return;
        }
        transform.position = Vector2.Lerp(prevPos, newPos, t);
        t += Time.deltaTime/ (1f / serverTickRate);
        print(t);
        t = Mathf.Clamp(t, 0, 1);
    }
        
}