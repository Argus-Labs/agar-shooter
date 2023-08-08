using System;
using UnityEngine;
using UnityEngine.Serialization;

public class BestPlayerIndicator : MonoBehaviour
{
    public Transform bestPlayerTransform;
    public Renderer extractionPointSpriteRenderer;
    public GameObject indicator;

    public void SetExtractPoint(Transform extractPoint,Renderer spriteRenderer)
    {
        this.enabled= true;
        bestPlayerTransform = extractPoint;
        extractionPointSpriteRenderer = spriteRenderer;
    }

    private void LateUpdate()
    {
        if (bestPlayerTransform == null)
        {
            return;
        }
        Vector3 extractionPointScreenPos = Camera.main.WorldToScreenPoint(bestPlayerTransform.position);
        Vector3 screenCenter = new Vector3(Screen.width, Screen.height, 0) / 2f;
        Vector3 enemyToCenterVector = screenCenter - extractionPointScreenPos;
        if (!extractionPointSpriteRenderer.isVisible)
        {
            if (indicator.activeSelf == false)
            {
                indicator.SetActive(true);
            }

            Vector3 indicatorPosition = GetIndicatorPosition(enemyToCenterVector);
            Vector3 newpos = Camera.main.ScreenToWorldPoint(indicatorPosition);
            newpos.z = 0;
            indicator.transform.position = newpos;
            indicator.transform.LookAt(newpos-enemyToCenterVector,new Vector3(0,0,-1));
        }
        else
        {
            if (indicator.activeSelf)
            {
                indicator.SetActive(false);
            }
        }
    }

    private void OnDisable()
    {
        if (indicator!=null)
        {
            indicator.SetActive(false);
        }
    }

    private Vector3 GetIndicatorPosition(Vector3 direction)
    {
        Vector3 screenCenter = new Vector3(Screen.width, Screen.height, 0) / 2f;

        float xRatio = Mathf.Abs(direction.x) / Screen.width;
        float yRatio = Mathf.Abs(direction.y) / Screen.height;

        float indicatorX;
        float indicatorY;
        if (xRatio > yRatio)
        {
            indicatorX = direction.x > 0 ? 0 : Screen.width;
            indicatorY = screenCenter.y - (direction.y / direction.x) * Mathf.Abs(indicatorX - screenCenter.x) *
                Mathf.Sign(direction.x);
        }
        else
        {
            indicatorY = direction.y > 0 ? 0 : Screen.height;
            indicatorX =
                screenCenter.x - (direction.x / direction.y) * Mathf.Abs(indicatorY - screenCenter.y) *
                Mathf.Sign(direction.y);
        }

        Vector3 indicatorPosition = new Vector3(indicatorX, indicatorY, 0f);
        indicatorPosition+=direction.normalized*50;
        return indicatorPosition;
    }
}