package world

import (
	"encoding/json"
	"fmt"
)

type pos struct {
	X uint16
	Y uint16
}

func (self pos) eachNeighbour(w *world, f func(p pos) bool) {
	nx := 0
	ny := 0
	np := pos{}
	for xd := -1; xd < 2; xd++ {
		for yd := -1; yd < 2; yd++ {
			if xd != 0 || yd != 0 {
				nx = int(self.X) + xd
				ny = int(self.Y) + yd
				if nx >= 0 && nx < int(w.Width) && ny >= 0 && ny < int(w.Height) {
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
