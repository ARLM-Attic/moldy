package world

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Delta struct {
	Created posStringMap
	Removed posStringMap
}

type Subscriber func(ev interface{}) error

type CmdChan chan cmd

func (self CmdChan) send(c cmd) interface{} {
	c.ret = make(chan interface{})
	self <- c
	return <-c.ret
}

func (self CmdChan) State() *world {
	return self.send(cmd{
		typ: cmdGetState,
	}).(*world)
}

func (self CmdChan) Subscribe(s Subscriber) {
	self.send(cmd{
		typ: cmdSubscribe,
		arg: s,
	})
}

type cmdType int

const (
	cmdGetState = iota
	cmdSubscribe
)

type cmd struct {
	typ cmdType
	ret chan interface{}
	arg interface{}
}

type mold struct {
	world      *world
	Name       string
	Bits       posBoolMap
	roomy      posBoolMap
	threatened map[pos]uint8
	power      map[pos]uint8
}

func (self *mold) size() uint16 {
	return uint16(len(self.Bits))
}

func (self *mold) set(p pos) {
	if _, found := self.Bits[p]; !found {
		self.Bits[p] = true
		var threat uint8
		p.eachNeighbour(self.world.Dimensions, func(p2 pos) bool {
			self.power[p2]++
			if owner, found := self.world.owner(p2); found && owner != self.Name {
				threat++
			}
			return false
		})
		if threat > 0 {
			self.threatened[p] = threat
		}
		self.world.set(p, self.Name)
		if self.world.hasSpace(p) {
			self.roomy[p] = true
		}
	} else {
		panic(fmt.Errorf("Tried to set %v in %v, but it was already set", p, self.Name))
	}
}

func (self *mold) noSpace(p pos) {
	delete(self.roomy, p)
}

func (self *mold) space(p pos) {
	if self.Bits[p] {
		self.roomy[p] = true
	}
}

func (self *mold) threaten(p pos) {
	if self.Bits[p] {
		self.threatened[p]++
	}
}

func (self *mold) secure(p pos) {
	if pow, found := self.threatened[p]; found {
		if pow == 1 {
			delete(self.threatened, p)
		} else {
			self.threatened[p] = pow - 1
		}
	}
}

func (self *mold) pow(p pos) uint8 {
	return self.power[p]
}

func (self *mold) unset(p pos) {
	if _, found := self.Bits[p]; found {
		delete(self.Bits, p)
		p.eachNeighbour(self.world.Dimensions, func(p2 pos) bool {
			self.power[p2]--
			return false
		})
		delete(self.roomy, p)
		delete(self.threatened, p)
		self.world.unset(p)
	} else {
		panic(fmt.Errorf("Tried to unset %v in %v, but it was never set", p, self.Name))
	}
}

func (self *mold) get(p pos) bool {
	return self.Bits[p]
}

func (self *mold) shrink(delta *Delta) {
	for bit, threat := range self.threatened {
		if self.power[bit] < threat {
			delta.Removed[bit] = self.Name
			self.unset(bit)
		}
	}
}

func (self *mold) grow(delta *Delta, k float32) {
	num := int(k * float32(self.world.MoldGrowth))
	for bit, _ := range self.roomy {
		bit.eachNeighbour(self.world.Dimensions, func(neigh pos) bool {
			if !self.world.hasMold(neigh) {
				delta.Created[neigh] = self.Name
				self.set(neigh)
				return true
			}
			return false
		})
		if num--; num < 1 {
			break
		}
	}
}

type world struct {
	Dimensions  pos
	Molds       map[string]*mold
	MaxMoldSize uint16
	MoldGrowth  uint8
	neighbours  map[pos]uint8
	moldBits    map[pos]string
	cmd         CmdChan
	subscribers map[*Subscriber]bool
}

func New(width, height, maxMoldSize uint16, moldGrowth uint8) CmdChan {
	w := &world{
		Molds:       make(map[string]*mold),
		Dimensions:  pos{width, height},
		MaxMoldSize: maxMoldSize,
		neighbours:  make(map[pos]uint8),
		moldBits:    make(map[pos]string),
		cmd:         make(CmdChan),
		subscribers: make(map[*Subscriber]bool),
		MoldGrowth:  moldGrowth,
	}
	for i := 0; i < 10; i++ {
		w.newMold(fmt.Sprintf("test%v", i))
	}
	go w.mainLoop()
	return w.cmd
}

func (self *world) hasSpace(p pos) bool {
	return self.neighbours[p] < 8
}

func (self *world) owner(p pos) (result string, found bool) {
	result, found = self.moldBits[p]
	return
}

func (self *world) set(p pos, name string) {
	if n, found := self.moldBits[p]; !found {
		self.moldBits[p] = name
		p.eachNeighbour(self.Dimensions, func(p2 pos) bool {
			for _, mold := range self.Molds {
				if mold.Name != name {
					mold.threaten(p2)
				}
			}
			self.neighbours[p2]++
			if self.neighbours[p2] > 7 {
				for _, mold := range self.Molds {
					mold.noSpace(p2)
				}
			}
			return false
		})
	} else {
		panic(fmt.Errorf("Tried to set %v which was already occupied by %v", p, n))
	}
}

func (self *world) unset(p pos) {
	if owner, found := self.moldBits[p]; found {
		delete(self.moldBits, p)
		p.eachNeighbour(self.Dimensions, func(p2 pos) bool {
			for _, mold := range self.Molds {
				if mold.Name != owner {
					mold.secure(p2)
				}
			}
			if self.neighbours[p2] == 8 {
				for _, mold := range self.Molds {
					mold.space(p2)
				}
			}
			self.neighbours[p2]--
			return false
		})
	} else {
		panic(fmt.Errorf("Tried to unset %v which was not occupied by anyone", p))
	}
}

func (self *world) hasMold(p pos) (result bool) {
	_, result = self.moldBits[p]
	return
}

func (self *world) newMold(name string) {
	p := pos{uint16(rand.Int()) % self.Dimensions.X, uint16(rand.Int()) % self.Dimensions.Y}
	for self.hasMold(p) {
		p = pos{uint16(rand.Int()) % self.Dimensions.X, uint16(rand.Int()) % self.Dimensions.Y}
	}
	m := &mold{
		world:      self,
		Bits:       make(posBoolMap),
		Name:       name,
		roomy:      make(posBoolMap),
		threatened: make(map[pos]uint8),
		power:      make(map[pos]uint8),
	}
	m.set(p)
	self.Molds[name] = m
	return
}

func (self *world) growMolds(delta *Delta) {
	var diff uint16
	for _, mold := range self.Molds {
		if diff = self.MaxMoldSize - mold.size(); diff > 0 {
			mold.grow(delta, float32(diff)/float32(self.MaxMoldSize))
		}
	}
}

func (self *world) shrinkMolds(delta *Delta) {
	for _, mold := range self.Molds {
		mold.shrink(delta)
	}
}

func (self *world) emit(ev interface{}) {
	for subscriber, _ := range self.subscribers {
		if err := (*subscriber)(ev); err != nil {
			fmt.Println(err)
			delete(self.subscribers, subscriber)
		}
	}
}

func (self *world) tick() {
	delta := &Delta{
		Created: make(posStringMap),
		Removed: make(posStringMap),
	}
	self.growMolds(delta)
	self.shrinkMolds(delta)
	self.emit(delta)
}

func (self *world) handleCommand(c cmd) {
	switch c.typ {
	case cmdGetState:
		c.ret <- self
	case cmdSubscribe:
		sub := c.arg.(Subscriber)
		self.subscribers[&sub] = true
		c.ret <- nil
	}
}

func (self *world) mainLoop() {
	timer := time.Tick(time.Millisecond * 20)
	for {
		select {
		case <-timer:
			self.tick()
		case c := <-self.cmd:
			self.handleCommand(c)
		}
	}
}
