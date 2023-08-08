# TBD

TBD is a agar.io style real-time multiplayer PVP game. The purpose of this game is to demonstrate usage of the [World Engine](https://www.github.com/argus-labs/world-engine) and more specifically, World Engine's ECS-based game shard [Cardinal](https://www.github.com/Argus-Labs/world-engine/tree/main/cardinal). In this game we look to demonstrate that capabilities of Cardinal and create the **first ever, real-time, fully on-chain game** :).

## Usage 

Once you've cloned the repo and are in the directory, you can start the local Nakama and Cardinal servers by doing the following: 

```bash
./start.sh
```

After updating the gameplay server, the server can be rebuilt via:

```bash
./restart_server.sh
```
To play the game, you must connect a game client to Nakama to allow it to communicate with Cardinal. The game client can be written in any language, but for this game we write the client in Unity with code in the `Client` folder.

Note, if any server endpoints have been added or removed, Nakama must be relaunched (via the start.sh script). 

Once Nakama and the gameplay server are running, visit `localhost:7351` to access Nakama. For local development, use `admin:password` as your login credentials. The Account tab on the left will give you access to a valid account ID. The API Explorer tab on the left will allow you to make requests to the gameplay server.

## Structure 

```bash
├── Client: Unity client
│   ├── Assets
│   ├── Packages
│   └── ProjectSettings
├── cardinal: the cardinal ECS game shard
│   ├── components: ECS component definitions
│   ├── game: global constants
│   ├── main.go: main entrypoint for the cardinal ECS game shard
│   ├── read: read endpoints for the game state
│   ├── systems: ECS system definitions
│   ├── tx: transaction endpoints for the game state, each maps to a system
│   ├── types: game related types
│   ├── utils: utility functions
├── nakama: provides auth and netcode between clients and cardinal
│   ├── cardinal.go: helper functions for interating with cardinal
│   ├── main.go: main entrypoint for the nakama server
│   ├── match.go: Nakama Match related functions
├── restart_backend.sh: restarts the cardinal server
└── start.sh: starts the nakama and cardinal servers
```

## Server Optimizations
The current coding framework incorporates functions to assess player attacks, manage coin collection and insertion, and identify nearby coins. However, this process could benefit from optimization.

Coins are currently stored in a map of cells, each corresponding to a specific area on the game board. To find the nearest player-player and coin-coin, we iterate over all players or coins in the neighboring cells of a player. Then, we evaluate whether the closest player is too close.

There are two ways to accelerate this process:
1. Each object should only check for objects within a certain radius (for example, within a weapon's range).
2. The method of querying for the nearest neighbor should be faster than the current iteration over all other players.

The second issue can be addressed using a 2D Binary Search Tree (BST) or quadtree. However, the use of a quadtree necessitates the limitation of player coordinate precision when storing in the tree. This is because the tree depth is directly proportional to the number of players divided by the minimum distance between them. Therefore, when two players are sufficiently close, we can add them to a queue or map. In the worst-case scenario where all players are very close to each other, the time taken for nearest-neighbor-checking can degenerate to quadratic time.

Now, let's turn our attention to coin optimization. Currently, coins are stored in a map of cell pairs, which occupies the same space as a 2D array. Therefore, the switch to a 2D array seems justified. However, we must account for the fact that each cell may contain multiple coins. While we can limit this number to ensure a high probability of randomly choosing a location in the cell, we will retain the current coin storage and nearest-neighbor methodologies since the number of coins in each cell is relatively small. Instead, we propose capping the total number of coins based on the game board's size and the number of players in the game. New coins will only be added if the total number is below this cap, and coins will be regenerated when players collect them.

For player optimization, we propose using a balanced kD tree, which will allow for finding the nearest neighbor in logarithmic time - an optimal solution. We can leverage an existing kD tree implementation in Go for this purpose.
