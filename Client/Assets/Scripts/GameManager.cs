using System;
using System.Collections.Generic;
using Nakama;
using Nakama.TinyJson;
using TMPro;
using Unity.Mathematics;
using UnityEngine;
using UnityEngine.SceneManagement;
using UnityEngine.Serialization;

public class GameManager : MonoBehaviour
{
    enum opcode
    {
        playerStatus = 0,
        coinsInfo = 1,
        otherPlayerDie = 2,
        attack = 3,
        die = 5,
        addHealth = 6,
        playerName = 9,
        healthPackInfo =10,
        configFile = 14,
        playerMove = 17,
    }

    struct ServerPacket
    {
        public string PersonaTag;
        public int Health;
        public int Coins;
        public float LocX;
        public float LocY;
        public bool IsRight;
        public int InputNum;
        public int Level;

        public ServerPacket(string name, int health, int coins, int locX, int locY, bool isRight, int inputNum)
        {
            PersonaTag = name;
            Health = health;
            Coins = coins;
            LocX = locX;
            LocY = locY;
            IsRight = isRight;
            InputNum = inputNum;
            Level = 0;
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
    struct HealthPack
    {
        public float X;
        public float Y;
        public int Value;
    }

    struct PlayerName
    {
        public string UserId;
        public string Name;
    }

    public bool gameInitialized;
    public RemotePlayer prefab;
    private Dictionary<string, RemotePlayer> otherPlayers;
    public List<NakamaConnection> nakamaConnectionCandidates;
    public NakamaConnection nakamaConnection;
    public Player player;
    public string UserId;
    public Action<IMatchState> OnMatchStateReceived;
    public DynamicPara config;
    public bool configReceived;
    public int serverTick;
    public Transform weaponRangeIndicator;
    [Header("PoolingObjects")]
    #region PoolingObjects

    public List<SpriteRenderer> coins;
    public List<Sprite> coinSprites;
    public SpriteRenderer coinPrefab;
    public Transform coinsParent;
    public DamageTextSpawner dmgTextSpawner;
    public AttackAnimSpawner attackAnimSpawner;
    public List<GameObject> healthPacks;
    public GameObject healthPackPrefab;
    public Transform healthPackParent;

    #endregion
   
    [Header("UI")]

    #region DifferentScreen

    [FormerlySerializedAs("startScreen")]
    public Transform loadingScreen;

    public Transform introScreen;
    public Transform gameOverScreen;
    public TextMeshProUGUI bestRankText;
    public TextMeshProUGUI bestScoreText;
    private int bestRank = 0;
    private int bestScore = 0;

    #endregion

    # region scoreBoard

    public ScoreBoard scoreBoard;
    public float refreshRate = 1f;
    private float refreshTimer = 0f;
    public int scoreboardSize = 5;

    #endregion

    #region Input
    
    public GameObject virtualJoystick;

    #endregion

    #region BestPlayer

    public BestPlayerIndicator bestPlayerIndicator;

    #endregion

    // Start is called before the first frame update
    private void Awake()
    {
        // let the game run 60fps
        Application.targetFrameRate = 60;
        
    }

    void Start()
    {
        introScreen.gameObject.SetActive(true);
        virtualJoystick.SetActive(Utility.WebglIsMobile());
    }

    public async void StartGame()
    {
        loadingScreen.gameObject.SetActive(true);
        await nakamaConnection.Connect();
        UserId = nakamaConnection.session.UserId;
        var mainThread = UnityMainThreadDispatcher.Instance();
        otherPlayers = new Dictionary<string, RemotePlayer>();
        for (int i = 0; i < 200; i++)
        {
            SpriteRenderer temp = Instantiate(coinPrefab, coinsParent);
            temp.gameObject.SetActive(false);
            coins.Add(temp);
            var healthPack = Instantiate(healthPackPrefab, healthPackParent);
            healthPack.SetActive(false);
            healthPacks.Add(healthPack);
        }

        

        // nakamaConnection.socket.ReceivedMatchmakerMatched += m => mainThread.Enqueue(() => OnReceivedMatchmakerMatched(m));
        nakamaConnection.socket.ReceivedMatchPresence += m => mainThread.Enqueue(() => OnReceivedMatchPresence(m));
        nakamaConnection.socket.ReceivedMatchState += m => mainThread.Enqueue(() => MatchStatusUpdate(m));

        bestRank = int.MaxValue;
        bestScore = int.MinValue;
    }

    private void OnReceivedMatchPresence(IMatchPresenceEvent matchPresenceEvent)
    {
        foreach (var join in matchPresenceEvent.Joins)
        {
            print("join:" + join.UserId);
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
        // detect key "o" to call addHealth
        if (Input.GetKeyDown(KeyCode.O))
        {
            AddHealth();
        }

        if (!gameInitialized)
        {
            return;
        }

        // refresh the score board every 1 second
        refreshTimer += Time.deltaTime;
        if (refreshTimer > refreshRate)
        {
            refreshTimer = 0f;
            Dictionary<string, int> players = new Dictionary<string, int>();
            foreach (KeyValuePair<string, RemotePlayer> otherPlayer in otherPlayers)
            {
                players.Add(otherPlayer.Value.nameText.text, otherPlayer.Value.coin);
            }

            players.Add(player.nameText.text, player.Coin);
            int currRank = scoreBoard.Refresh(players, player.nameText.text, scoreboardSize);
            if (currRank < bestRank)
            {
                bestRank = currRank;
            }

            if (player.Coin > bestScore)
            {
                bestScore = player.Coin;
            }

            RemotePlayer bestPlayerPos =  FindTheBestPlayerPos();
            if (bestPlayerPos == null)
            {
                bestPlayerIndicator.enabled = false;
                return;
            }
            bestPlayerIndicator.SetExtractPoint(bestPlayerPos.transform,bestPlayerPos.body);
        }
    }

    private RemotePlayer FindTheBestPlayerPos()
    {
        // based on the level if level tie, use coin to find the best position
        // higher level, higher coin
        RemotePlayer bestPlayer = null;
        int bestLevel = 0;
        int bestCoin = -1;
        
        foreach (KeyValuePair<string, RemotePlayer> otherPlayer in otherPlayers)
        {
            if (otherPlayer.Value.currLevel > bestLevel)
            {
                bestLevel = otherPlayer.Value.currLevel;
                bestPlayer = otherPlayer.Value;
                bestCoin = otherPlayer.Value.coin;
            }
            else if (otherPlayer.Value.currLevel == bestLevel)
            {
                if (otherPlayer.Value.coin > bestCoin)
                {
                    bestCoin = otherPlayer.Value.coin;
                    bestPlayer = otherPlayer.Value;
                }
            }
        }
        //compare with the player itself
        if (player.currLevel > bestLevel)
        {
            return null;
        }
        if (player.currLevel == bestLevel)
        {
            if (player.Coin > bestCoin)
            {
                return null;
            }
        }
        return bestPlayer;
        
       
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
        if (content == "null\n")
        {
            // print($"useless info opcode:{newState.OpCode}");
            return;
        }

        switch (newState.OpCode)
        {
            // add a new case opcode.configFile
            case ((long) opcode.configFile):
                // print("configFile: " + content);
                try
                {
                    config = content.FromJson<DynamicPara>();
                }
                catch (Exception e)
                {
                    print("content: " + content);
                    Console.WriteLine(e);
                    throw;
                }

                if (configReceived)
                {
                    return;
                }
                configReceived = true;
                player.SetSpeed(config.PlayerSpeed);
                player.SetWidthHeight(config.Width, config.Height);
                player.SetCoinCapParameters(config.LevelCoinParameters[0],config.LevelCoinParameters[1],config.LevelCoinParameters[2]);
                player.SetHealthCapParameters(config.LevelHealthParameters[0],config.LevelHealthParameters[1],config.LevelHealthParameters[2]);
                serverTick = config.TickRate;
                weaponRangeIndicator.localScale = new Vector3(config.WeaponRadius/2f, config.WeaponRadius/2, 1f);
                
                
                
                
            break;
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
                if (packet.PersonaTag != UserId && configReceived)
                {
                    // print("content: " + content);
                    if (!otherPlayers.ContainsKey(packet.PersonaTag))
                    {
                        RemotePlayer newPlayer = Instantiate(prefab, Vector3.one * -1f, quaternion.identity);
                        otherPlayers.Add(packet.PersonaTag, newPlayer);
                        newPlayer.transform.position = new Vector2(packet.LocX, packet.LocY);
                        newPlayer.prevPos = new Vector2(packet.LocX, packet.LocY);
                        newPlayer.isRight = packet.IsRight;
                        newPlayer.coin = packet.Coins;
                        newPlayer.currLevel = packet.Level;
                        // newPlayer.SetName(packet.Name);
                        newPlayer.SetColor(Color.HSVToRGB(Mathf.Abs((float) packet.PersonaTag.GetHashCode() / int.MaxValue),
                            0.75f, 0.75f));
                        newPlayer.SetServerTickRate(serverTick);
                        newPlayer.SetHealthCapParameters(config.LevelHealthParameters[0],config.LevelHealthParameters[1],config.LevelHealthParameters[2]);

                        
                    }
                    else
                    {
                        RemotePlayer otherPlayer = otherPlayers[packet.PersonaTag];
                        otherPlayer.prevPos = otherPlayer.newPos;
                        otherPlayer.newPos = new Vector2(packet.LocX, packet.LocY);
                        otherPlayer.t = 0;
                        otherPlayer.isRight = packet.IsRight;
                        otherPlayer.coin = packet.Coins;
                        otherPlayer.UpdateHealth(packet.Health);
                        otherPlayer.CheckUpgrade(packet.Level);
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
                if (!gameInitialized && configReceived)
                {
                    gameInitialized = true;
                    player.PlayerInit(serverPayload.pos);
                    // assign a color based on UserID
                    player.SetColor(
                        Color.HSVToRGB(Mathf.Abs((float) UserId.GetHashCode()) / int.MaxValue, 0.75f, 0.75f));
                    player.enabled = true;
                    loadingScreen.gameObject.SetActive(false);
                    break;
                }

                player.UpdateCoins(packet.Coins);
                player.UpdateHealth(packet.Health);
                player.UpdatePosText($"{packet.LocX},{packet.LocY}");
                player.CheckUpgrade(packet.Level);
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
                        coins[i].transform.position = new Vector3(coinsInfo[i].X, coinsInfo[i].Y, 0);
                        // if the coin value is not 1 set the color to sky blue
                        switch (coinsInfo[i].Value)
                        {
                            case 1:
                                coins[i].sprite = coinSprites[0];
                                break;
                            case 5:
                                coins[i].sprite = coinSprites[1];
                                break;
                            case 10:
                                coins[i].sprite = coinSprites[2];
                                break;
                            default:
                                Debug.LogError($"Invalid coin value{coinsInfo[i].Value}");
                                break;
                        }
                    }
                    else
                    {
                        coins[i].gameObject.SetActive(false);
                    }
                }
                break;
            case (long) opcode.healthPackInfo:
                List<HealthPack> HealthPackInfos;
                try
                {
                    HealthPackInfos = content.FromJson<List<HealthPack>>();
                }
                catch (Exception e)
                {
                    print("content: " + content);
                    Console.WriteLine(e);
                    throw;
                }
                for (int i = 0; i < healthPacks.Count; i++)
                {
                    if (i < HealthPackInfos.Count)
                    {
                        healthPacks[i].gameObject.SetActive(true);
                        healthPacks[i].transform.position = new Vector3(HealthPackInfos[i].X, HealthPackInfos[i].Y, 0);
                    }
                    else
                    {
                        healthPacks[i].gameObject.SetActive(false);
                    }
                }
                break;
            case (long) opcode.attack:
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

                    attackAnimSpawner.Create(origin, target, attackInfo.Damage);
                }

                break;
            case (long) opcode.playerName:
                List<PlayerName> playerNames;
                try
                {
                    playerNames = content.FromJson<List<PlayerName>>();
                }
                catch (Exception e)
                {
                    print("content: " + content);
                    Console.WriteLine(e);
                    throw;
                }
                foreach (PlayerName playerName in playerNames)
                {
                    if (playerName.UserId == UserId)
                    {
                        player.UpdateNameText(playerName.Name);
                    }
                    else
                    {
                        if ( !otherPlayers.ContainsKey(playerName.UserId) )
                        {
                            return;
                        }
                    
                        otherPlayers[playerName.UserId].SetName(playerName.Name);
                    }
                }
                
                break;
            case (long) opcode.die:
                Debug.Log("You die");
#if UNITY_EDITOR
                // UnityEditor.EditorApplication.isPlaying = false;
                gameOverScreen.gameObject.SetActive(true);
                SetFinalScore();
                player.enabled = false;

#endif
                // Application.Quit();
                gameOverScreen.gameObject.SetActive(true);
                SetFinalScore();
                player.enabled = false;


                break;
            case(long) opcode.otherPlayerDie:
                Debug.Log("Other player die");
                if (!otherPlayers.ContainsKey(content))
                {
                    return;
                }
                Destroy(otherPlayers[content].gameObject);
                otherPlayers.Remove(content);
                break;
            default:
                // print("opcode: " + newState.OpCode + " "+ content);
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
        SendMessageToServer((int) opcode.addHealth, "");
    }

    public void Restart()
    {
        SceneManager.LoadScene(SceneManager.GetActiveScene().buildIndex);
    }

    public void SetFinalScore()
    {
        bestRankText.text = bestRank.ToString();
        bestScoreText.text = bestScore.ToString();
    }

    public void OnServerChange(TMP_Dropdown change)
    {
        nakamaConnection= nakamaConnectionCandidates[change.value];
    }
}