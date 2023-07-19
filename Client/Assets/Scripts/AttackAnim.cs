using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class AttackAnim : MonoBehaviour
{
    private Action<AttackAnim> killAction;
    public float speed;
    private Vector3 moveVector;
    private float t = 0;
    public Vector2 origin, target;
    public int damage;

    public void Init(Action<AttackAnim> killAction)
    {
        this.killAction = killAction;
    }
    public void Setup(Vector2 origin,Vector2 target,int damage)
    {
        this.origin = origin;
        this.target = target;
        this.damage = damage;
        transform.position = origin;
        moveVector = target - origin;
        moveVector.Normalize();
        transform.up =  moveVector;
        t = 0;
    }

    private void Update()
    {
        // move the attack anim towards the target but not overshoot and call Release() when it reaches the target
        
        t+=Time.deltaTime*speed;
        transform.position = Vector2.Lerp(origin, target, t);
        if (t>1)
        {
            Release();
        }
        
        
    }

    private void Release()
    {
        killAction?.Invoke(this);

    }

}