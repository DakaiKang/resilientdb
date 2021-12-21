package main

// Still pretty messy. Needs to be cleaned up.
import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	"verifier/stats"
	"verifier/transport"
	"verifier/vconfig"

	// Used for keeping track of requests
	_ "github.com/lib/pq"
	cmap "github.com/orcaman/concurrent-map"
)

var clientMsgChan = make(chan string)

var verified_map = make(map[int]bool)
var last_verified int = 0
var verified_channel = make(chan int)

func checkNext(c <-chan int) {
	for seq := range c {

		// fmt.Printf("just verified:  ;%d\n", seq)
		verified_map[seq] = true
		for verified_map[last_verified] {
			last_verified++
			fmt.Printf("last verified:  ;%d;    map_size:  %d\n", last_verified-1, len(verified_map))
		}
	}

}

type ReadReq struct {
	SeqNum  string   `json:"sequenceNumber"`
	ReadSet []string `json:"readSet"`
}

type WriteReq struct {
	SeqNum          string            `json:"sequenceNumber"`
	ReadSetSnapshot map[string]string `json:"readSet"`
	WriteSet        map[string]string `json:"contracts"`
	Uuid            string            `json:"uuid"`
	Topic           string            `json:"clientId"`
	ClientTs        string            `json:"clientTs"`
}

type WriteReqData struct {
	Uuids    []string
	numRecvd int
}

type StatsReq struct {
	ReqType string `json:"request"`
}

type RequestMap struct {
	f         int
	SeqNumMap interface {
		Set(key string, value interface{})
		Get(key string) (interface{}, bool)
		Has(key string) bool
		Remove(key string)
		RemoveCb(key string, cb cmap.RemoveCb) bool
		Upsert(key string, value interface{}, cb cmap.UpsertCb) (res interface{})
	}
}

type WriteHandler struct {
	db              *sql.DB
	rm              *RequestMap
	recvCntChan     chan uint64
	writeCntChan    chan uint64
	verifiedCntChan chan uint64
	logTimeChan     chan time.Time
	//clientMsgChan chan string
}

type ReadHandler struct {
	db          *sql.DB
	recvCntChan chan uint64
	logTimeChan chan time.Time
}

type StatsHandler struct {
	sts *stats.Stats
}

func handleSig(db *sql.DB, sts *stats.Stats) {
	// Catch any signals that kill the process and shut down  a little more 'gracefully':
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	fmt.Printf("Handle Sig\n")
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("\nCaught signal %s: shutting down.%s\n", sig, sts.String())
		db.Close()
		os.Exit(0)
	}(sigc)
}

func runTimer(start int64, dur time.Duration, db *sql.DB, sts *stats.Stats) {
	timer := time.NewTimer(dur)
	<-timer.C
	fmt.Printf("Timer fired. Stats on timeout: %s\n", sts.String())
	elapsed := time.Now().UnixNano() - start
	fmt.Printf("Ran for %d ns\n", elapsed)
	db.Close()
	os.Exit(0)
}

func sernClientResponseMsg(m string) {

}

func runPublisher(c <-chan string, url string) {
	sock, err := transport.NewPubSocket(url)
	if err != nil {
		fmt.Printf("err on pub sock creation: %s\n", err.Error())
	}
	for msg := range c {
		go func() {
			err = sock.Send([]byte(msg))
			if err != nil {
				fmt.Printf("err on send msg: %s\n", err.Error())
			}
		}()
	}
}

func getHash(wreq *WriteReq) string {
	hasher := sha1.New()
	s := fmt.Sprintf("%s%v", wreq.SeqNum, wreq.WriteSet)
	hasher.Write([]byte(s))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (rm *RequestMap) VerifyWrite(wreq WriteReq) bool {
	hash := getHash(&wreq)
	// fmt.Printf("hash: %s\n", hash)

	upsert := func(exist bool, valueInMap interface{}, newValue interface{}) interface{} {
		nv := newValue.(WriteReqData)
		if exist {
			v := valueInMap.(WriteReqData)
			v.numRecvd += 1
			// v.Uuids = append(v.Uuids, nv.Uuids...)
			return v
		}
		return nv
	}
	// otherwise we record that we have a matching write:
	wd := WriteReqData{numRecvd: 1}
	updated_wd := rm.SeqNumMap.Upsert(hash, wd, upsert).(WriteReqData)
	if updated_wd.numRecvd == rm.f {
		return true
	}
	// fmt.Printf("seqnum ;%s; has been updated to %d.\n", wreq.SeqNum, updated_wd.numRecvd) // debug

	return false
}

func testHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "verifier is up")
}

func (wh *WriteHandler) VerifyReadSnapshot(ReadSetSnapshot map[string]string) bool {
	for key, val := range ReadSetSnapshot {
		stmt := `select balance from accounts where id=$1`
		row := wh.db.QueryRow(stmt, key)
		var expectedVal string
		switch err := row.Scan(&expectedVal); err {
		case sql.ErrNoRows:
			return false
		case nil:
			if expectedVal != val {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// Service Write Request. First verify request and, if verified, follow through with write:
func (wh *WriteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// for name, values := range r.Header {
	// 	// Loop over all values for the name.
	// 	for _, value := range values {
	// 		fmt.Println(name, value)
	// 	}
	// }
	// fmt.Printf("--------------------\n")

	wh.recvCntChan <- 1
	wh.logTimeChan <- time.Now()
	var ws WriteReq
	err := json.NewDecoder(r.Body).Decode(&ws)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)

		fmt.Print(err.Error())
		os.Exit(0)
		return
	}

	// fmt.Printf("recvd wreq: %v\n with batch sz: %d\n", ws, len(ws.WriteSet))
	// fmt.Printf("Connection: %v\n", r.Header.Get("Connection"))

	// TODO STORAGE

	// if !wh.VerifyReadSnapshot(ws.ReadSetSnapshot) {
	// 	fmt.Fprintf(w, "ReadSet Snapshot Invalid")
	// 	return
	// }

	if wh.rm.VerifyWrite(ws) {
		wh.verifiedCntChan <- 1
		vMsg := fmt.Sprintf("%s|{\"sequenceNumber\":%s,\"clientTs\":%s,\"numWrites\":%s}", ws.Topic, ws.SeqNum, ws.ClientTs, strconv.Itoa(len(ws.WriteSet)))
		clientMsgChan <- vMsg
		seq_number, _ := strconv.Atoi(ws.SeqNum)
		verified_channel <- seq_number
		// fmt.Printf("seqnum %s has been verified.\n", ws.SeqNum) // debug
		// for key, val := range ws.WriteSet {
		// 	stmt := `INSERT INTO accounts (id, balance) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET balance=$2`
		// 	_, err = wh.db.Exec(stmt, key, val)
		// 	if err != nil {
		// 		fmt.Fprintf(w, "Write Error, %v", err)
		// 		return
		// 	}
		// 	wh.writeCntChan <- 1
		// }
		// fmt.Fprintf(w, "Write Succeed")
	}
	fmt.Fprintf(w, "%s", "ok")
	r.Body.Close()
	// else {
	// 	fmt.Fprintf(w, "Write not validated")
	// }

}

// Service Read Request:
func (rh *ReadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	rh.recvCntChan <- 1
	rh.logTimeChan <- time.Now()
	var rs ReadReq
	err := json.NewDecoder(r.Body).Decode(&rs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rsMap := make(map[string]string)
	fmt.Printf("recvd reads: %v\n", rs.ReadSet)
	for _, key := range rs.ReadSet {
		var val string
		keyStr := "\"" + key + "\""
		stmt := `select balance from accounts where id=$1`
		row := rh.db.QueryRow(stmt, key)
		switch err := row.Scan(&val); err {
		case sql.ErrNoRows:
			//rsMap[keyStr] = "null"
			// Do Nothing
		case nil:
			rsMap[keyStr] = "\"" + val + "\""
		default:
			fmt.Fprintf(w, "Error: %s", err.Error())
		}
	}
	ret := "{"
	for key, val := range rsMap {
		ret += key + ":" + val + ","
	}
	ret = ret[:len(ret)-1] + "}"
	fmt.Fprintf(w, "%s", ret)
}

// The following isn't safe. They should only be called during completely idle periods:
func (sh *StatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req StatsReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Printf("err on request: %s\n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.ReqType == "reset" {
		sh.sts.NumRecvd = 0
		sh.sts.NumWrites = 0
		fmt.Fprintf(w, "reset")
	} else if req.ReqType == "get" {
		fmt.Printf("reporting stats\n")
		fmt.Fprintf(w, sh.sts.String())
	} else {
		fmt.Fprintf(w, "invalid stats request")
	}
}

func main() {
	// Initialize:
	// Get config:
	conf, err := vconfig.NewVerifierConfig()
	// TODO reevaluate sslmode
	db, err := sql.Open("postgres", "user=expodb dbname=expodb password=123456 sslmode=disable")
	if err != nil {
		log.Fatalf(err.Error())
		log.Fatalf("Database Connection Failure")
		panic(err)
	}

	sts := stats.Stats{NumRecvd: 0, NumWrites: 0, NumVerified: 0}
	rm := RequestMap{f: conf.F, SeqNumMap: cmap.New()}

	// Create Channels:
	recvCntChan := make(chan uint64)
	logTimeChan := make(chan time.Time)
	verifiedCntChan := make(chan uint64)
	writeChanBufferSize := 10000
	writeCntChan := make(chan uint64, writeChanBufferSize)

	// Start Timer:
	dur := time.Duration(conf.Timer)
	simtime := dur * 1000000000 * time.Nanosecond
	go runTimer(time.Now().UnixNano(), simtime, db, &sts)

	go checkNext(verified_channel)
	// Use goroutines that check for a channel and accept changes through those channels:
	//go runPublisher(clientMsgChan, "tcp://172.31.10.42:4000")
	go runPublisher(clientMsgChan, "tcp://"+conf.Vip+":4000")
	go sts.IncRecvd(recvCntChan)
	go sts.LogReqTime(logTimeChan)
	go sts.IncWrite(writeCntChan)
	go sts.IncVerfied(verifiedCntChan)
	handleSig(db, &sts)

	log.Printf("Started <LIMITED>. Awaiting requests (f = %d) (timed/simtime = %d sec)\n", conf.F, conf.Timer)

	// Register Handlers:
	http.HandleFunc("/test", testHandler)
	http.Handle("/write", &WriteHandler{db: db, rm: &rm, recvCntChan: recvCntChan, writeCntChan: writeCntChan, verifiedCntChan: verifiedCntChan, logTimeChan: logTimeChan})
	http.Handle("/read", &ReadHandler{db: db, recvCntChan: recvCntChan, logTimeChan: logTimeChan})
	http.Handle("/stats", &StatsHandler{sts: &sts})
	log.Fatal(http.ListenAndServe(":80", nil))
}
