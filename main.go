package main

import (
	"raft/config"
	"raft/raft"
	"raft/transport"
)

func main() {

	config := config.GetConfig()
	if config == nil {
		return
	}

	recieveAppendEntry := make(chan (raft.AppendEntry), 3)
	recieveAppendEntryReply := make(chan (raft.AppendEntryReply), 3)
	recieveElection := make(chan (raft.Election), 3)
	recieveElectionReply := make(chan (raft.ElectionReply), 3)

	clients := make([]transport.GRPCTransportClient, len(config.Peer_ips))
	server := transport.NewServer(
		recieveAppendEntry, recieveAppendEntryReply, recieveElection, recieveElectionReply,
	)

	for i := range clients {
		clients[i] = transport.NewClient(config.Peer_ips[i])
	}

	go func() {
		server.Start(config.Port)
	}()

	peers := make([]raft.Peer, len(config.Peer_ids))

	for i := range peers {
		peers[i] = raft.Peer{
			Ip:          config.Peer_ips[i],
			CommitIndex: -1, Id: config.Peer_ids[i],
			Transport: clients[i],
		}
	}

	raftServer := raft.New(peers, config.NodeId,
		recieveAppendEntry, recieveAppendEntryReply, recieveElection, recieveElectionReply)

	raftServer.Start()
}
