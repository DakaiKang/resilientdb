package main

import (
	"fmt"
	"invoker/config"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	// For handling json objects:
	"encoding/json"
	// Import mangos for ipc with rundb:
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/pair"
	_ "go.nanomsg.org/mangos/v3/transport/all"

	// Import aws-sdk for function invocation:
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var log_channel = make(chan string)

func logResponse(c <-chan string) {
	f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for seq := range c {
		if _, err = f.WriteString(seq); err != nil {
			panic(err)
		}
	}

}

type stats struct {
	invokeCnt int32
	recvd     int32
	startTime time.Time
}

var sts = stats{invokeCnt: 0, recvd: 0}

type invokerSocket struct {
	sock mangos.Socket
}

type serverlessClient struct {
	sess         *session.Session
	lambdaClient *lambda.Lambda
}

func die(format string, v ...interface{}) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func shutDown(msg string) {
	fmt.Printf("RECEIVED: %s from replica process. shutting down\n", msg)
	sts.outputStats()
	os.Exit(0)
}

func (s *stats) outputStats() {
	elapsed := time.Since(s.startTime).Seconds()
	thpt := float64(s.invokeCnt) / elapsed
	log.Printf("Invoke count: %d\n", s.invokeCnt)
	log.Printf("Received from replica: %d\n", s.recvd)
	log.Printf("Ran for: %f\n", elapsed)
	log.Printf("Throughput (invocations per second): %f invocations/sec\n", thpt)
}

func newInvokerSocket(url string, timeLimit int) *invokerSocket {
	var err error
	is := new(invokerSocket)
	is.sock, err = pair.NewSocket()
	if err != nil {
		die("invokerSocket error :: new socket creation failed: %s", err.Error())
	}
	// fmt.Printf("url :: %s\n", url)
	err = is.sock.Listen(url)
	if err != nil {
		die("invokerSocket error :: socket Listen failed: %s", err.Error())
	}
	err = is.sock.SetOption(mangos.OptionRecvDeadline, time.Duration(timeLimit)*time.Millisecond)
	if err != nil {
		die("invokerSocket error :: set option failed: %s", err.Error())
	}
	return is
}

func newServerlessClient(region string) *serverlessClient {
	si := new(serverlessClient)
	si.sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	si.lambdaClient = lambda.New(si.sess, &aws.Config{Region: aws.String(region)})
	//fmt.Printf("serverlessClient :: client and session initialized\n")
	return si
}

func (si *serverlessClient) invoke(i *lambda.InvokeInput, wg *sync.WaitGroup, num int, seq string) {
	defer wg.Done()
	//fmt.Printf("invoking function %d\n", num)

	statusCode := "0"
	tries := 0
	maxRetries := 4
	responseText := ""

	for tries < maxRetries && statusCode == "0" {
		if tries > 0 {
			// fmt.Printf("seq number: ;%v; retrying: %d\n", seq, tries) // debug
			fmt.Printf("try: %d;  code: %s;  text: %s;\n", tries, statusCode, responseText) // debug
		}
		tries++
		//invokeStart := time.Now()
		res, err := si.lambdaClient.Invoke(i)
		//invokeTime := time.Since(invokeStart).Milliseconds()
		//fmt.Printf("invocationt time: %v\n", invokeTime)
		if err != nil {
			//sts.outputStats()
			log.Printf("serverlessClient error :: lambda invocation: %s", err.Error())
			statusCode = "0"
			continue
		}
		var response map[string]interface{}
		json.Unmarshal(res.Payload, &response)

		log_channel <- string(res.Payload) + "\n\n"

		statusCode = fmt.Sprintf("%v", response["veriferStatusCode"])
		responseText = fmt.Sprintf("%v", response["verifierResponse"])
		// if statusCode == "0" {
		// 	fmt.Printf("seq number: ;%v; statusCode: %v; response: %v\n", seq, response["veriferStatusCode"], response["verifierResponse"]) // debug
		// }

	}
	if tries == maxRetries {
		die("--------------------- seq number: ;%v; couldn't make it\n", seq) // debug
	}

	// fmt.Printf("seq number: ;%v; res: %v\n", seq, response) // debug

	// fmt.Printf("fcn result: %v\n", res)
	//fmt.Printf("invoked fcn: %d\n", num)
}

func handleMessage(servClient *serverlessClient, conf *config.InvokerConfig, msg []byte, msgID int) {
	var wg sync.WaitGroup
	var msgJSON map[string]interface{}
	err := json.Unmarshal(msg, &msgJSON)
	if err != nil {
		die("handleMessage error :: unmarshal error: %s", err.Error())
	}
	seq_num := fmt.Sprintf("%v", msgJSON["sequenceNumber"])
	for i := 0; i < (2*conf.F + 1); i++ {
		msgJSON["uuid"] = conf.Uuids[i]
		newMsg, err := json.Marshal(msgJSON)
		if err != nil {
			die("handleMessage error :: marshal error: %s", err.Error())
		}
		invokeInput := lambda.InvokeInput{FunctionName: aws.String(conf.FcnName),
			// InvocationType: aws.String("Event"),
			Payload: newMsg}
		wg.Add(1)
		sts.invokeCnt = atomic.AddInt32(&sts.invokeCnt, int32(1))

		//  invokeStart := time.Now()
		// fmt.Printf(string(newMsg))
		// fmt.Printf("----------------------------")

		go servClient.invoke(&invokeInput, &wg, msgID, string(seq_num))
		//  invokeTime := time.Since(invokeStart).Milliseconds()
		//  fmt.Printf("invocationt time: %v\n", invokeTime)
		// fmt.Printf("seq number: %v;    invoke_cnt: %v\n", msgJSON["sequenceNumber"], sts.invokeCnt) // debug
	}
	wg.Wait()
}

func handleMessageUnderThreshold(msgWg *sync.WaitGroup, servClient *serverlessClient, conf *config.InvokerConfig, msg []byte, msgID int) {
	defer msgWg.Done()
	var wg sync.WaitGroup
	var msgJSON map[string]interface{}
	err := json.Unmarshal(msg, &msgJSON)
	if err != nil {
		die("handleMessage error :: unmarshal error: %s", err.Error())
	}
	seq_num := fmt.Sprintf("%v", msgJSON["sequenceNumber"])
	for i := 0; i < (2*conf.F + 1); i++ {
		msgJSON["uuid"] = conf.Uuids[i]
		newMsg, err := json.Marshal(msgJSON)
		if err != nil {
			die("handleMessage error :: marshal error: %s", err.Error())
		}
		invokeInput := lambda.InvokeInput{FunctionName: aws.String(conf.FcnName),
			InvocationType: aws.String("Event"),
			Payload:        newMsg}
		wg.Add(1)
		sts.invokeCnt = atomic.AddInt32(&sts.invokeCnt, int32(1))
		go servClient.invoke(&invokeInput, &wg, msgID, seq_num)
	}
	wg.Wait()
}

func runInvoker(servClients []*serverlessClient, conf *config.InvokerConfig) {
	var err error
	var msg []byte
	iSock := newInvokerSocket("ipc:///tmp/invoker", 100)
	msgID := 0
	if conf.ThresholdOn {
		var msgWg sync.WaitGroup
		log.Printf("Thresholding is ON\n")
		cnt := uint64(0)
		for {
			if msg, err = iSock.sock.Recv(); err == nil {
				if string(msg) == "END" {
					shutDown(string(msg))
				}
				if msgID == 0 {
					sts.startTime = time.Now()
				}
				msgID++
				sts.recvd++
				if cnt < conf.Threshold {
					msgWg.Add(1)
					go handleMessageUnderThreshold(&msgWg, servClients[0], conf, msg, msgID)
					cnt++
				} else {
					break
				}
			}
		}
		msgWg.Wait()
		sts.outputStats()
	} else {
		for {
			if msg, err = iSock.sock.Recv(); err == nil {
				if string(msg) == "END" {
					shutDown(string(msg))
				}
				if msgID == 0 {
					sts.startTime = time.Now()
				}
				sts.recvd++
				numServClients := len(servClients)
				go handleMessage(servClients[msgID%numServClients], conf, msg, msgID)
				msgID++
				// fmt.Printf("%d\n", msgID)
			}
		}
	}
}

func main() {
	conf, err := config.NewInvokerConfig()
	if err != nil {
		die("config error :: could not get config: %s", err.Error())
	}
	lambda_regions := 13
	servClients := make([]*serverlessClient, lambda_regions)
	regions := [13]string{
		"us-west-1",
		"us-west-2",
		"us-east-2",
		// "us-east-1",
		"ca-central-1",
		"eu-central-1",
		"eu-west-1",
		"eu-west-2",
		"eu-west-3",
		"eu-north-1",
		// "ap-northeast-3",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
	}
	// support for multiple regions possible. new client needed for each region.
	// for new region create new client with region name and add to servClients
	for i := 0; i < len(servClients); i++ {
		servClients[i] = newServerlessClient(regions[i])
	}

	fmt.Printf("starting invoker (event) (synced) (f = %d) (regions size = %d)\n", conf.F, len(servClients))
	// Handle common process-killing signals so we can 'gracefully' shut down:
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go logResponse(log_channel)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		sts.outputStats()
		os.Exit(0)
	}(sigc)
	runInvoker(servClients, conf)
}
