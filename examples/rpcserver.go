package main

import (
	"eegos/cluster"
	"eegos/log"
	//"time"
)

type Test struct {
	a, b int
}

func (this *Test) PrintAB() {
	log.Debug("run func PrintAB()", this.a, this.b)
}

func (this *Test) TestArgs(aa int) int {
	log.Debug("run func TestArgs(aa int)", aa)
	return aa + this.a + this.b
}

func (this *Test) TestString(s string) int {
	log.Debug("run func TestString(s string)", s)
	return this.a + this.b
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	cluster.Open(":1234")
	log.Debug("listen on 1234")
	tmp := &Test{1, 2}
	cluster.Register(tmp)
	cluster.Start()

	select {}
	//c := make(chan bool)
	//<-c
}
