using System;
using UnityEngine;
using UnityEngine.UI;

public class RemotePlayer : MonoBehaviour
{
    public Vector2 newPos, prevPos;
    public float t = 0f;
    public int serverTickRate = 5;
    public bool isRight = true;
    public int coin = 0;
    public string userID;
    public Transform sprite;
    public Slider healthBar;
    public SpriteRenderer body;
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
        // update player orientation as soon as possible 
        sprite.localScale = new Vector3(isRight ? 1 : -1, 1, 1);
        t += Time.deltaTime/ (1f / serverTickRate);
        // print(t);
        t = Mathf.Clamp(t, 0, 1);
    }
    public void UpdateHealth(int newHealth)
    {
        healthBar.value = newHealth / 100f;
    }

    public void SetColor(Color hsvToRGB)
    {
        print(hsvToRGB);
        body.color = hsvToRGB;
    }
}