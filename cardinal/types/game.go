package types

type Void struct{}

var Pewp Void

type WeaponData struct {
	Attack int
	Range  float64
	MaxAmmo	int
	Reload int64
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
	X, Y		float64
	Value       int
}

type NearbyHealth struct { // nearby coins for client retrieval
	X, Y		float64
	Value       int
}

type Package[T any] struct { // used to parse JSON objects with the new PersonaTag stuff
	Body T
}
