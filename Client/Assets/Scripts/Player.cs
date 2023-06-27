using System;
using System.Collections;
using System.Collections.Generic;
using System.Threading.Tasks;
using Nakama.TinyJson;
using UnityEngine;

public class Player : MonoBehaviour
{
    public GameManager gameManager;
    [SerializeField] private int health = 100;
    [SerializeField] private int coin = 0;
    [SerializeField] private float speed;
    [SerializeField] private bool isRight = true;
    [SerializeField] private InputProfile inputProfile;

    public bool IsRight
    {
        get { return isRight; }
        set
        {
            isRight = value;
            transform.localScale = new Vector3(isRight ? 1 : -1, 1, 1);
        }
    }
    // player can use wasd to move up down left right and they can also press two button to move diagonally 
    // the direction "isRight" of the player is based on the player previous move direction default is right. Filp the player if isLeft
    struct PlayerInput
    {
        public string playerID;
        public bool up;
        public bool down;
        public bool left;
        public bool right;

        public PlayerInput(string playerID, bool up, bool down, bool left, bool right)
        {
            this.playerID = playerID;
            this.up = up;
            this.down = down;
            this.left = left;
            this.right = right;
        }
    }



    private void Update()
    {
        
        //
        //
        //
        // string inputJson = input.ToJson();
        // inputJson = @"{
        //     ""Player"": ""175ebc6a-c56b-4ad2-95c1-f9e98c0c8c66"",
        //     ""Up"": true,
        //     ""Down"": false,
        //     ""Left"": false,
        //     ""Right"": false
        // }";
        // print(inputJson);
        // gameManager.CallRPC("games/move", inputJson);
        // var task = gameManager.CallRPC("games/status", playerStatus.ToJson());
        // task.ContinueWith(t =>
        // {
        //     var result = t.Result;
        //     var resultJson = result.Payload;
        //     Debug.Log(resultJson);
        //     Dictionary<string, string> resultDict = resultJson.FromJson<Dictionary<string, string>>();
        //     Dictionary<string, string> resultDict2 = resultDict["Loc"].FromJson<Dictionary<string, string>>();
        //     newPos = new Vector2(float.Parse(resultDict2["First"]), float.Parse(resultDict2["Second"]));
        //     // Debug.Log(newPos);
        // });
        UploadPlayerInput();
    }



    private void UploadPlayerInput()
    {
        PlayerInput input = new PlayerInput(gameManager.UserId, Input.GetKey(inputProfile.up),
            Input.GetKey(inputProfile.down), Input.GetKey(inputProfile.left), Input.GetKey(inputProfile.right));
        int opCode = 1;
        gameManager.SendMessageToServer(opCode,input.ToJson());
    }

    public void UpdatePlayerStatus(Vector2 newPos, bool isRight)
    {
        transform.localPosition = newPos;
        IsRight = isRight;
    }

}