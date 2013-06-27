package world

import (
	"fmt"
	"math/rand"
	"testing"
)

func BenchmarkTick(b *testing.B) {
	b.StopTimer()
	w := &world{
		Molds:       make(map[string]*mold),
		Dimensions:  pos{Width, Height},
		cmd:         make(CmdChan),
		subscribers: make(map[*Subscriber]bool),
	}
	for i := 0; i < 10; i++ {
		w.newMold(fmt.Sprintf("test%v", i))
		w.addTarget(fmt.Sprintf("test%v", i), 5, pos{rand.Int() % Width, rand.Int() % Height})
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		w.tick()
	}
}
