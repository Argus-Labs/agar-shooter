using System;
using System.Collections;
using System.Collections.Generic;
using System.Reflection.Emit;
using System.Threading.Tasks;
using Nakama;
using Nakama.TinyJson;
using UnityEngine;

public class GameManager : MonoBehaviour
{
    public bool gameInitialized,gameInitialized2;
    enum opcode
    {
        playerStatus = 0,
        playerMove = 17,
    }
    struct ServerPacket
    {
        public string Name;
        public int Health;
        public int Coins;
        public float LocX;
        public float LocY;
        public bool IsRight;
        public int InputNum;
        public ServerPacket(string name, int health, int coins, int locX, int locY, bool isRight, int inputNum)
        {
            Name = name;
            Health = health;
            Coins = coins;
            LocX = locX;
            LocY = locY;
            IsRight = isRight;
            InputNum = inputNum;
        }
    }
    
    public NakamaConnection nakamaConnection;
    public Player player;
    public RemotePlayer player2;
    public string UserId;
    public Action<IMatchState> OnMatchStateReceived;
    // Start is called before the first frame update
    private void Awake()
    {
        // let the game run 60fps
        Application.targetFrameRate = 60;
        
    }

    async void Start()
    {
        await nakamaConnection.Connect();
        UserId = nakamaConnection.session.UserId;
        nakamaConnection.socket.ReceivedMatchState += (newstate) =>
        {
            OnMatchStateReceived.Invoke(newstate);
        };
        OnMatchStateReceived += MatchStatusUpdate;
        
       
    }

    private void Update()
    {
        if (gameInitialized && !player.enabled)
        {
            player.enabled = true;
        }
        // TODO need 2 frame to interpolate right now just stay 
        if (gameInitialized2 && !player2.enabled)
        {
            player2.transform.position = player2.prevPos;
            player2.enabled = true;
        }
    }

    private void OnApplicationQuit()
    {
        // quit the match
        nakamaConnection.socket.LeaveMatchAsync(nakamaConnection.matchID);
        OnMatchStateReceived -= MatchStatusUpdate;
    }
    /*
     * type PlayerComponent struct {
	Name string// username; ip for now
	Health int// current player health (cap enforced in update loop)
	Coins int// how much money the player has
	Weapon Weapon// current player weapon; default is 0 for Melee
	Loc Pair[int, int]// current location
	Dir Direction// direction player faces & direction player moves; currently, both are the same
}
     */

    private void MatchStatusUpdate(IMatchState newState)
    {
        
        switch ( newState.OpCode)
        {
            case ((long)0):
               
                // only care about it self
                // TODO check the userID
                
                var enc = System.Text.Encoding.UTF8;
                var content = enc.GetString(newState.State);
                // print(content);
                Dictionary<string, string> resultDict = content.FromJson<Dictionary<string, string>>();
                ServerPacket packet = content.FromJson<ServerPacket>();
                // handle other player
                if (packet.Name != UserId)
                {
                    if (!gameInitialized2)
                    {
                        player2.prevPos = new Vector2(packet.LocX,packet.LocY);
                        gameInitialized2 = true;
                    }
                    else
                    {
                        player2.prevPos = player2.newPos;
                        player2.newPos = new Vector2(packet.LocX,packet.LocY);
                        player2.t = 0;
                    }
                    
                    break;
                }
                Player.ServerPayload serverPayload;
                // serverPayload.isRight = resultDict["IsRight"] == "True";
                // serverPayload.pos = new Vector2(float.Parse(resultDict["LocX"]),float.Parse(resultDict["LocY"]));
                // serverPayload.lastProcessedInput = int.Parse(resultDict["InputNum"]);
                serverPayload.isRight = packet.IsRight;
                serverPayload.pos = new Vector2(packet.LocX,packet.LocY);
                serverPayload.lastProcessedInput = packet.InputNum;
                if (!gameInitialized)
                {
                    gameInitialized = true;
                    player.PlayerInit(serverPayload.pos);
                    break;
                }
                player.ReceiveNewMsg(serverPayload);
                break;
        }
    }
    
    public void SendMessageToServer(int opcode ,string message)
    {
        nakamaConnection.socket.SendMatchStateAsync(nakamaConnection.matchID, opcode,message );
    }
    
    

    
}
