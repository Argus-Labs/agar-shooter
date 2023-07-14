package types

type void struct{}

var pewp void

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

type Move struct {
	PlayerID              string
	Up                    bool
	Down                  bool
	Left                  bool
	Right                 bool
	Input_sequence_number int
	Delta                 float64
}

type AttackTriple struct {
	AttackerID, DefenderID string
	Damage                 int
}

type NearbyCoin struct { // nearby coins for client retrieval
	X, Y  float64
	Value int
}
