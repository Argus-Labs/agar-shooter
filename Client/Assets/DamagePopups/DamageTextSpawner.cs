using System;
using UnityEngine;
using UnityEngine.Pool;

public class DamageTextSpawner : MonoBehaviour
    {
        [SerializeField] private DamagePopup damagePopupPrefab;

        private ObjectPool<DamagePopup> _pool;
        private  int sortingOrder = 0;

        private void Start()
        {
            _pool = new ObjectPool<DamagePopup>(() =>
            {
                return Instantiate(damagePopupPrefab);
            }, text =>
            {
                text.gameObject.SetActive(true);
            }, text =>
            {
                text.gameObject.SetActive(false);
            }, text =>
            {
                Destroy(text.gameObject);
            },false,10,100);
        }

        public void Create(int damageNumber,Vector2 position)
        {
            var text = _pool.Get();
            text.Setup(damageNumber,sortingOrder++,position);
            text.Init(KillText);
            
        }

        public void KillText(DamagePopup popup)
        {
            _pool.Release(popup);
        }
    }
