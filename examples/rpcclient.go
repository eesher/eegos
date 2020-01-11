package main

import (
	"eegos/cluster"
	"eegos/log"

	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	cluster.Connect("test", "127.0.0.1:1234")

	for {
		time.Sleep(3 * time.Second)
		log.Debug("send: Test.PrintAB")
		cluster.Send("test", "Test.PrintAB")

		log.Debug("call:2: Test.TestArgs 10")
		ret := cluster.Call("test", "Test.TestArgs", 10)
		log.Debug("call:2:Test.TestArgs ret", ret[0])

		log.Debug("call:3: Test.TestString aaaa")
		ret = cluster.Call("test", "Test.TestString", "aaaa")
		log.Debugf("call:3:Test.TestString ret", ret[0])
	}
}
