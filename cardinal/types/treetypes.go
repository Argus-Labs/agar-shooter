package types

import (
	"github.com/downflux/go-geometry/nd/vector"
)

type P struct {
	Point	vector.V
	Name	string
}

func (p *P) P() vector.V {
	return p.Point
}

func (p *P) Equal(q *P) bool {
	return vector.Within(p.P(), q.P()) && p.Name == q.Name
}
