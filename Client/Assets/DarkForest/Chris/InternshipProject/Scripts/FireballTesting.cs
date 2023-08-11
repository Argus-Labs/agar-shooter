using UnityEngine;

namespace ArgusLabs
{
    public class FireballTesting : MonoBehaviour
    {
        [SerializeField] float _firingRate;
        [SerializeField] GameObject _fireballGO;
        Vector3 mousePos;
        [SerializeField] GameObject shooter;

        float _timeSinceLastShot;
     
        // Start is called before the first frame update
        void Start()
        {
            _timeSinceLastShot = 0f;
            
        }

        // Update is called once per frame
        void Update()
        {
            mousePos = Input.mousePosition;
            if (Input.GetMouseButton(0))
            {
                _timeSinceLastShot += Time.deltaTime;
                if(_timeSinceLastShot >= _firingRate)
                {
                    Instantiate(_fireballGO, Camera.main.ViewportToWorldPoint(mousePos), Quaternion.identity);
                    _timeSinceLastShot = 0f;
                }
            }
        }
    }
}
