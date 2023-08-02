package types

type Mult interface {
	GetFirst() float64
	GetSecond() float64
}

type Pair[T1 any, T2 any] struct { // inherits Mult
	First  T1
	Second T2
}

func (p Pair[float64, any]) GetFirst() float64 {
	return p.First
}

func (p Pair[any, float64]) GetSecond() float64 {
	return p.Second
}

type Triple[T1 any, T2 any, T3 any] struct { // inherits Mult
	First  T1
	Second T2
	Third  T3
}

func (t Triple[float64, any, void]) GetFirst() float64 {
	return t.First
}

func (t Triple[any, float64, void]) GetSecond() float64 {
	return t.Second
}

type Weapon int
