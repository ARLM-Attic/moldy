package world

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	Width    = 800
	Height   = 600
	moldSize = 1000
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Subscriber func(ev interface{}) error

type world struct {
	Dimensions  pos
	Molds       map[string]*mold
	cmd         CmdChan
	subscribers map[*Subscriber]bool
}

func New() CmdChan {
	w := &world{
		Molds:       make(map[string]*mold),
		Dimensions:  pos{Width, Height},
		cmd:         make(CmdChan),
		subscribers: make(map[*Subscriber]bool),
	}
	go w.mainLoop()
	return w.cmd
}

func (self *world) owner(p pos) (*mold, bool) {
	for _, mold := range self.Molds {
		if mold.bitMap[p.X][p.Y] != nil {
			return mold, true
		}
	}
	return nil, false
}

func (self *world) rand() pos {
	return pos{rand.Int() % Width, rand.Int() % Height}
}

func (self *world) newMold(name string) {
	p := self.rand()
	for _, found := self.owner(p); found; _, found = self.owner(p) {
		p = self.rand()
	}
	m := &mold{
		world:   self,
		Name:    name,
		Targets: make(posUint16Map),
	}
	m.set(p)
	self.Molds[name] = m
	return
}

func (self *world) moveMolds(delta *Delta) {
	for _, mold := range self.Molds {
		mold.move(delta)
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

func (self *world) clearTargets(name string) {
	delta := &Delta{
		RemovedTargets: map[string]posUint16Map{
			name: self.Molds[name].clearTargets(),
		},
	}
	self.emit(delta)
}

func (self *world) addTarget(name string, precision uint16, p pos) {
	self.Molds[name].addTarget(precision, p)
	delta := &Delta{
		CreatedTargets: map[string]posUint16Map{
			name: posUint16Map{
				p: precision,
			},
		},
	}
	self.emit(delta)
}

func (self *world) growMolds(delta *Delta) {
	diff := 0
	for _, mold := range self.Molds {
		if diff = moldSize - mold.size; diff > 0 {
			if rand.Float32() < float32(diff)/moldSize {
				mold.grow(delta)
			}
		}
	}
}

func (self *world) tick() {
	delta := &Delta{
		Created: make(posStringMap),
		Removed: make(posStringMap),
	}
	self.growMolds(delta)
	self.moveMolds(delta)
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
	case cmdNewMold:
		name := c.arg.(string)
		self.newMold(name)
		c.ret <- nil
	case cmdAddTarget:
		t := c.arg.(target)
		self.addTarget(t.name, uint16(t.precision), t.pos)
		c.ret <- nil
	case cmdClearTargets:
		name := c.arg.(string)
		self.clearTargets(name)
		c.ret <- nil
	}
}

func (self *world) mainLoop() {
	timer := time.Tick(time.Millisecond * 5)
	for {
		select {
		case <-timer:
			self.tick()
		case c := <-self.cmd:
			self.handleCommand(c)
		}
	}
}
