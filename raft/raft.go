package raft

import (
	"log"
	"math"
	"math/rand"
	"time"

	raftLog "raft/log"
)

type Raft struct {
	peers           []Peer
	nodeid          int
	state           State
	votes           int
	term            int64
	log             raftLog.Log
	votedInLastTerm int64

	electionMinTimeoutDuration time.Duration
	electionMaxTimeoutDuration time.Duration
	electionTimeout            time.Ticker

	heartBeatInterval time.Duration
	heartBeatTimeout  time.Ticker

	stateChanged            chan (State)
	recieveAppendEntry      chan (AppendEntry)
	recieveAppendEntryReply chan (AppendEntryReply)

	recieveElection      chan (Election)
	recieveElectionReply chan (ElectionReply)
}

func New(peers []Peer, nodeid int,
	recieveAppendEntry chan (AppendEntry),
	recieveAppendEntryReply chan (AppendEntryReply),
	recieveElection chan (Election),
	recieveElectionReply chan (ElectionReply),
) Raft {
	return Raft{peers: peers,
		state:                      Candidate,
		nodeid:                     nodeid,
		votedInLastTerm:            -1,
		recieveAppendEntry:         recieveAppendEntry,
		recieveElection:            recieveElection,
		recieveAppendEntryReply:    recieveAppendEntryReply,
		recieveElectionReply:       recieveElectionReply,
		stateChanged:               make(chan State, 10),
		electionMinTimeoutDuration: 3 * time.Second,
		electionMaxTimeoutDuration: 5 * time.Second,
		heartBeatInterval:          2000 * time.Millisecond,
		electionTimeout:            time.Ticker{},
	}
}

func (raft *Raft) BecomeCanditate() {
	log.Println("Become Candidate")
	raft.term += 1
	raft.votedInLastTerm = raft.term
	raft.votes = 1

	for _, p := range raft.peers {
		p.Transport.SendElection(Election{
			Term:   raft.term,
			Nodeid: raft.nodeid,
		})
	}

	//incase there is only 1 node
	raft.checkMajority()
}

func (raft *Raft) BecomeLeader() {
	log.Println("Become Leader")

}

func (raft *Raft) BecomeFollower() {
	log.Println("Become Follower")

}

func (raft *Raft) HandleStateChanged() {
	if raft.state == Candidate {
		raft.BecomeCanditate()
	} else if raft.state == Leader {
		raft.BecomeLeader()
	} else if raft.state == Follower {
		raft.BecomeFollower()
	}
}

func (raft *Raft) electionTimeoutDuration() time.Duration {
	rand.Seed(time.Now().UnixNano())
	dif := raft.electionMaxTimeoutDuration - raft.electionMinTimeoutDuration
	newTimeout := time.Duration(rand.Int63n(dif.Nanoseconds()+1) + raft.electionMinTimeoutDuration.Nanoseconds())
	log.Println(newTimeout)
	return newTimeout
}

func (raft *Raft) resetElectionTimeoutDuration() {
	raft.electionTimeout.Stop()
	raft.electionTimeout = *time.NewTicker(raft.electionTimeoutDuration())
	// raft.electionTimeout.Reset(raft.electionTimeoutDuration())
}

func (raft *Raft) setState(state State) {
	//Candidate state can be reset
	if raft.state == state && state != Candidate {
		return
	}

	//Leader -> Candiate
	if raft.state == Leader && state == Candidate {
		return
	}

	//Follower -> Leader
	if raft.state == Follower && state == Leader {
		return
	}
	log.Printf("Old: %v, New: %v", raft.state, state)
	raft.state = state
	raft.stateChanged <- state
}

func (raft *Raft) HandleAppendEntry(appendEntry AppendEntry) AppendEntryReply {
	log.Println("Handle Append")
	if (raft.state == Leader && raft.term < appendEntry.Term) ||
		raft.state == Candidate ||
		raft.state == Follower {

		raft.setState(Follower)
		raft.term = appendEntry.Term
		log.Println("Reset Election")
		raft.resetElectionTimeoutDuration()
	}

	return AppendEntryReply{
		CurrentIndex: raft.log.CommitIndex(),
	}
}

func (raft *Raft) HandleAppendEntryReply(appendEntryReply AppendEntryReply) {

}

func (raft *Raft) HandleElection(election Election) ElectionReply {
	log.Println("Election Message Receive")
	raft.electionTimeout.Reset(raft.electionTimeoutDuration())

	if raft.votedInLastTerm < election.Term && raft.state == Follower {
		raft.votedInLastTerm = election.Term
		log.Println("Election Reply: True")
		return ElectionReply{Voted: true, Term: raft.votedInLastTerm}
	}

	log.Println("Election Reply: False")
	return ElectionReply{Voted: false, Term: raft.votedInLastTerm}
}

func (raft *Raft) checkMajority() {
	votesRequired := int(math.Floor((float64(len(raft.peers)+1) / 2.0) + 1))
	log.Printf("Voted Required: %d, Votes: %d, Term: %d", votesRequired, raft.votes, raft.term)
	if raft.votes >= votesRequired {
		raft.setState(Leader)
	}
}

func (raft *Raft) HandleElectionReply(electionReply ElectionReply) {
	log.Println("Election Reply Message Receive")

	if electionReply.Term == raft.term && electionReply.Voted {
		raft.votes++
		raft.checkMajority()
	}

}

func (raft *Raft) HeartBeat() {
	if raft.state == Leader {
		for _, p := range raft.peers {
			p.Transport.SendAppendEntry(AppendEntry{
				CommitIndex: raft.log.CommitIndex(),
				Term:        raft.term,
				Nodeid:      raft.nodeid,
			})
		}
	}
}

func (raft *Raft) PeerWithId(id int) *Peer {
	for _, p := range raft.peers {
		if p.Id == id {
			return &p
		}
	}
	return nil
}

func (raft *Raft) Start() {

	raft.setState(Follower)
	raft.electionTimeout = *time.NewTicker((raft.electionTimeoutDuration()))
	raft.heartBeatTimeout = *time.NewTicker(raft.heartBeatInterval)

	for {
		select {
		case <-raft.stateChanged:
			raft.HandleStateChanged()

		case appendEntry := <-raft.recieveAppendEntry:
			raft.PeerWithId(appendEntry.Nodeid).Transport.SendAppendEntryReply(raft.HandleAppendEntry(appendEntry))
		case appendEntryReply := <-raft.recieveAppendEntryReply:
			raft.HandleAppendEntryReply(appendEntryReply)

		case election := <-raft.recieveElection:
			raft.PeerWithId(election.Nodeid).Transport.SendElectionReply(raft.HandleElection(election))
		case electionReply := <-raft.recieveElectionReply:
			raft.HandleElectionReply(electionReply)

		case <-raft.electionTimeout.C:
			log.Println("election timeout")
			if raft.state == Candidate {
				raft.setState(Follower)
			} else if raft.state == Follower {
				raft.setState(Candidate)
			}
			raft.resetElectionTimeoutDuration()

		case <-raft.heartBeatTimeout.C:
			raft.HeartBeat()
		}
	}

}
