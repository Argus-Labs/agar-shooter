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
    private int sequenceNumber = 0;
    CircularArray<PlayerInputExtraInfo> pendingInputs = new CircularArray<PlayerInputExtraInfo>();

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
        // the variable name first letter must be capital because in Golang, public variable must start with capital letter
        public string PlayerID;
        public bool Up;
        public bool Down;
        public bool Left;
        public bool Right;
        public int Input_sequence_number;

        public PlayerInput(string playerID, bool up, bool down, bool left, bool right, int input_sequence_number)
        {
            this.PlayerID = playerID;
            this.Up = up;
            this.Down = down;
            this.Left = left;
            this.Right = right;
            this.Input_sequence_number = input_sequence_number;
        }
    }

    struct PlayerInputExtraInfo
    {
        public PlayerInput input;
        public float deltaTime;
        public Vector2 position;
        public PlayerInputExtraInfo(PlayerInput input, float deltaTime, Vector2 position)
        {
            this.input = input;
            this.deltaTime = deltaTime;
            this.position = position;
        }
    }

    public struct ServerPayload
    {
        public int lastProcessedInput;
        public Vector2 pos;
        public bool isRight;
    }



    private void  Update()
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
        sequenceNumber++;
    }



    private void UploadPlayerInput()
    {
        PlayerInput input = new PlayerInput(gameManager.UserId, Input.GetKey(inputProfile.up),
            Input.GetKey(inputProfile.down), Input.GetKey(inputProfile.left), Input.GetKey(inputProfile.right),sequenceNumber);
        int opCode = 1;
        gameManager.SendMessageToServer(opCode,input.ToJson());
        ApplyInput(input);
        pendingInputs.Enqueue(new PlayerInputExtraInfo(input, Time.deltaTime, transform.localPosition));

    }

    private void ApplyInput(PlayerInput input)
    {
        // based on the input update player position and direction
        int y= input.Up? 1 : input.Down? -1 : 0;
        int x= input.Right? 1 : input.Left? -1 : 0;
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

    public void ReceiveNewMsg(ServerPayload payload)
    {
        // delete all the inputs that have been processed by the server
        var j = 0;
        while (j < pendingInputs.Count) {
            var input = pendingInputs[j];
            if (input.input.Input_sequence_number < payload.lastProcessedInput) {
                // Already processed. Its effect is already taken into account into the world update
                // we just got, so we can drop it.
                // pendingInputs.splice(j, 1);
                pendingInputs.Dequeue();
            }
            else if (input.input.Input_sequence_number == payload.lastProcessedInput)
            {
                if (Vector2.Distance(payload.pos,input.position)<0.05f)
                {
                    // Already processed. Its effect is already taken into account into the world update
                    // we just got, so we can drop it.
                    pendingInputs.Dequeue();
                    break;
                }
                pendingInputs.Dequeue();
                Debug.Log("there is a difference");
                
            }
            else {
                // Not processed by the server yet. Re-apply it.
                
                
                ApplyInput(input.input);
                // update the player position in the pendingInputs
                pendingInputs[j] = new PlayerInputExtraInfo(input.input, input.deltaTime, transform.localPosition);
                j++;
            }
        }
        
    }
    public void UpdatePlayerStatus(Vector2 newPos, bool isRight)
    {
        transform.localPosition = newPos;
        IsRight = isRight;
    }

}