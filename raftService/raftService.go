package raftService

import (
	"raft/config"
	"raft/raft"
	"raft/transport"
)

type RaftService struct {
	raftServer raft.Raft
	server     transport.GRPCTransportServer
	port       int
}

func New() *RaftService {

	config := config.GetConfig()
	if config == nil {
		return nil
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

	return &RaftService{
		raftServer: raftServer,
		server:     server,
		port:       config.Port,
	}

}

func (service *RaftService) Start() {
	go func() {
		service.server.Start(service.port)
	}()

	service.raftServer.Start()
}

func (service *RaftService) Append(payload string) {
	service.raftServer.Append(payload)
}
