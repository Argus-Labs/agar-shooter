# Vampire-Farming Game created with Cardinal, Unity, & Nakama
The purpose of this game is to use the Cardinal ECS game development library and verify that Cardinal operations work well with distributed client-server interfaces like Nakama. The theme of the game is vampire/farming with most game interaction occurring in the vampire part of the game and individual character growth in the farming part of the game. Server and client files are described below, as are features of the game.
## Server
Server code is located in [server](server). The `.go` files contain the server code and testcases while the Dockerfilee contains commands for [start.sh](start.sh) to run. We describe the function of the following `.go` files:
* [main.go](server/main.go)
* [endpoints.go](server/endpoints.go)
* [server.go](server/server.go)
* [vars.go](server/vars.go)
* [structs.go](server/structs.go)
* [components.go](server/components.go)
* [systems.go](server/systems.go)
* [beeg_test.go](server/beeg_test.go)

`beeg_test.go` contains testcases verifying that various game operations work as intended on Cardinal.  
`main.go` contains code for associating names with go functions so that Nakama can call endpoints by user-specified name. It also contains some helper functions for parsing HTTP requests and responses.  
`endpoints.go` contains the endpoint functions, which includes client test functions as well as functions called by Nakama's MatchLoop and player add and remove functions. There are currently 11 endpoint functions that add and remove players, move players, get the coins near a player and their status for client display, check whether a player is near an extraction point, add health for testing purposes, get all player attacks executed within the last game tick, create a game instance in Cardinal, and execute a Cardinal game tick.  
`server.go` contains the functions called by the endpoint functions. The endpoint functions parse the client request and send the requisite data to the server functions, which output data. The endpoint functions then package this into a response and send it back to Nakama. This file also contains an add player function `AddTestPlayer` used only for server testing.  
`vars.go` contains all global variables and constants used by the server during the game. This includes the Cardinal ECSWorld object; coin, health, weapon, and player maps and components; a transaction queue; the number of cells that span the grid; a map of weapons; a mutex for allowing asynchronous coin addition and removal; a pair representing the size of a client's POV on the game board, and a list of recently-executed attacks.  
`structs.go` contains all game-related structs and their struct methods that are not Cardinal component structs. This includes pairs, triples, an interface for the two, weapon structs, player structs for sending information and testing, game structs, and item structs  
`components.go` contains ECS component structs for use in directly interacting with Cardinal, specifically a health, coin, wepaon, and player component as well as the necessary struct methods.  
`systems.go` contains the Cardinal systems used to update entities during each game tick as well as helper functions used to make running these systems easier to understand. The current systems are
* `processMoves`: a system for taking all player inputs sent within the last tick and simulating them at the tickrate at which they were sent rather than the server tickrate, saving the resulting direction 
* `makeMoves`: a system that finds the nearest neighbor of each player, stores all attacks between players, applies the average direction of all inputs processed within the last tick to each player, finds all coins close enough to the line segment between the previous location and the current location and gives them to the player, then executes all player attacks  

### Server Optimizations
The code currently checks for player attacks, coin collection and insertion, and nearby coins naively --- coins are stored in a map of cells where each cell corresponds to some area of the game board and finding the player-player and coin-coin nearest neighbor is done by iterating over all players or coins, respectively, in the cells neighboring a player and checking whether the closest player is too close. This could be sped up in two ways: each object only checks nearby objects within some radius (within some weapon range, for example), and each object queries the nearest neighbor faster than naively iterating over all other players. We can solve the second problem using a 2d-BST or quadtree --- if we were to use a quadtree, we would need to limit the precision of player coordinates when storing them in the tree because the depth of the tree is linear in the number of players divided by the minimum distance between two players. Thus, when two players are sufficiently close to each other, we can add them to a queue or map. This leads to the worst-case occurring when all players are very close to each other, in which case nearest-neighbor-checking degenerates to quadratic time.  
We start with coin optimization. Currently, coins are stored in a map of cell pairs but this takes the same space as storing in a 2d array, so I'll switch them to a 2d array. However, there may be many coins in each cells (though we can limit this to some number so that randomly choosing a location in the cell succeeds with some sufficiently high probability). We will not make any changes to coin storage and nearest-neighbor because the number of coins in each cell is small, but we switch to capping the number of coins based on the size of the game board and the number of players in the game, only adding coins if the number of coins is less than the coin cap, and regenerating coins when players eat them. 
We optimize players by putting them into a balanced kD tree; we can find the nearest neighbor in log time, which is optimal. For this, we use an existing kD tree implementation in Go.  

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
