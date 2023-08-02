package types

type BarePlayer struct {
	Name		string
	Health		int
	Coins		int
	LocX		float64
	LocY		float64
	IsRight		bool
	InputNum	int
	Level		int
}

type TestPlayer struct {
	Name		string
	Health		int
	Coins		int
	Weapon		Weapon
	LocX		float64
	LocY		float64
}

type ModPlayer struct { // for modifying players
	Name		string
}

type AddPlayer struct {// for adding players
	Name		string
	Coins		int
}
