package log

type Log struct {
	entries     []string
	commitIndex int64
}

func (log *Log) Append(entry string) {
	log.entries = append(log.entries, entry)
}

func (log *Log) RemoveFromHead(amount int) {

}

func (log *Log) CommitIndex() int64 {
	return log.commitIndex
}

func (log *Log) SetCommitIndex(commitIndex int64) {
	log.commitIndex = commitIndex
}
