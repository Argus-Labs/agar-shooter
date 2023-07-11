This is a sample gameplay server that uses Nakama to proxy gameplay requests to a separate server.
# Vampire-Farming Game created with Cardinal, Unity, & Nakama
The purpose of this game is to use the Cardinal ECS game development library and verify that Cardinal operations work well with distributed client-server interfaces like Nakama. The theme of the game is vampire/farming with most game interaction occurring in the vampire part of the game and individual character growth in the farming part of the game. Server and client files are described below, as are features of the game.
## Server
Server code is located in [server](server). The `.go` files contain the server code and testcases while the Dockerfilee contains commands for [start.sh](start.sh) to run.

## Unity
Client code is located in [Client](Client).

## Nakama
Nakama code is located in [nakama/main.go](nakama/main.go).

## Running the game
To start nakama and the gameplay server, run the following command:
```bash
./start.sh
```

After updating the gameplay server, the server can be rebuilt via:
```bash
./restart_server.sh
```

Note, if any server endpoints have been added or removed nakama must be relaunched (via the start.sh script)

Once nakama and the gameplay server are running, visit `localhost:7351` to access nakama. For local development, use `admin:password` as your login credentials.

The Account tab on the left will give you access to a valid account ID.

The API Explorer tab on the left will allow you to make requests to the gameplay server.
