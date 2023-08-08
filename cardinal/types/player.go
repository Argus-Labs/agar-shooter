package types

type BarePlayer struct {
	PersonaTag	string
	Health		int
	Coins		int
	LocX		float64
	LocY		float64
	IsRight		bool
	InputNum	int
	Level		int
}

type TestPlayer struct {
	PersonaTag	string
	Health		int
	Coins		int
	Weapon		Weapon
	LocX		float64
	LocY		float64
}

type ModPlayer struct { // for modifying players
	PersonaTag	string
}

type AddPlayer struct {// for adding players
	PersonaTag	string
	Coins		int
}
