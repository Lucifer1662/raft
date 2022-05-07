package raft

type ITransport interface {
	SendAppendEntry(AppendEntry)
	SendAppendEntryReply(AppendEntryReply)

	SendElection(Election)
	SendElectionReply(ElectionReply)
}
