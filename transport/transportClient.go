package transport

import (
	"context"
	"log"
	pb "raft/pb/raft"
	"raft/raft"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCTransportClient struct {
	ip string
}

func NewClient(ip string) GRPCTransportClient {
	return GRPCTransportClient{ip: ip}
}

func DeadLineContext() context.Context {
	clientDeadline := time.Now().Add(time.Second * 2)
	ctx, _ := context.WithDeadline(context.Background(), clientDeadline)
	return ctx
}

func (t GRPCTransportClient) SendAppendEntry(r raft.AppendEntry) {

	go func() {
		conn, err := grpc.Dial(t.ip, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return
		}
		defer conn.Close()
		client := pb.NewRaftClient(conn)

		entries := make([]*pb.Entry, len(r.Entries))

		for i, e := range r.Entries {
			entries[i] = &pb.Entry{
				Index:   e.Index,
				Payload: e.Payload,
			}
		}

		m := pb.AppendEntryRequest{
			NodeId:      int32(r.Nodeid),
			Entries:     entries,
			Term:        r.Term,
			CommitIndex: r.CommitIndex,
		}

		log.Println("Transport: Start Sending Heartbeat Message")
		client.AppendEntry(DeadLineContext(), &m, grpc.FailFastCallOption{})
		log.Println("Transport: Finished Sending Heartbeat Message")
	}()

}

func (t GRPCTransportClient) SendElection(e raft.Election) {
	go func() {

		conn, err := grpc.Dial(t.ip, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Println(err.Error())
			return
		}

		defer conn.Close()
		client := pb.NewRaftClient(conn)

		m := pb.ElectionRequest{Term: e.Term, NodeId: int32(e.Nodeid)}
		log.Println("Transport: Send Election Message")
		_, err = client.Election(DeadLineContext(), &m)
		if err != nil {
			log.Println(err.Error())
			return
		}
	}()
}

func (t GRPCTransportClient) SendAppendEntryReply(r raft.AppendEntryReply) {
	go func() {

		conn, err := grpc.Dial(t.ip, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return
		}
		defer conn.Close()
		client := pb.NewRaftClient(conn)

		m := pb.AppendEntryReply{CurrentIndex: r.CurrentIndex, NodeId: int32(r.Nodeid)}

		client.ReplyAppendEntry(DeadLineContext(), &m)
	}()

}

func (t GRPCTransportClient) SendElectionReply(r raft.ElectionReply) {
	go func() {
		log.Println("Transport: Send Election Reply Message")
		conn, err := grpc.Dial(t.ip, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return
		}
		defer conn.Close()
		client := pb.NewRaftClient(conn)

		m := pb.ElectionReply{Vote: r.Voted, Term: r.Term}
		client.ReplyElection(DeadLineContext(), &m)
		if err != nil {
			log.Println(err.Error())
			return
		}
	}()
}
