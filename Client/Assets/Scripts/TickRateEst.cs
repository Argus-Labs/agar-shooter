using Nakama;
using UnityEngine;

// This class is used to estimate the tick rate of the server by tracking the last time we received a match state
public class TickRateEst : MonoBehaviour
{
    public GameManager gameManager;
    private float[] packetArrivalTimes = new float[5]; // Array to store the arrival times of the last 5 packets
    private int currentIndex = 0; // Index to keep track of the current position in the array

    private void Start()
    {
        gameManager.OnMatchStateReceived += OnMatchStateReceived;
    }

    private void OnApplicationQuit()
    {
        gameManager.OnMatchStateReceived -= OnMatchStateReceived;
    }

    private void OnMatchStateReceived(IMatchState obj)
    {
        // Calculate the time since the last packet arrival
        float timeSinceLastPacket = Time.time - packetArrivalTimes[currentIndex];

        // Update the packet arrival time at the current index
        packetArrivalTimes[currentIndex] = Time.time;

        // Increment the current index or reset it if it reaches the end of the array
        currentIndex = (currentIndex + 1) % packetArrivalTimes.Length;

        // Calculate the average packet arrival time
        float avgPacketArrivalTime = CalculateAveragePacketArrivalTime();

        // Calculate the tick rate based on the average packet arrival time
        float tickRate = 1f / avgPacketArrivalTime;

        // Print the tick rate
        Debug.Log("Tick Rate: " + tickRate);
    }

    private float CalculateAveragePacketArrivalTime()
    {
        float sum = 0f;

        // Iterate through the packet arrival times and sum them up
        for (int i = 0; i < packetArrivalTimes.Length; i++)
        {
            sum += packetArrivalTimes[i];
        }

        // Calculate the average by dividing the sum by the number of elements
        return sum / packetArrivalTimes.Length;
    }
}