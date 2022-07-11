package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"v_recv/stats"
	"v_recv/transport"
)

type VerifierMsg struct {
	SeqNum    uint64 `json:"sequenceNumber"`
	ClientTs  uint64 `json:"clientTs"`
	NumWrites uint64 `json:"numWrites"`
}

// Used for debugging:
func (v *VerifierMsg) toString() string {
	return fmt.Sprintf("SeqNum: %d :: NumWrites: %d :: ClientTs: %d", v.SeqNum, v.NumWrites, v.ClientTs)
}

func handleSig(sts *stats.Stats) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down. numWrites: %s\n", sig, strconv.FormatUint(sts.NumWrites, 10))
		os.Exit(0)
	}(sigc)
}

/* Called to collect messages from the verifier @ vURL. Upon receiving message,
 * service registers this message and sends notification to the co-located client.
 */
func collectMessages(vURL string, rURL string, topic string, writeCntChan chan uint64, recvMap *map[uint64]int) {
	subSock, subErr := transport.NewSubSocket(vURL, topic)
	if subErr != nil {
		log.Printf("%s\n", subErr.Error())
	}
	pairSock, pairErr := transport.NewPairListenSocket(rURL)
	if pairErr != nil {
		log.Printf("%s\n", pairErr.Error())
	}
	// Listen for messages:
	for {
		msg, rerr := transport.Receive(subSock)
		if rerr != nil {
			log.Printf("%s\n", rerr.Error())
		}

		// Obtain VerifierMsg:
		// Seperate topic header from message:
		splitMsg := bytes.Split(msg, []byte("|"))
		if topic != string(splitMsg[0]) {
			continue
		}
		// Build vMsg
		vMsg := VerifierMsg{}
		json.Unmarshal(splitMsg[1], &vMsg)

		// Send to runcl:
		tempBytes := make([]byte, 8)
		toSend := make([]byte, 0)
		binary.LittleEndian.PutUint64(tempBytes, vMsg.NumWrites)
		toSend = append(toSend, tempBytes...)
		binary.LittleEndian.PutUint64(tempBytes, vMsg.ClientTs)
		toSend = append(toSend, tempBytes...)
		binary.LittleEndian.PutUint64(tempBytes, vMsg.SeqNum)
		toSend = append(toSend, tempBytes...)
		serr := transport.Send(pairSock, toSend)
		if serr != nil {
			log.Printf("%s\n", serr.Error())
		}
		// fmt.Print(vMsg.toString() + "\n") // debug

		writeCntChan <- vMsg.NumWrites
		/* record the reception of a verifier response message:
		 * can be used for debugging. not strictly necessary
		 */
		if _, exists := (*recvMap)[vMsg.SeqNum]; exists {
			// This should never be more than one:
			(*recvMap)[vMsg.SeqNum] += 1
		} else {
			(*recvMap)[vMsg.SeqNum] = 1
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		log.Println("Usage: ./v_recv <topic>")
		os.Exit(1)
	}
	log.Printf("Listening For msgs from V & Notifying runcl\n")
	topic := os.Args[1]
	sts := stats.Stats{NumRecvd: 0, NumWrites: 0}
	writeCntChan := make(chan uint64)
	go sts.IncWrite(writeCntChan)
	recvMap := make(map[uint64]int)
	// Pretty dumb. shouldn't be hardcoded:
	vURL := "tcp://10.1.0.182:4000"
	rURL := "ipc:///tmp/verifier"
	collectMessages(vURL, rURL, topic, writeCntChan, &recvMap)
}
