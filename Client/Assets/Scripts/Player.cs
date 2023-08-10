using System;
using Nakama.TinyJson;
using TMPro;
using Unity.VisualScripting;
using UnityEngine;
using UnityEngine.InputSystem;
using UnityEngine.InputSystem.Controls;
using UnityEngine.UI;

public class Player : MonoBehaviour
{
    public GameManager gameManager;
    [SerializeField] private int health = 100;
    [SerializeField] private int coin = 0;
    [SerializeField] private float speed;
    [SerializeField] private bool isRight = true;
    [SerializeField] private TextMeshProUGUI coinText;
    [SerializeField] private Slider healthBar;
    [SerializeField] private Transform sprite;
    [SerializeField] private SpriteRenderer body;
    [SerializeField] TextMeshProUGUI posText;
    [SerializeField] private TextMeshProUGUI levelText;
    [SerializeField] private Animator animator;
    private int sequenceNumber = 0;
    CircularArray<PlayerInputExtraInfo> pendingInputs = new CircularArray<PlayerInputExtraInfo>(100);
    public Vector2 pos = new Vector2(0, 0);
    public TextMeshProUGUI nameText;
    public PlayerAction playerAction;
    public int currLevel = 1;
    public int width, height;
    // may need introduce other parameters
    public void PlayerInit(Vector2 pos)
    {
        this.pos = pos;
    }

    public bool IsRight
    {
        get { return isRight; }
        set
        {
            isRight = value;
            sprite.localScale = new Vector3(isRight ? 1 : -1, 1, 1);
        }
    }

    public float Speed
    {
        // get { return speed/* Mathf.Exp(-0.01f*coin)*/; }
        get { return speed* Mathf.Exp(-0.01f*coin); }
    }

    public int Coin
    {
        get { return coin; }
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
        public float Delta;

        public PlayerInput(string playerID, bool up, bool down, bool left, bool right, int input_sequence_number,float delta)
        {
            this.PlayerID = playerID;
            this.Up = up;
            this.Down = down;
            this.Left = left;
            this.Right = right;
            this.Input_sequence_number = input_sequence_number;
            this.Delta = delta;
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


    private void Awake()
    {
        playerAction = new PlayerAction();
        currLevel = 1;
    }

    private void OnEnable()
    {
        playerAction.Enable();
    }
    
    private void OnDisable()
    {
        playerAction.Disable();
    }

    public const float cooldown = 1f/60f;
    private float lastInputTime = 0f;
    
    private void Update()                       
    {
        
        // only update when cooldown is over
        if (Time.time - lastInputTime < cooldown)
        {
            return;
        }
        lastInputTime = Time.time;
        // print(sequenceNumber);
        UploadPlayerInput();
        sequenceNumber++;
        // update the player position on main thread
        transform.localPosition = pos;
        IsRight = isRight;

    }


    private void UploadPlayerInput()
    {
        Vector2to4button(playerAction.Player.Movement.ReadValue<Vector2>(), out bool up, out bool down, out bool left, out bool right);
        // PlayerInput input = new PlayerInput(gameManager.UserId, Input.GetKey(inputProfile.up),
        //     Input.GetKey(inputProfile.down), Input.GetKey(inputProfile.left), Input.GetKey(inputProfile.right),
        //     sequenceNumber,Time.deltaTime);
        PlayerInput input = new PlayerInput(gameManager.UserId, up, down, left, right,
            sequenceNumber,Time.deltaTime);
        
        int opCode = 17;;
        gameManager.SendMessageToServer(opCode, input.ToJson()); 
        ApplyInput(input, Time.deltaTime);
        // print player pos (x,y) isright and also sequence number
        //print("Client Sim: Pos: (" + pos.x + "," + pos.y + ") isRight: " + isRight + " sequenceNumber: " + sequenceNumber);
        
        pendingInputs.Enqueue(new PlayerInputExtraInfo(input, Time.deltaTime, pos));
    }
    private void Vector2to4button(Vector2 input, out bool up, out bool down, out bool left, out bool right)
    {
        up = input.y > 0;
        down = input.y < 0;
        left = input.x < 0;
        right = input.x > 0;
    }
    private void OnMovement(InputAction action)
    {
    }
    
    // change above code to C#
    private int diff(bool a, bool b)
    {
        if (a == b)
        {
            return 0;
        }

        if ( a && !b)
        {
            return 1;
        }

        return -1;
    }
    private void ApplyInput(PlayerInput input, float deltaTime)
    {
        // based on the input update player position and direction
        int y = diff(input.Up , input.Down );
        int x = diff(input.Right,input.Left);
        Vector2 speedVector = new Vector2(x, y).normalized;
        if (x == 1)
        {
            isRight = true;
        }
        else if (x == -1)
        {
            isRight = false;
        }

        //2m/s
        pos += speedVector * (Speed * deltaTime);
        // check the boundary map is from left bottom(0,0) to up right(100,100)
        pos.x = Mathf.Clamp(pos.x, 0, width);
        pos.y = Mathf.Clamp(pos.y, 0, height);
    }

    public void ReceiveNewMsg(ServerPayload payload)
    {
        // delete all the inputs that have been processed by the server
        var j = 0;
        while (j < pendingInputs.Count)
        {
            var input = pendingInputs[j];
            if (input.input.Input_sequence_number < payload.lastProcessedInput)
            {
                // Already processed. Its effect is already taken into account into the world update
                // we just got, so we can drop it.
                // pendingInputs.splice(j, 1);
                pendingInputs.Dequeue();
            }
            else if (input.input.Input_sequence_number == payload.lastProcessedInput)
            {
                if (Vector2.Distance(payload.pos, input.position) < 0.05f)
                {
                    // Already processed. Its effect is already taken into account into the world update
                    // we just got, so we can drop it.
                    pendingInputs.Dequeue();
                    break;
                }

                pendingInputs.Dequeue();
                // print expected pos and actual pos and sequence number
                Debug.Log("there is a difference:" + "Server pos:" + payload.pos + "player pos:" + input.position +
                          "sequence number:" + payload.lastProcessedInput + "current sequence number:" +
                          sequenceNumber);
                pos = payload.pos;
            }
            else
            {
                // Not processed by the server yet. Re-apply it.
                ApplyInput(input.input, input.deltaTime);
                // update the player position in the pendingInputs
                pendingInputs[j] = new PlayerInputExtraInfo(input.input, input.deltaTime, pos);
                j++;
            }
        }
    }

    public void UpdateHealth(int newHealth)
    {
        health = newHealth;
        healthBar.value = health / HealthCap(currLevel);
    }

    public void UpdateCoins(int newCoinCount)
    {
        coin = newCoinCount;
        coinText.text = $"{coin}/{CoinCap(currLevel)}";
        // update UI
    }

    public void SetColor(Color color)
    {
        body.color = color;
    }
    public void SetSpeed(float speed)
    {
        this.speed = speed;
    }
    public void SetWidthHeight(int width, int height)
    {
        this.width = width;
        this.height = height;
    }

    public void UpdatePosText(string newPos)
    {
        posText.text = newPos;
    }

    public void UpdateNameText(string newName)
    {
        nameText.text = newName;
    }

    public void CheckUpgrade(int newlevel)
    {
        if (newlevel > currLevel)
        {
            currLevel = newlevel;
            // update UI
            levelText.text = $"LVL. {currLevel}";
            animator.Play("levelUp");
        }
    }

    private float baseCoin,coinMultiplier,coinCap;
    private float baseHealth, healthMultiplier, healthCap;

    private int CoinCap(int level)
    {
        // base 20 
        return Mathf.Min((int)baseCoin+ level * (int)coinMultiplier,(int)coinCap);
    }
    private float HealthCap(int level)
    {
        // base 100
        return Mathf.Min(baseHealth+ level * healthMultiplier,healthCap);
    }
    public void SetCoinCapParameters(float baseCoin, float coinMultiplier,float cap)
    {
     
        this.baseCoin = baseCoin;
        this.coinMultiplier = coinMultiplier;
        this.coinCap = cap;
    }
    public void SetHealthCapParameters(float baseHealth, float healthMultiplier,float cap)
    {
        this.baseHealth = baseHealth;
        this.healthMultiplier = healthMultiplier;
        this.healthCap = cap;
    }
    
}