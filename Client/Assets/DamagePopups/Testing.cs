// create a test script that call DamageTextSpawner.create when I press right click pass the mouse position,
// and the damage number is fixed 5

using UnityEngine;

public class Testing : MonoBehaviour
{
    public DamageTextSpawner _damageTextSpawner;
    public AttackAnimSpawner _attackAnimSpawner;
    private void Update()
    {
        // if (Input.GetMouseButtonDown(1))
        // {
        //     Debug.Log("right click");
        //     var mousePos = Camera.main.ScreenToWorldPoint(Input.mousePosition);
        //     mousePos.z = 0;
        //     _damageTextSpawner.Create(5,mousePos);
        // }
        
        //now test the AttackAnimSpawner
        if (Input.GetMouseButtonDown(1))
        {
            Debug.Log("right click");
            var mousePos = Camera.main.ScreenToWorldPoint(Input.mousePosition);
            mousePos.z = 0;
            _attackAnimSpawner.Create(transform.position,mousePos,1);
        }
    }
}