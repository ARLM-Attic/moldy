package world

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

type pos struct {
	X int
	Y int
}

func (self pos) eachNeighbour(dim pos, f func(p pos) bool) {
	nx := 0
	ny := 0
	np := pos{}
	for xd := -1; xd < 2; xd++ {
		for yd := -1; yd < 2; yd++ {
			if xd != 0 || yd != 0 {
				nx = self.X + xd
				ny = self.Y + yd
				if nx >= 0 && nx < dim.X && ny >= 0 && ny < dim.Y {
					np.X = nx
					np.Y = ny
					if f(np) {
						return
					}
				}
			}
		}
	}
	return
}

func (self pos) sqrDistance(p pos) int {
	dx := self.X - p.X
	dy := self.Y - p.Y
	return dx*dx + dy*dy
}

func (self pos) neighbourTowards(dim pos, target pos) (result *pos) {
	if target == self {
		return nil
	}
	dx := float32(target.X) - float32(self.X)
	dy := float32(target.Y) - float32(self.Y)
	mx := 1
	my := 1
	if dx < 0 {
		mx = -1
	}
	if dy < 0 {
		my = -1
	}
	dx *= float32(mx)
	dy *= float32(my)
	if rand.Float32() < dx/(dx+dy) {
		return &pos{self.X + mx, self.Y}
	}
	return &pos{self.X, self.Y + my}
}

type posUint16Map map[pos]uint16

func (self posUint16Map) MarshalJSON() (result []byte, err error) {
	m := make(map[string]uint16)
	for p, n := range self {
		m[fmt.Sprintf("%v-%v", p.X, p.Y)] = n
	}
	return json.Marshal(m)
}

type posStringMap map[pos]string

func (self posStringMap) MarshalJSON() (result []byte, err error) {
	m := make(map[string]string)
	for p, n := range self {
		m[fmt.Sprintf("%v-%v", p.X, p.Y)] = n
	}
	return json.Marshal(m)
}

type posBoolMap map[pos]bool

func (self posBoolMap) MarshalJSON() (result []byte, err error) {
	m := make(map[string]bool)
	for p, n := range self {
		m[fmt.Sprintf("%v-%v", p.X, p.Y)] = n
	}
	return json.Marshal(m)
}

type posEl struct {
	pos
	next *posEl
	prev *posEl
}
