using System.Collections;
using System.Collections.Generic;
using UnityEngine;

namespace ArgusLabs
{
    public class FloorBuilder : MonoBehaviour
    {

        [SerializeField] GameObject _floorTile;
        int _tileCount;
        // Start is called before the first frame update
        void Start()
        {
            _tileCount = 0;
            while(_tileCount <= 64)
            {
                int posX = _tileCount%16;
                int posY = 0 + (int)_tileCount / 16;
                Instantiate(_floorTile, new Vector3(posX, posY, 0f), Quaternion.identity);
                _tileCount++;
            }
        }

        // Update is called once per frame
        void Update()
        {
        
        }
    }
}
