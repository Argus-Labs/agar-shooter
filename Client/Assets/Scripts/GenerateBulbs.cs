using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class GenerateBulbs : MonoBehaviour
{
    public GameObject bulb;
    public int bulbNum = 10;
    public float xRange = 8;
    public float yRange = 5;

    private void Start()
    {
        // generate bulbs with random position and random color
        for (int i = 0; i < bulbNum; i++)
        {
            Vector2 pos = new Vector2(UnityEngine.Random.Range(-xRange, xRange), UnityEngine.Random.Range(-yRange, yRange));
            GameObject newBulb = Instantiate(bulb, pos, Quaternion.identity);
            newBulb.GetComponent<SpriteRenderer>().color = UnityEngine.Random.ColorHSV();
        }
    }
}
