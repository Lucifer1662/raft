package log

type Log struct {
	entries     []interface{}
	commitIndex int64
}

func (log *Log) Append(entry interface{}) {
	log.entries = append(log.entries, entry)
}

func (log *Log) Length() int {
	return len(log.entries)
}

func (log *Log) RemoveFromHead(amount int) {

}

func (log *Log) CommitIndex() int64 {
	return log.commitIndex
}

func (log *Log) SetCommitIndex(commitIndex int64) {
	log.commitIndex = commitIndex
}

func (log *Log) GetBack(from int) []interface{} {
	return log.entries[from:]
}

func (log *Log) At(i int) interface{} {
	return log.entries[i]
}

func (log *Log) SetFrom(appendEntry.Entries[0].Index, appendEntry.Entries){
	
}
