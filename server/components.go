// Cardinal component structs
package main

import (
	"strconv"
	"github.com/argus-labs/world-engine/cardinal/ecs/storage"
)

type HealthComponent struct {
	Loc Pair[float64, float64]// location
	Val int// how much health it contains
}

type CoinComponent struct {
	Loc Pair[float64, float64]
	Val int// how many coins the component represents; could represent different denominations as larger or different colored coins
}

type WeaponComponent struct {
	Loc Pair[float64, float64]
	Val Weapon// weapon type
	Ammo int// number of attacks left
	LastAttack int64// time of last attack
}

type PlayerComponent struct {
	Name string// username; ip for now
	Health int// current player health (cap enforced in update loop)
	Coins int// how much money the player has
	Weapon storage.EntityID// current player weapon; default is 0 for Melee
	Loc Pair[float64, float64]// current location
	Dir Pair[float64, float64]// array of movement directions with range [[-1,1],[-1,1]] where each pair is the movement at a given timestep (divided uniformly over the tick) and the first direction is the one that determines player movement
	LastMove Pair[float64, float64]// last player move; this must be a pair of ints in [[-1,1],[-1,1]]
	Extract Pair[float64, float64]// extraction point; as long as the player is within some distance of the extraction point, player coins are offloaded
	IsRight bool// whether player is facing right
	MoveNum int// most recently-processed move
}

func (p PlayerComponent) Simplify() BarePlayer {
	return BarePlayer{p.Name, p.Health, p.Coins, p.Loc.First, p.Loc.Second, p.IsRight, p.MoveNum}// update Simplify for weapons & extraction point
}

func (p PlayerComponent) Testify() TestPlayer {
	weapon, _ := WeaponComp.Get(World, p.Weapon)
	return TestPlayer{p.Name, p.Health, p.Coins, weapon.Val, p.Extract.First, p.Extract.Second, p.Loc.First, p.Loc.Second}
}

func (p PlayerComponent) String() string {
	s := ""
	s += "Name: " + p.Name + "\n"
	s += "Health: " + strconv.Itoa(p.Health) + "\n"
	s += "Coins: " + strconv.Itoa(p.Coins) + "\n"
	weapon, _ := WeaponComp.Get(World, p.Weapon)
	s += "Weapon: " + strconv.Itoa(int(weapon.Val)) + "\n"
	s += "Loc: " + strconv.FormatFloat(float64(p.Loc.First), 'e', -1, 32) + " " + strconv.FormatFloat(float64(p.Loc.Second), 'e', -1, 32) + "\n"
	s += "Dir: " + strconv.FormatFloat(float64(p.Dir.First), 'e', -1, 32) + " " + strconv.FormatFloat(float64(p.Dir.Second), 'e', -1, 32)

	return s
}
