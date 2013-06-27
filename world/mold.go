package world

import (
	"fmt"
	"math/rand"
)

type mold struct {
	world     *world
	Name      string
	bitMap    [Width][Height]*int
	Bits      [moldSize]*pos
	bitsIndex int
	size      int
	Targets   posUint16Map
}

func (self *mold) clearTargets() (result posUint16Map) {
	result = self.Targets
	self.Targets = make(posUint16Map)
	return
}

func (self *mold) addTarget(precision uint16, p pos) {
	self.Targets[p] = precision
}

func (self *mold) set(p pos) (replaced *pos) {
	if self.bitMap[p.X][p.Y] == nil {
		if oldPos := self.Bits[self.bitsIndex]; oldPos != nil {
			replaced = oldPos
			self.bitMap[oldPos.X][oldPos.Y] = nil
		} else {
			self.size++
		}
		self.Bits[self.bitsIndex] = &p
		x := self.bitsIndex
		self.bitMap[p.X][p.Y] = &x
		self.bitsIndex = (self.bitsIndex + 1) % moldSize
	} else {
		panic(fmt.Errorf("Tried to set %v in %v, but it was already set", p, self.Name))
	}
	return
}

func (self *mold) unset(p pos) {
	if index := self.bitMap[p.X][p.Y]; index != nil {
		self.size--
		self.Bits[*index] = nil
		self.bitMap[p.X][p.Y] = nil
	} else {
		panic(fmt.Errorf("Tried to unset %v in %v, but it was not set", p, self.Name))
	}
}

func (self *mold) grow(delta *Delta) {
	done := false
	for _, index := range rand.Perm(moldSize) {
		if p := self.Bits[index]; p != nil {
			p.eachNeighbour(self.world.Dimensions, func(n pos) bool {
				if _, found := self.world.owner(n); !found {
					self.set(n)
					delta.Created[n] = self.Name
					done = true
					return true
				}
				return false
			})
		}
		if done {
			break
		}
	}
}

func (self *mold) moveTowards(delta *Delta, target pos, precision uint16) {
	var bestPos *pos
	var bestDistance int
	for _, index := range rand.Perm(moldSize) {
		if p := self.Bits[index]; p != nil {
			if n := p.neighbourTowards(self.world.Dimensions, target); n != nil {
				if owner, found := self.world.owner(*n); !found || owner != self {
					if dist := n.sqrDistance(target); bestPos == nil || dist < bestDistance {
						bestPos = n
						bestDistance = dist
					}
					precision--
				}
			}
		}
		if precision < 1 {
			break
		}
	}
	if bestPos != nil {
		if owner, found := self.world.owner(*bestPos); found {
			owner.unset(*bestPos)
		}
		if replaced := self.set(*bestPos); replaced != nil {
			delta.Removed[*replaced] = self.Name
		}
		delta.Created[*bestPos] = self.Name
	}
}

func (self *mold) move(delta *Delta) {
	for target, precision := range self.Targets {
		self.moveTowards(delta, target, precision)
	}
}
