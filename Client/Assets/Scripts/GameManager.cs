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
    enum opcode
    {
        playerStatus = 0,
        playerMove = 1,
    }
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
        nakamaConnection.socket.ReceivedMatchState += (newstate) =>
        {
            Debug.Log("Received match state");
            OnMatchStateReceived.Invoke(newstate);
        };
        OnMatchStateReceived += MatchStatusUpdate;
        player.enabled = true;
        
       
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
                print(content);
                Dictionary<string, string> resultDict = content.FromJson<Dictionary<string, string>>();
                Player.ServerPayload serverPayload;
                serverPayload.isRight = resultDict["IsRight"] == "True";
                serverPayload.pos = new Vector2(float.Parse(resultDict["LocX"]),float.Parse(resultDict["LocY"]));
                serverPayload.lastProcessedInput = int.Parse(resultDict["InputNum"]);
                player.ReceiveNewMsg(serverPayload);
                // Dictionary<string, string> resultDict2 = resultDict["Loc"].FromJson<Dictionary<string, string>>();
                // Vector2 newPos =new Vector2(float.Parse(resultDict2["First"]), float.Parse(resultDict2["Second"]));
                // player.UpdatePlayerStatus(newPos,true);
                break;
        }
    }

    // we shouldn't use right now, enable it when we need it
    // public async Task<IApiRpc> CallRPC(string endpoint, string payload)
    // {
    //     return await nakamaConnection.socket.RpcAsync(endpoint, payload);
    // }
    
    public void SendMessageToServer(int opcode ,string message)
    {
        nakamaConnection.socket.SendMatchStateAsync(nakamaConnection.matchID, opcode,message );
    }
    
    

    
}
