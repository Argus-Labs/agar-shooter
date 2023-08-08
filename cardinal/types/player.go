package types

type BarePlayer struct {
	Name   string
	Health int
	Coins  int
	//Weapon Weapon
	//ExtractX float64
	//ExtractY float64
	LocX     float64
	LocY     float64
	IsRight  bool
	InputNum int
}

type TestPlayer struct {
	Name     string
	Health   int
	Coins    int
	Weapon   Weapon
	ExtractX float64
	ExtractY float64
	LocX     float64
	LocY     float64
}

type ModPlayer struct { // for adding and removing players
	Name string
}
