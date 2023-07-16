package types

import (
	"math"
)

type Mult interface {
	getFirst() float64
	getSecond() float64
}

func GetCell(loc Mult, cellSize float64) Pair[int, int] {
	return Pair[int, int]{int(math.Floor(loc.getFirst() / cellSize)), int(math.Floor(loc.getSecond() / cellSize))}
}

type Pair[T1 any, T2 any] struct { // inherits Mult
	First  T1
	Second T2
}

func (p Pair[float64, any]) getFirst() float64 {
	return p.First
}

func (p Pair[any, float64]) getSecond() float64 {
	return p.Second
}

type Triple[T1 any, T2 any, T3 any] struct { // inherits Mult
	First  T1
	Second T2
	Third  T3
}

func (t Triple[float64, any, void]) getFirst() float64 {
	return t.First
}

func (t Triple[any, float64, void]) getSecond() float64 {
	return t.Second
}
