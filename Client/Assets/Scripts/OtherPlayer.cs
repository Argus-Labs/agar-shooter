// using System;
// using System.Collections;
// using System.Collections.Generic;
// using UnityEngine;
//
// public class OtherPlayer : MonoBehaviour,OnlineObject
// {
//     [SerializeField] private int health = 100;
//     [SerializeField] private int coin = 0;
//     [SerializeField] private float speed;
//     [SerializeField] private bool isRight = true;
//     [SerializeField] private InputProfile inputProfile;
//     int x;
//     int y;
//     private Vector2 speedVector;
//     private Rigidbody2D rb;
//     // player can use wasd to move up down left right and they can also press two button to move diagonally 
//     // the direction "isRight" of the player is based on the player previous move direction default is right. Filp the player if isLeft
//     private void Awake()
//     {
//         rb = GetComponent<Rigidbody2D>();
//     }
//
//     private void Update()
//     {
//         // int y= Input.GetKey(KeyCode.W)? 1 : Input.GetKey(KeyCode.S)? -1 : 0;
//         // int x= Input.GetKey(KeyCode.D)? 1 : Input.GetKey(KeyCode.A)? -1 : 0;
//          y= Input.GetKey(inputProfile.up)? 1 : Input.GetKey(inputProfile.down)? -1 : 0;
//          x= Input.GetKey(inputProfile.right)? 1 : Input.GetKey(inputProfile.left)? -1 : 0;
//       
//         
//     }
//
//     // private void FixedUpdate()
//     // {
//     //     // // flip the player if isLeft
//     //     // transform.localScale = new Vector3(isRight ? 1 : -1, 1, 1);
//     //     // transform.Translate(speedVector * (speed * Time.fixedDeltaTime));
//     //     // instead of using transform.Translate, we use rb.MovePosition to move the player
//     //     transform.localScale = new Vector3(isRight ? 1 : -1, 1, 1);
//     //     rb.MovePosition(rb.position + speedVector * (speed * Time.fixedDeltaTime));
//     //     Debug.DrawLine(transform.position, transform.position + (Vector3)speedVector, Color.red);
//     //     
//     // }
//
//     #region Online
//
//     // online part 
//     public OnlineObject.Input NewInput
//     {
//         get
//         {
//             OnlineObject.Input ret = new OnlineObject.Input();
//             ret.x = x;
//             ret.y = y;
//             return ret;
//         }
//     }
//
//     public void UpdateStatus(Server.Status newStatus)
//     {
//         transform.localScale = new Vector3(newStatus.isRight ? 1 : -1, 1, 1);
//         transform.position = new Vector3(newStatus.pos.x, newStatus.pos.y, 0);
//         // rb.MovePosition(rb.position + speedVector * (speed * Time.fixedDeltaTime));
//         
//     }
//     #endregion
//
//     
// }