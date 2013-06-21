package world

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type posStringMap map[pos]string

func (self posStringMap) MarshalJSON() (result []byte, err error) {
	m := make(map[string]string)
	for p, n := range self {
		m[fmt.Sprintf("%v-%v", p.X, p.Y)] = n
	}
	return json.Marshal(m)
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

type pos struct {
	X uint16
	Y uint16
}

func (self pos) neighbours(w *world) (result []pos) {
	result = make([]pos, 0, 8)
	nx := 0
	ny := 0
	for xd := -1; xd < 2; xd++ {
		for yd := -1; yd < 2; yd++ {
			if xd != 0 || yd != 0 {
				nx = int(self.X) + xd
				ny = int(self.Y) + yd
				if nx >= 0 && nx < int(w.Width) && ny >= 0 && ny < int(w.Height) {
					result = append(result, pos{uint16(nx), uint16(ny)})
				}
			}
		}
	}
	return
}

type posBoolMap map[pos]bool

func (self posBoolMap) MarshalJSON() (result []byte, err error) {
	m := make(map[string]bool)
	for p, n := range self {
		m[fmt.Sprintf("%v-%v", p.X, p.Y)] = n
	}
	return json.Marshal(m)
}

type mold struct {
	world *world
	Name  string
	Bits  posBoolMap
}

func (self *mold) size() uint16 {
	return uint16(len(self.Bits))
}

func (self *mold) set(p pos) {
	self.Bits[p] = true
}

func (self *mold) unset(p pos) {
	delete(self.Bits, p)
}

func (self *mold) get(p pos) bool {
	return self.Bits[p]
}

func (self *mold) grow(delta *Delta) {
	n := rand.Int() % len(self.Bits)
	for bit, _ := range self.Bits {
		n--
		for _, neigh := range bit.neighbours(self.world) {
			if n < 0 && !self.world.hasMold(neigh) {
				delta.Created[neigh] = self.Name
				self.Bits[neigh] = true
				return
			}
		}
	}
}

type world struct {
	Width       uint16
	Height      uint16
	Molds       map[string]*mold
	MaxMoldSize uint16
	cmd         CmdChan
	subscribers map[*Subscriber]bool
}

func New(width, height, maxMoldSize uint16) CmdChan {
	w := &world{
		Molds:       make(map[string]*mold),
		Width:       width,
		Height:      height,
		MaxMoldSize: maxMoldSize,
		cmd:         make(CmdChan),
		subscribers: make(map[*Subscriber]bool),
	}
	w.newMold("test")
	go w.mainLoop()
	return w.cmd
}

func (self *world) hasMold(p pos) bool {
	for _, mold := range self.Molds {
		if mold.get(p) {
			return true
		}
	}
	return false
}

func (self *world) newMold(name string) {
	p := pos{uint16(rand.Int()) % self.Width, uint16(rand.Int()) % self.Height}
	for self.hasMold(p) {
		p = pos{uint16(rand.Int()) % self.Width, uint16(rand.Int()) % self.Height}
	}
	m := &mold{
		world: self,
		Bits:  make(map[pos]bool),
		Name:  name,
	}
	m.set(p)
	self.Molds[name] = m
	return
}

func (self *world) growMolds(delta *Delta) {
	var diff uint16
	for _, mold := range self.Molds {
		if diff = self.MaxMoldSize - mold.size(); diff > 0 {
			if rand.Float32() < float32(diff)/float32(self.MaxMoldSize) {
				mold.grow(delta)
			}
		}
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
