/*
Copyright 2021 Heroic Labs

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using Nakama;
using UnityEngine;

/// <summary>
/// A singleton class that handles all connectivity with the Nakama server.
/// </summary>
[Serializable]
[CreateAssetMenu]
public class NakamaConnection : ScriptableObject
{
    public string Scheme = "http";
    public string Host = "localhost";
    public int Port = 7350;
    public string ServerKey = "defaultkey";
    public string matchName = "singleton_match";
    
    // if we want the player to enter the game using still valid session 
    // private const string SessionPrefName = "nakama.session";

    public IClient client;
    public ISession session;
    public ISocket socket;
    public string matchID = "";
    

    private string currentMatchmakingTicket;
    private string currentMatchId;

    /// <summary>
    /// Connects to the Nakama server using device authentication and opens socket for realtime communication.
    /// </summary>
    public async Task<string> Connect()
    {
        // Connect to the Nakama server.
        client = new Nakama.Client(Scheme, Host, Port, ServerKey, UnityWebRequestAdapter.Instance);

        // // Attempt to restore an existing user session.
        // var authToken = PlayerPrefs.GetString(SessionPrefName);
        // if (!string.IsNullOrEmpty(authToken))
        // {
        //     var session = Nakama.Session.Restore(authToken);
        //     if (!session.IsExpired)
        //     {
        //         Session = session;
        //     }
        // }
        try
        {
            session = await client.AuthenticateDeviceAsync(Utility.DeviceUniqueIdentifier, create: false);
        }
        catch (ApiResponseException e)
        {
            Debug.Log("Failed to authenticate with Nakama server, creating new account");
            Debug.Log(e);
            session = await client.AuthenticateDeviceAsync(Utility.DeviceUniqueIdentifier, create: true);
        }
        socket = client.NewSocket();
        await socket.ConnectAsync(session,true);
        Debug.Log("Connected to Nakama server");
        Debug.Log(session);
        Debug.Log(socket);
        // join the match 
        var minPlayers = 0;
        var maxPlayers = 10;
        var limit = 10;
        var authoritative = true;
        var label = "";
        var query = "";
        var result = await client.ListMatchesAsync(session, minPlayers, maxPlayers, limit, authoritative, label, query);
        // because we only have one match, we can just join the first one 
        foreach (var match in result.Matches)
        {
            Debug.LogFormat("{0}: {1}/10 players", match.MatchId, match.Size);
            await socket.JoinMatchAsync(match.MatchId);
            matchID = match.MatchId;
            break;
        }
        
        return session.Username;
    }

}