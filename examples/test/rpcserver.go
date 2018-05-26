package main

import (
	"../../cluster"
	"log"
)

type Test struct {
	a, b int
}

func (this *Test) PrintAB() {
	log.Println("run func PrintAB()", this.a, this.b)
}

func (this *Test) TestArgs(aa int) int {
	log.Println("run func TestArgs(aa int)", aa)
	return aa + this.a + this.b
}

func (this *Test) TestString(s string) int {
	log.Println("run func TestString(s string)", s)
	return this.a + this.b
}

func main() {
	cluster.Open(":1234")
	log.Println("listen on 1234")
	tmp := &Test{1, 2}
	cluster.Register(tmp)
	for {
	}
}
