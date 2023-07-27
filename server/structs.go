// Non-Cardinal game structs & interfaces
package main

import (
	"math"
	"github.com/downflux/go-geometry/nd/vector"
)

type void struct{}
var pewp void

type Mult interface {
	getFirst() float64
	getSecond() float64
}

func GetCell(loc Mult) Pair[int,int] {
	return Pair[int,int]{int(math.Floor(loc.getFirst()/GameParams.CSize)), int(math.Floor(loc.getSecond()/GameParams.CSize))}
}

type P struct {
	p   vector.V
	Name string
}

func (p *P) P() vector.V {
	return p.p
}

func (p *P) Equal(q *P) bool {
	return vector.Within(p.P(), q.P()) && p.Name == q.Name
}

type Pair[T1 any, T2 any] struct {// inherits Mult
	First T1
	Second T2
}

func (p Pair[float64, any]) getFirst() float64 {
	return p.First
}

func (p Pair[any, float64]) getSecond() float64 {
	return p.Second
}

type Triple[T1 any, T2 any, T3 any] struct {// inherits Mult
	First T1
	Second T2
	Third T3
}

func (t Triple[float64, any, void]) getFirst() float64 {
	return t.First
}

func (t Triple[any, float64, void]) getSecond() float64 {
	return t.Second
}

type Weapon int

type WeaponData struct {
	Attack int
	Range float64
	MaxAmmo int
	Reload int64
}

type BarePlayer struct {
	Name string
	Health int
	Coins int
	LocX float64
	LocY float64
	IsRight bool
	InputNum int
}

type TestPlayer struct {
	Name string
	Health int
	Coins int
	Weapon Weapon
	ExtractX float64
	ExtractY float64
	LocX float64
	LocY float64
}

type Game struct {
	Dims	Pair[float64, float64]
	CSize	float64// cell size
	Players	[]string// list of players
}

type Move struct {
	PlayerID				string
	Up						bool
	Down					bool
	Left					bool
	Right					bool
	Input_sequence_number	int
	Delta					float64
}

type AttackTriple struct {
	AttackerID, DefenderID	string
	Damage					int
}

type AddPlayer struct {// for adding and removing players
	Name	string
	Coins	int
}

type ModPlayer struct {// for adding and removing players
	Name	string
}

type NearbyCoin struct {// nearby coins for client retrieval
	X, Y	float64
	Value	int
}

type NearbyHealth struct {// nearby coins for client retrieval
	X, Y	float64
	Value	int
}
