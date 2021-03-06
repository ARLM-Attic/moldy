package world

type cmdType int

const (
	cmdGetState = iota
	cmdSubscribe
	cmdNewMold
	cmdAddTarget
	cmdClearTargets
)

type cmd struct {
	typ cmdType
	ret chan interface{}
	arg interface{}
}

type Delta struct {
	Created        posStringMap
	Removed        posStringMap
	CreatedTargets map[string]posUint16Map
	RemovedTargets map[string]posUint16Map
}

type target struct {
	pos       pos
	name      string
	precision int
}

type CmdChan chan cmd

func (self CmdChan) send(c cmd) interface{} {
	c.ret = make(chan interface{})
	self <- c
	return <-c.ret
}

func (self CmdChan) AddTarget(name string, precision, x, y int) {
	self.send(cmd{
		typ: cmdAddTarget,
		arg: target{
			pos:       pos{x, y},
			name:      name,
			precision: precision,
		},
	})
}

func (self CmdChan) ClearTargets(name string) {
	self.send(cmd{
		typ: cmdClearTargets,
		arg: name,
	})
}

func (self CmdChan) NewMold(name string) {
	self.send(cmd{
		typ: cmdNewMold,
		arg: name,
	})
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
