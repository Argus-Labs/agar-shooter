using UnityEngine;

namespace ArgusLabs
{
    public class FireBallMovementTesting : MonoBehaviour
    {

        [SerializeField] GameObject _target;
        // Start is called before the first frame update
        void Start()
        {
        
        }

        // Update is called once per frame
        void Update()
        {
            gameObject.transform.Translate(Vector3.Normalize((_target.transform.position - transform.position)) * Time.deltaTime * 10);
        }
    }
}
