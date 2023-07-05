using System;
using System.Collections;
using System.Collections.Generic;
using System.Reflection.Emit;
using System.Threading.Tasks;
using Nakama;
using Nakama.TinyJson;
using Unity.Mathematics;
using UnityEngine;

public class GameManager : MonoBehaviour
{
    enum opcode
    {
        playerStatus = 0,
        coinsInfo = 1,
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
    public bool gameInitialized;
    public RemotePlayer prefab;
    private Dictionary<string,RemotePlayer> otherPlayers;
    public NakamaConnection nakamaConnection;
    public Player player;
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
        // nakamaConnection.socket.ReceivedMatchState += (newstate) =>
        // {
        //     OnMatchStateReceived.Invoke(newstate);
        // };
        // OnMatchStateReceived += MatchStatusUpdate;
        var mainThread = UnityMainThreadDispatcher.Instance();
        // Setup network event handlers.
        // nakamaConnection.socket.ReceivedMatchmakerMatched += m => mainThread.Enqueue(() => OnReceivedMatchmakerMatched(m));
        // nakamaConnection.socket.ReceivedMatchPresence += m => mainThread.Enqueue(() => OnReceivedMatchPresence(m));
        nakamaConnection.socket.ReceivedMatchState += m => mainThread.Enqueue(() => MatchStatusUpdate(m));
       
    }

    private void Update()
    {
        if (gameInitialized && !player.enabled)
        {
            player.enabled = true;
        }
        // TODO need 2 frame to interpolate right now just stay 
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
        
        var enc = System.Text.Encoding.UTF8;
        var content = enc.GetString(newState.State);
        switch ( newState.OpCode)
        {
            case ((long)0):
               
                // only care about it self
                // TODO check the userID
                
                print(content);
                Dictionary<string, string> resultDict = content.FromJson<Dictionary<string, string>>();
                ServerPacket packet = content.FromJson<ServerPacket>();
                // handle other player
                if (packet.Name != UserId)
                {
                    if (!otherPlayers.ContainsKey(packet.Name))
                    {
                        RemotePlayer newPlayer = Instantiate(prefab,Vector3.one*-1f, quaternion.identity);
                        otherPlayers.Add(packet.Name,newPlayer);
                        newPlayer.prevPos = new Vector2(packet.LocX,packet.LocY);
                        newPlayer.isRight = packet.IsRight;
                    }
                    else
                    {
                        var otherPlayer = otherPlayers[packet.Name];
                        otherPlayer.prevPos = otherPlayer.newPos;
                        otherPlayer.newPos = new Vector2(packet.LocX,packet.LocY);
                        otherPlayer.t = 0;
                        otherPlayer.isRight = packet.IsRight;
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
            case 2:
                print(content);
                break;
                
        }
    }
    
    public void SendMessageToServer(int opcode ,string message)
    {
        nakamaConnection.socket.SendMatchStateAsync(nakamaConnection.matchID, opcode,message );
    }
    
    
    

    
}
