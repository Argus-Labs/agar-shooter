using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class Player : MonoBehaviour
{
    [SerializeField] private int health = 100;
    [SerializeField] private int coin = 0;
    [SerializeField] private float speed;
    [SerializeField] private bool isRight = true;
    [SerializeField] private InputProfile inputProfile;
    // player can use wasd to move up down left right and they can also press two button to move diagonally 
    // the direction "isRight" of the player is based on the player previous move direction default is right. Filp the player if isLeft
    private void Update()
    {
        // int y= Input.GetKey(KeyCode.W)? 1 : Input.GetKey(KeyCode.S)? -1 : 0;
        // int x= Input.GetKey(KeyCode.D)? 1 : Input.GetKey(KeyCode.A)? -1 : 0;
        int y= Input.GetKey(inputProfile.up)? 1 : Input.GetKey(inputProfile.down)? -1 : 0;
        int x= Input.GetKey(inputProfile.right)? 1 : Input.GetKey(inputProfile.left)? -1 : 0;
        Vector2 speedVector = new Vector2(x, y).normalized;
        if (x==1)
        {
            isRight = true;
        }
        else if (x==-1)
        {
            isRight = false;
        }
        // flip the player if isLeft
        transform.localScale = new Vector3(isRight ? 1 : -1, 1, 1);
        transform.Translate(speedVector * (speed * Time.deltaTime));
    }
}
