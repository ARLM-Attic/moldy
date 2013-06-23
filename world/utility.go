package world

import (
	"encoding/json"
	"fmt"
)

type pos struct {
	X uint16
	Y uint16
}

func (self pos) eachNeighbour(dim pos, f func(p pos) bool) {
	nx := 0
	ny := 0
	np := pos{}
	for xd := -1; xd < 2; xd++ {
		for yd := -1; yd < 2; yd++ {
			if xd != 0 || yd != 0 {
				nx = int(self.X) + xd
				ny = int(self.Y) + yd
				if nx >= 0 && nx < int(dim.X) && ny >= 0 && ny < int(dim.Y) {
					np.X = uint16(nx)
					np.Y = uint16(ny)
					if f(np) {
						return
					}
				}
			}
		}
	}
	return
}

func (self pos) distance(p pos) int64 {
	dx := int64(self.X - p.X)
	dy := int64(self.Y - p.Y)
	return dx*dx + dy*dy
}

func (self pos) eachNeighbourTowards(dim pos, target pos, f func(p pos) bool) {
	nx := 0
	ny := 0
	np := pos{}
	minXd := -1
	maxXd := 1
	minYd := -1
	maxYd := 1
	if target.X > self.X {
		minXd = 0
	} else if target.X < self.X {
		maxXd = 0
	} else {
		minXd = 0
		maxXd = 0
	}
	if target.Y > self.Y {
		minYd = 0
	} else if target.Y < self.Y {
		maxYd = 0
	} else {
		minYd = 0
		maxYd = 0
	}
	for xd := minXd; xd < maxXd+1; xd++ {
		for yd := minYd; yd < maxYd+1; yd++ {
			if xd != 0 || yd != 0 {
				nx = int(self.X) + xd
				ny = int(self.Y) + yd
				if nx >= 0 && nx < int(dim.X) && ny >= 0 && ny < int(dim.Y) {
					np.X = uint16(nx)
					np.Y = uint16(ny)
					if f(np) {
						return
					}
				}
			}
		}
	}
	return
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
