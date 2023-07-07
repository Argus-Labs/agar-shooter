using System;
using System.Collections;
using System.Collections.Generic;
using System.Reflection.Emit;
using System.Threading.Tasks;
using Nakama;
using Nakama.TinyJson;
using Unity.Mathematics;
using UnityEngine;
using UnityEngine.Pool;

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
    private Dictionary<string, RemotePlayer> otherPlayers;
    public NakamaConnection nakamaConnection;
    public Player player;
    public string UserId;
    public Action<IMatchState> OnMatchStateReceived;

    public List<Transform> coins;

    public GameObject coinPrefab;

    public Transform coinsParent;

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
        var mainThread = UnityMainThreadDispatcher.Instance();
        otherPlayers = new Dictionary<string, RemotePlayer>();
        for (int i = 0; i < 200; i++)
        {
            var temp = Instantiate(coinPrefab, coinsParent);
            temp.SetActive(false);
            coins.Add(temp.transform);
        }

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
        switch (newState.OpCode)
        {
            case ((long) 0):

                // only care about it self
                // TODO check the userID
                ServerPacket packet;
                try
                {
                    packet = content.FromJson<ServerPacket>();
                }
                catch (Exception e)
                {
                    print("content: " + content);
                    Console.WriteLine(e);
                    throw;
                }

                // handle other player
                if (packet.Name != UserId)
                {
                    if (!otherPlayers.ContainsKey(packet.Name))
                    {
                        RemotePlayer newPlayer = Instantiate(prefab, Vector3.one * -1f, quaternion.identity);
                        otherPlayers.Add(packet.Name, newPlayer);
                        newPlayer.transform.position = new Vector2(packet.LocX, packet.LocY);
                        newPlayer.prevPos = new Vector2(packet.LocX, packet.LocY);
                        newPlayer.isRight = packet.IsRight;
                    }
                    else
                    {
                        print(content);
                        var otherPlayer = otherPlayers[packet.Name];
                        otherPlayer.prevPos = otherPlayer.newPos;
                        otherPlayer.newPos = new Vector2(packet.LocX, packet.LocY);
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
                serverPayload.pos = new Vector2(packet.LocX, packet.LocY);
                serverPayload.lastProcessedInput = packet.InputNum;
                if (!gameInitialized)
                {
                    gameInitialized = true;
                    player.PlayerInit(serverPayload.pos);
                    break;
                }

                player.UpdateCoins(packet.Coins);
                player.ReceiveNewMsg(serverPayload);
                break;
            case 1:
                Dictionary<string, List<double>> coinsDict = content.FromJson<Dictionary<string, List<double>>>();
                List<double> x_array = coinsDict["First"];
                List<double> y_array = coinsDict["Second"];
                if (x_array.Count != y_array.Count)
                {
                    Debug.LogError("x and y array size not match");
                }

                // check the length of coins and x_array, if there is no enough coins add to the list
                if (coins.Count < x_array.Count)
                {
                    for (int i = coins.Count; i < x_array.Count; i++)
                    {
                        var temp = Instantiate(coinPrefab, coinsParent);
                        temp.SetActive(false);
                        coins.Add(temp.transform);
                    }
                }

                // for [0,x_arry.count] coins set transform to right pos, others set active false
                for (int i = 0; i < coins.Count; i++)
                {
                    if (i < x_array.Count)
                    {
                        coins[i].gameObject.SetActive(true);
                        coins[i].position = new Vector3((float) x_array[i], (float) y_array[i], 0);
                    }
                    else
                    {
                        coins[i].gameObject.SetActive(false);
                    }
                }

                break;
        }
    }

    public void SendMessageToServer(int opcode, string message)
    {
        nakamaConnection.socket.SendMatchStateAsync(nakamaConnection.matchID, opcode, message);
    }
}