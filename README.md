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

## Function TL:DR's
* `main.go` contains code for associating names with go functions so that Nakama can call endpoints by user-specified name. It also contains some helper functions for parsing HTTP requests and responses.  
* `endpoints.go` contains the endpoint functions, which includes client test functions as well as functions called by Nakama's MatchLoop and player add and remove functions. There are currently 11 endpoint functions that add and remove players, move players, get the coins near a player and their status for client display, check whether a player is near an extraction point, add health for testing purposes, get all player attacks executed within the last game tick, create a game instance in Cardinal, and execute a Cardinal game tick.  
* `server.go` contains the functions called by the endpoint functions. The endpoint functions parse the client request and send the requisite data to the server functions, which output data. The endpoint functions then package this into a response and send it back to Nakama. This file also contains an add player function `AddTestPlayer` used only for server testing.  
* `vars.go` contains all global variables and constants used by the server during the game. This includes the Cardinal ECSWorld object; coin, health, weapon, and player maps and components; a transaction queue; the number of cells that span the grid; a map of weapons; a mutex for allowing asynchronous coin addition and removal; a pair representing the size of a client's POV on the game board, and a list of recently-executed attacks.  
* `structs.go` contains all game-related structs and their struct methods that are not Cardinal component structs. This includes pairs, triples, an interface for the two, weapon structs, player structs for sending information and testing, game structs, and item structs  
* `components.go` contains ECS component structs for use in directly interacting with Cardinal, specifically a health, coin, wepaon, and player component as well as the necessary struct methods.  
* `systems.go` contains the Cardinal systems used to update entities during each game tick as well as helper functions used to make running these systems easier to understand. The current systems are
* `processMoves`: a system for taking all player inputs sent within the last tick and simulating them at the tickrate at which they were sent rather than the server tickrate, saving the resulting direction 
* `makeMoves`: a system that finds the nearest neighbor of each player, stores all attacks between players, applies the average direction of all inputs processed within the last tick to each player, finds all coins close enough to the line segment between the previous location and the current location and gives them to the player, then executes all player attacks 
* `beeg_test.go` contains testcases verifying that various game operations work as intended on Cardinal.  

### Server Optimizations
The current coding framework incorporates functions to assess player attacks, manage coin collection and insertion, and identify nearby coins. However, this process could benefit from optimization.

Coins are currently stored in a map of cells, each corresponding to a specific area on the game board. To find the nearest player-player and coin-coin, we iterate over all players or coins in the neighboring cells of a player. Then, we evaluate whether the closest player is too close.

There are two ways to accelerate this process:
1. Each object should only check for objects within a certain radius (for example, within a weapon's range).
2. The method of querying for the nearest neighbor should be faster than the current iteration over all other players.

The second issue can be addressed using a 2D Binary Search Tree (BST) or quadtree. However, the use of a quadtree necessitates the limitation of player coordinate precision when storing in the tree. This is because the tree depth is directly proportional to the number of players divided by the minimum distance between them. Therefore, when two players are sufficiently close, we can add them to a queue or map. In the worst-case scenario where all players are very close to each other, the time taken for nearest-neighbor-checking can degenerate to quadratic time.

Now, let's turn our attention to coin optimization. Currently, coins are stored in a map of cell pairs, which occupies the same space as a 2D array. Therefore, the switch to a 2D array seems justified. However, we must account for the fact that each cell may contain multiple coins. While we can limit this number to ensure a high probability of randomly choosing a location in the cell, we will retain the current coin storage and nearest-neighbor methodologies since the number of coins in each cell is relatively small. Instead, we propose capping the total number of coins based on the game board's size and the number of players in the game. New coins will only be added if the total number is below this cap, and coins will be regenerated when players collect them.

For player optimization, we propose using a balanced kD tree, which will allow for finding the nearest neighbor in logarithmic time - an optimal solution. We can leverage an existing kD tree implementation in Go for this purpose.

## Unity
Client code is located in [Client](Client).

## Nakama
Nakama code is located in [nakama/main.go](nakama/main.go).

## Running the game
To start Nakama and the gameplay server, run the following command:
```bash
./start.sh
```

After updating the gameplay server, the server can be rebuilt via:
```bash
./restart_server.sh
```

Note, if any server endpoints have been added or removed, Nakama must be relaunched (via the start.sh script)

Once Nakama and the gameplay server are running, visit `localhost:7351` to access Nakama. For local development, use `admin:password` as your login credentials.

The Account tab on the left will give you access to a valid account ID.

The API Explorer tab on the left will allow you to make requests to the gameplay server.
