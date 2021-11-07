package stats

// Stats package:
import (
	"fmt"
	"sort"
	"time"
)

type logTimes []time.Time

type Stats struct {
	NumRecvd    uint64
	NumWrites   uint64
	NumVerified uint64
	TimeStamps  logTimes
}

func (lt *logTimes) Less(i, j int) bool {
	return (*lt)[i].Before((*lt)[j])
}

func (lt *logTimes) Swap(i, j int) {
	(*lt)[i], (*lt)[j] = (*lt)[j], (*lt)[i]
}

func (lt *logTimes) Len() int {
	return len(*lt)
}

func (sts *Stats) GetSimpleTime(t *time.Time) string {
	var hr int
	if t.Hour() == 0 || t.Hour() == 12 {
		hr = 12
	} else {
		hr = t.Hour() % 12
	}
	return fmt.Sprintf("Date & time: %d:%d:%d", hr, t.Minute(), t.Second())
}

func (sts *Stats) getSortedTimes() logTimes {
	sortedTimes := make(logTimes, len(sts.TimeStamps))
	copy(sortedTimes, sts.TimeStamps)
	sort.Sort(&sortedTimes)
	return sortedTimes
}

func (sts *Stats) LogReqTime(c <-chan time.Time) {
	for ts := range c {
		sts.TimeStamps = append(sts.TimeStamps, ts)
	}
}

func (sts *Stats) IncRecvd(c <-chan uint64) {
	sts.NumRecvd = 0
	for inc := range c {
		sts.NumRecvd += inc
		// fmt.Printf("http count stats %d\n", sts.NumRecvd)
	}
}

func (sts *Stats) IncWrite(c <-chan uint64) {
	sts.NumWrites = 0
	for inc := range c {
		sts.NumWrites += inc
	}
}

func (sts *Stats) IncVerfied(c <-chan uint64) {
	sts.NumVerified = 0
	for inc := range c {
		sts.NumVerified += inc
	}
}

func (sts *Stats) String() string {
	sortedTimes := sts.getSortedTimes()
	numTimes := len(sortedTimes)
	if numTimes > 0 {
		numReqs := fmt.Sprintf("number of requests: %d", sts.NumRecvd)
		numWrites := fmt.Sprintf("number of writes: %d", sts.NumWrites)
		numVerified := fmt.Sprintf("number verified: %d", sts.NumVerified)
		firstReq := fmt.Sprintf("received first request at: %s", sts.GetSimpleTime(&sortedTimes[0]))
		lastReq := fmt.Sprintf("recieved last request at: %s", sts.GetSimpleTime((&sortedTimes[numTimes-1])))
		s := "\n*** stats ***\n" +
			numReqs + "\n" +
			numWrites + "\n" +
			numVerified + "\n" +
			firstReq + "\n" +
			lastReq + "\n" +
			"*** end stats ***\n"
		return s
	} else {
		return "\n*** stats ***\nno requests received\n*** end stats ***\n"
	}
}
