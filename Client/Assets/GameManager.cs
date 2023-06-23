using System.Collections;
using System.Collections.Generic;
using System.Threading.Tasks;
using UnityEngine;

public class GameManager : MonoBehaviour
{
    public NakamaConnection nakamaConnection;
    // Start is called before the first frame update
    async Task Start()
    {
        await nakamaConnection.Connect();
    }
    

    
}
