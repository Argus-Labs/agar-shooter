using System;
using System.Collections.Generic;
using Nakama;
using Nakama.TinyJson;
using Unity.Mathematics;
using UnityEngine;
using UnityEngine.SceneManagement;

public class GameManager : MonoBehaviour
{
    enum opcode
    {
        playerStatus = 0,
        coinsInfo = 1,
        attack = 3,
        die = 5,
        addHealth = 6,
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

    struct Attack
    {
        public string AttackerID;
        public string DefenderID;
        public int Damage;
    }

    struct Coin
    {
        public float X;
        public float Y;
        public int Value;
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
    public DamageTextSpawner dmgTextSpawner;
    public AttackAnimSpawner attackAnimSpawner;

    #region DifferentScreen
    public Transform startScreen;
    public Transform introScreen;
    public Transform gameOverScreen;
    #endregion

    // Start is called before the first frame update
    private void Awake()
    {
        // let the game run 60fps
        Application.targetFrameRate = 60;
    }

    async void Start()
    {
        startScreen.gameObject.SetActive(true);
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
        nakamaConnection.socket.ReceivedMatchPresence += m => mainThread.Enqueue(() => OnReceivedMatchPresence(m));
        nakamaConnection.socket.ReceivedMatchState += m => mainThread.Enqueue(() => MatchStatusUpdate(m));
    }

    private void OnReceivedMatchPresence(IMatchPresenceEvent matchPresenceEvent)
    {
        foreach (var join in matchPresenceEvent.Joins)
        {
            print("join:"+join.UserId);
        }

        foreach (var leave in matchPresenceEvent.Leaves)
        {
            print("leave" + leave.UserId);
            // check whether the player is in the game
            if (otherPlayers.ContainsKey(leave.UserId))
            {
                Destroy(otherPlayers[leave.UserId].gameObject);
                otherPlayers.Remove(leave.UserId);
            }

            // if the player itself left the game quit the game
        }
    }

    private void Update()
    {
        if (gameInitialized && !player.enabled)
        {
            player.enabled = true;
            startScreen.gameObject.SetActive(false);
            introScreen.gameObject.SetActive(true);
        }
        // detect key "o" to call addHealth
        if (Input.GetKeyDown(KeyCode.O))
        {
            AddHealth();
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
        if (content=="null\n")
        {
            // print($"useless info opcode:{newState.OpCode}");
            return;
        }
        switch (newState.OpCode)
        {
            case ((long) opcode.playerStatus):

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
                    print("content: " + content);
                    if (!otherPlayers.ContainsKey(packet.Name))
                    {
                        RemotePlayer newPlayer = Instantiate(prefab, Vector3.one * -1f, quaternion.identity);
                        otherPlayers.Add(packet.Name, newPlayer);
                        newPlayer.transform.position = new Vector2(packet.LocX, packet.LocY);
                        newPlayer.prevPos = new Vector2(packet.LocX, packet.LocY);
                        newPlayer.isRight = packet.IsRight;
                        newPlayer.SetColor(Color.HSVToRGB(Mathf.Abs((float)packet.Name.GetHashCode() / int.MaxValue), 0.75f, 0.75f));

                    }
                    else
                    {
                   
                        RemotePlayer otherPlayer = otherPlayers[packet.Name];
                        otherPlayer.prevPos = otherPlayer.newPos;
                        otherPlayer.newPos = new Vector2(packet.LocX, packet.LocY);
                        otherPlayer.t = 0;
                        otherPlayer.isRight = packet.IsRight;
                        otherPlayer.UpdateHealth(packet.Health);
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
                    // assign a color based on UserID
                    player.SetColor(Color.HSVToRGB(Mathf.Abs((float) UserId.GetHashCode()) / int.MaxValue, 0.75f, 0.75f));
                    break;
                }

                player.UpdateCoins(packet.Coins);
                player.UpdateHealth(packet.Health);
                player.UpdatePosText($"{packet.LocX},{packet.LocY}");
                player.ReceiveNewMsg(serverPayload);
                break;
            case (long) opcode.coinsInfo:
                // coins info
                List<Coin> coinsInfo;
                try
                {
                    coinsInfo = content.FromJson<List<Coin>>();
                }
                catch (Exception e)
                {
                    print("content: " + content);
                    Console.WriteLine(e);
                    throw;
                }

                // for [0,x_arry.count] coins set transform to right pos, others set active false
                for (int i = 0; i < coins.Count; i++)
                {
                    if (i < coinsInfo.Count)
                    {
                        coins[i].gameObject.SetActive(true);
                        coins[i].position = new Vector3( coinsInfo[i].X, coinsInfo[i].Y, 0);
                        // if the coin value is not 1 set the color to sky blue
                        if (coinsInfo[i].Value != 1)
                        {
                            coins[i].GetComponent<SpriteRenderer>().color = Color.cyan;
                        }
                        else
                        {
                            coins[i].GetComponent<SpriteRenderer>().color = Color.yellow;
                        }
                    }
                    else
                    {
                        coins[i].gameObject.SetActive(false);
                    }
                }

                break;
            case (long)opcode.attack:
                List<Attack> attackInfos;
                try
                {
                    attackInfos = content.FromJson<List<Attack>>();
                }
                catch (Exception e)
                {
                    print("content: " + content);
                    Console.WriteLine(e);
                    throw;
                }
                print(content);
                foreach (var attackInfo in attackInfos)
                {
                    Vector2 origin, target;
                    // origin is attacker transform position
                    if (attackInfo.AttackerID == UserId)
                    {
                        origin = player.transform.position;
                    }
                    else
                    {
                        if (!otherPlayers.ContainsKey(attackInfo.AttackerID))
                        {
                            return;
                        }
                        origin = otherPlayers[attackInfo.AttackerID].transform.position;
                    }
                    
                    // target is defender transform position
                    if (attackInfo.DefenderID == UserId)
                    {
                        target = player.transform.position;
                    }
                    else
                    {
                        if (!otherPlayers.ContainsKey(attackInfo.DefenderID))
                        {
                            return;
                        }
                        target = otherPlayers[attackInfo.DefenderID].transform.position;
                    }
                    attackAnimSpawner.Create(origin,target,attackInfo.Damage);
                }
                break;
            case (long)opcode.die:
                Debug.Log("You die");
#if UNITY_EDITOR
                // UnityEditor.EditorApplication.isPlaying = false;
                gameOverScreen.gameObject.SetActive(true);
#endif
                // Application.Quit();
                gameOverScreen.gameObject.SetActive(true);
                break;
        }
    }

    public void SendMessageToServer(int opcode, string message)
    {
        nakamaConnection.socket.SendMatchStateAsync(nakamaConnection.matchID, opcode, message);
    }

    public void AddHealth()
    {
        print("AddHealth");
        SendMessageToServer((int)opcode.addHealth, "");
    }
    public void Restart()
    {
        SceneManager.LoadScene(SceneManager.GetActiveScene().buildIndex);
    }
}