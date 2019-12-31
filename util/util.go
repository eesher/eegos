package util

import (
	"sync"
)

type Counter struct {
	sync.Mutex
	Num uint16
}

func (this *Counter) GetNum() uint16 {
	this.Lock()
	defer this.Unlock()
	this.Num++
	num := this.Num

	return num
}
