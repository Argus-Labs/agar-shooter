/* 
    ------------------- Code Monkey -------------------

    Thank you for downloading this package
    I hope you find it useful in your projects
    If you have any questions let me know
    Cheers!

               unitycodemonkey.com
    --------------------------------------------------
 */

using System;
using UnityEngine;
using TMPro;
using UnityEngine.UIElements;

public class DamagePopup : MonoBehaviour {
    


    private const float DISAPPEAR_TIMER_MAX = 1f;

    private TextMeshPro textMesh;
    private float disappearTimer;
    private Color textColor;
    private Vector3 moveVector;
    public float _moveYspeed;
    private Action<DamagePopup> killAction;

    public void Init(Action<DamagePopup> killAction)
    {
        this.killAction = killAction;
    }

    private void Awake() {
        textMesh = transform.GetComponent<TextMeshPro>();
    }

    public void Setup(int damageAmount,int sortingOrder,Vector2 position)
    {
        transform.localPosition = position;
        textMesh.SetText(damageAmount.ToString());
        textMesh.fontSize = 4;
        if (damageAmount<0)
        {
            textMesh.color = Color.yellow; 
            moveVector = new Vector3(-0.7f, 1) * 1f;
        }
        else
        {
            textMesh.color = Color.red;
            moveVector = new Vector3(.7f, 1) * 1f;
        }
        disappearTimer = DISAPPEAR_TIMER_MAX;
        textMesh.sortingOrder = sortingOrder;
        transform.localScale = Vector3.one;
        textColor.a = 1f;
    }

    private void Update() {
        transform.position += moveVector * Time.deltaTime;
        _moveYspeed = 0.1f;
        moveVector -= moveVector * (_moveYspeed * Time.deltaTime);

        if (disappearTimer > DISAPPEAR_TIMER_MAX * .5f) {
            // First half of the popup lifetime
            float increaseScaleAmount = 1f;
            transform.localScale += Vector3.one * (increaseScaleAmount * Time.deltaTime);
        } else {
            // Second half of the popup lifetime
            float decreaseScaleAmount = 1f;
            transform.localScale -= Vector3.one * (decreaseScaleAmount * Time.deltaTime);
        }

        disappearTimer -= Time.deltaTime;
        if (disappearTimer < 0) {
            // Start disappearing
            float disappearSpeed = 3f;
            textColor.a -= disappearSpeed * Time.deltaTime;
            textMesh.color = textColor;
            if (textColor.a < 0) {
                Release();
            }
        }
    }
    private void Release()
    {
        killAction?.Invoke(this);
    }

  

}
