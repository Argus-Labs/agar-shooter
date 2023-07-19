using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.Pool;

public class AttackAnimSpawner : MonoBehaviour
{
   // based on the code from DamageTextSpawner.cs, create a spawner for AttackAnim
   [SerializeField] private AttackAnim attackAnimPrefab;
   private ObjectPool<AttackAnim> _pool;
   [SerializeField] private DamageTextSpawner damageTextSpawner;

   private void Start()
   {
      _pool = new ObjectPool<AttackAnim>(() =>
      {
         return Instantiate(attackAnimPrefab);
      }, anim =>
      {
         anim.gameObject.SetActive(true);
      }, anim =>
      {
         anim.gameObject.SetActive(false);
      }, anim =>
      {
         Destroy(anim.gameObject);
      },false,10,100);
   }
   
   public void Create(Vector2 origin,Vector2 target,int damage)
   {
      AttackAnim anim = _pool.Get();
      anim.Setup(origin,target,damage);
      anim.Init(KillAnim);
   }
   
   public void KillAnim(AttackAnim anim)
   {
      _pool.Release(anim);
      damageTextSpawner.Create(anim.damage,anim.target);
   }
}
