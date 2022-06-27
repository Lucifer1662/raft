package main

import (
	"raft/raftService"
	"time"
)

func main() {

	raft := raftService.New()
	go func() {
		raft.Start()
	}()

	timer := time.NewTicker(time.Second * 2)

	for {
		<-timer.C
		raft.Append()

	}

}
