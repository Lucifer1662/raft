package transport

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	pb "raft/pb/raft"
	"raft/raft"

	"google.golang.org/grpc"
)

type GRPCTransportServer struct {
	client *pb.RaftClient

	recieveAppendEntry      chan (raft.AppendEntry)
	recieveAppendEntryReply chan (raft.AppendEntryReply)

	recieveElection      chan (raft.Election)
	recieveElectionReply chan (raft.ElectionReply)
}

func NewServer(recieveAppendEntry chan (raft.AppendEntry),
	recieveAppendEntryReply chan (raft.AppendEntryReply),
	recieveElection chan (raft.Election),
	recieveElectionReply chan (raft.ElectionReply),
) GRPCTransportServer {
	return GRPCTransportServer{
		recieveAppendEntry:      recieveAppendEntry,
		recieveAppendEntryReply: recieveAppendEntryReply,
		recieveElection:         recieveElection,
		recieveElectionReply:    recieveElectionReply,
	}
}

func (t *GRPCTransportServer) AppendEntry(ctx context.Context, in *pb.AppendEntryRequest) (*pb.Empty, error) {
	entries := make([]raft.Entry, len(in.Entries))

	for i, e := range in.Entries {
		entries[i] = raft.Entry{
			Index:   e.Index,
			Payload: e.Payload,
		}
	}
	t.recieveAppendEntry <- raft.AppendEntry{
		Entries:     entries,
		CommitIndex: in.CommitIndex, Term: in.Term, Nodeid: int(in.NodeId)}
	return &pb.Empty{}, nil
}

func (t *GRPCTransportServer) Election(ctx context.Context, in *pb.ElectionRequest) (*pb.Empty, error) {
	log.Println("Transport: Received Election")
	t.recieveElection <- raft.Election{Term: in.Term, Nodeid: int(in.NodeId)}
	log.Println("Transport: Pushed Election")
	return &pb.Empty{}, nil
}

func (t *GRPCTransportServer) ReplyAppendEntry(_ context.Context, in *pb.AppendEntryReply) (*pb.Empty, error) {
	t.recieveAppendEntryReply <- raft.AppendEntryReply{CurrentIndex: in.CurrentIndex, Nodeid: int(in.NodeId)}
	return &pb.Empty{}, nil
}

func (t *GRPCTransportServer) ReplyElection(_ context.Context, in *pb.ElectionReply) (*pb.Empty, error) {
	t.recieveElectionReply <- raft.ElectionReply{Voted: in.Vote, Term: in.Term}
	return &pb.Empty{}, nil
}

func (t *GRPCTransportServer) Start(port int) {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption = make([]grpc.ServerOption, 0)

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterRaftServer(grpcServer, t)
	grpcServer.Serve(lis)
}
