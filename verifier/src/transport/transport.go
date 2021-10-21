package transport

import (
	//  "time"
	"go.nanomsg.org/mangos/v3"
	"go.nanomsg.org/mangos/v3/protocol/pair"
	"go.nanomsg.org/mangos/v3/protocol/pub"
	"go.nanomsg.org/mangos/v3/protocol/sub"
	_ "go.nanomsg.org/mangos/v3/transport/all"
)

func NewPubSocket(url string) (mangos.Socket, error) {
	sock, err := pub.NewSocket()
	if err != nil {
		return nil, err
	}
	err = sock.Listen(url)
	if err != nil {
		return nil, err
	}
	return sock, nil
}

func NewSubSocket(url string, topic string) (mangos.Socket, error) {
	sock, err := sub.NewSocket()
	if err != nil {
		return nil, err
	}
	err = sock.Dial(url)
	if err != nil {
		return nil, err
	}
	err = sock.SetOption(mangos.OptionSubscribe, []byte(topic))
	if err != nil {
		return nil, err
	}
	return sock, nil
}

func NewPairListenSocket(url string) (mangos.Socket, error) {
	sock, err := pair.NewSocket()
	if err != nil {
		return nil, err
	}
	err = sock.Listen(url)
	if err != nil {
		return nil, err
	}
	return sock, nil
}

func Publish(psock mangos.Socket, msg string) error {
	err := psock.Send([]byte(msg))
	if err != nil {
		return err
	}
	return nil
}

func Receive(rsock mangos.Socket) ([]byte, error) {
	msg, err := rsock.Recv()
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func Send(psock mangos.Socket, bytes []byte) error {
	err := psock.Send(bytes)
	if err != nil {
		return err
	}
	return nil
}
