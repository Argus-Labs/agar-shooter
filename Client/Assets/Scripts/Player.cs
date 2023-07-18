using Nakama.TinyJson;
using TMPro;
using UnityEngine;
using UnityEngine.UI;

public class Player : MonoBehaviour
{
    public GameManager gameManager;
    [SerializeField] private int health = 100;
    [SerializeField] private int coin = 0;
    [SerializeField] private float speed;
    [SerializeField] private bool isRight = true;
    [SerializeField] private InputProfile inputProfile;
    [SerializeField] private TextMeshProUGUI coinText;
    [SerializeField] private Transform rangeIndicator;
    [SerializeField] private Slider healthBar;
    [SerializeField] private Transform sprite;
    [SerializeField] private SpriteRenderer body;
    [SerializeField] TextMeshProUGUI posText;
    private int sequenceNumber = 0;
    CircularArray<PlayerInputExtraInfo> pendingInputs = new CircularArray<PlayerInputExtraInfo>(100);
    public Vector2 pos = new Vector2(0, 0);

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


    private void Update()                       
    {
        // print(sequenceNumber);
        UploadPlayerInput();
        sequenceNumber++;
        // update the player position on main thread
        transform.localPosition = pos;
        IsRight = isRight;
    }


    private void UploadPlayerInput()
    {
        PlayerInput input = new PlayerInput(gameManager.UserId, Input.GetKey(inputProfile.up),
            Input.GetKey(inputProfile.down), Input.GetKey(inputProfile.left), Input.GetKey(inputProfile.right),
            sequenceNumber,Time.deltaTime);
        int opCode = 17;
        gameManager.SendMessageToServer(opCode, input.ToJson());
        ApplyInput(input, Time.deltaTime);
        pendingInputs.Enqueue(new PlayerInputExtraInfo(input, Time.deltaTime, pos));
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
        healthBar.value = health / 100f;
    }

    public void UpdateCoins(int newCoinCount)
    {
        coin = newCoinCount;
        coinText.text = coin.ToString();
        // update UI
    }

    public void SetColor(Color color)
    {
        body.color = color;
    }

    public void UpdatePosText(string newPos)
    {
        posText.text = newPos;
    }
}