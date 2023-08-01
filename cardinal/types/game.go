package types

type Void struct{}

var Pewp Void

type Weapon int

type WeaponData struct {
	Attack int
	Range  float64
}

type Game struct {
	Dims    Pair[float64, float64]
	CSize   float64  // cell size
	Players []string // list of players
}

type AttackTriple struct {
	AttackerID, DefenderID string
	Damage                 int
}

type NearbyCoin struct { // nearby coins for client retrieval
	First, Second float64
	Value         int
}
