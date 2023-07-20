using System.Collections.Generic;
using TMPro;
using UnityEngine;

public class ScoreBoard : MonoBehaviour
{
    public TextMeshProUGUI scoreText;

    // players dictionary contains player name and score
    // playerSelf is the player name of the current player
    public int Refresh(Dictionary<string, int> players, string playerSelf, int scoreboardSize)
    {
        // based on players dictionary, update the scoreText format "1. player1" just rank and name player itself must be in the output 
        // if player count <= scoreboardSize display all players
        // if player count > scoreboardSize display top scoreboardSize-1 players and last line is playerSelf
        
        List<string> playerList = new List<string>(players.Keys);
        playerList.Sort((a, b) => players[b].CompareTo(players[a]));
        bool selfInList = playerList.Contains(playerSelf);
        if (!selfInList)
        {
            Debug.LogError("ScoreBoard:Cannot find the playerSelf in the players");
            return -1;
        }
        if (scoreboardSize < 0)
        {
            Debug.LogError("ScoreBoard: Scoreboard size should be a positive number.");
            return -1;
        }
        bool enoughSpace = playerList.Count <= scoreboardSize;
        int count = Mathf.Min(playerList.Count, scoreboardSize);
        int selfRank = playerList.IndexOf(playerSelf);
        string newText = "";
        for (int i = 0; i < count-1; i++)
        {
            int rank = i + 1;
            string playerName = playerList[i];
            if (selfRank == i)
            {
                newText += $"<color=green><b>{rank}. {playerName}</b></color>\n";
            }
            else if (rank == 1)
            {
                newText += $"<color=red>{rank}. {playerName}</color>\n";
            }
            {
                newText += $"{rank}. {playerName}\n";
            }
        }
        // handle the last line
        if (enoughSpace)
        {
            string playerName = playerList[count - 1];
            if (selfRank == count - 1)
            {
                newText += $"<color=green><b>{count}. {playerName}</b></color>";
            }
            else
            {
                newText += $"{count}. {playerName}";
            }
        }
        else
        {
            newText += $"<color=green><b>{selfRank+1}. {playerSelf}</b></color>";
        }

        scoreText.text = newText;
        return selfRank+1;

    }
    
}