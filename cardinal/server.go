package main

import (
	"fmt"
	"github.com/argus-labs/new-game/components"
	"github.com/argus-labs/new-game/types"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
	"math"
)

func RemoveCoin(coinID types.Pair[storage.EntityID, types.Triple[float64, float64, int]]) (int, error) {
	coin, err := CoinComp.Get(World, coinID.First)

	if err != nil {
		return -1, fmt.Errorf("Cardinal: could not get coin entity: %w", err)
	}

	mutex.Lock()
	delete(CoinMap[types.Pair[int, int]{int(math.Floor(coinID.Second.First / GameParams.CSize)), int(math.Floor(coinID.Second.Second / GameParams.CSize))}], coinID)
	mutex.Unlock()

	if err := World.Remove(coinID.First); err != nil {
		return -1, err
	}

	totalCoins--

	return coin.Val, nil
}

func TickTock() error { // testing function used to make the game tick
	err := World.Tick()
	return err
}

func CheckExtraction(player ModPlayer) int {
	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		fmt.Errorf("Error getting  player component: %w", err)
	}

	if playercomp.Coins > 0 && distance(playercomp.Loc, playercomp.Extract) <= ExtractionRadius {
		PlayerComp.Update(World, Players[player.Name], func(comp components.PlayerComponent) components.PlayerComponent {
			comp.Coins = 0 // extraction point offloading

			return comp
		})

		return playercomp.Coins
	} else {
		return 0
	}
}

func AddTestPlayer(player components.PlayerComponent) error {
	if _, contains := Players[player.Name]; contains { // player already exists; don't do anything
		return nil
	}

	playerID, err := World.Create(PlayerComp) // creates new player
	if err != nil {
		return fmt.Errorf("Error adding player to world: %w", err)
	}
	Players[player.Name] = playerID

	PlayerComp.Set(World, Players[player.Name], player) // default player

	playercomp, err := PlayerComp.Get(World, Players[player.Name])

	if err != nil {
		return fmt.Errorf("Error getting location with callback function: %w", err)
	}

	newPlayer := types.Pair[storage.EntityID, types.Pair[float64, float64]]{Players[player.Name], playercomp.Loc}
	PlayerMap[GetCell(player.Loc)][newPlayer] = pewp

	return nil
}

func RecentAttacks() []AttackTriple {
	return Attacks
}
