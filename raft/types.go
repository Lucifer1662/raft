package raft

type State int

const (
	Candidate State = iota
	Follower
	Leader
)

type Entry struct {
	Payload string
	Index   int64
}

type AppendEntry struct {
	Entries     []Entry
	CommitIndex int64
	Term        int64
	Nodeid      int
}

type AppendEntryReply struct {
	CurrentIndex int64
	Nodeid       int
}

type Election struct {
	Term   int64
	Nodeid int
}

type ElectionReply struct {
	Voted bool
	Term  int64
}

type Peer struct {
	Ip          string
	CommitIndex int64
	Id          int
	Transport   ITransport
}
