package log

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"
)

const (
	KB                  = 1024
	MB                  = 1024 * KB
	bufferSizeConfigKey = "buffer_size"
	DefaultBufferSize   = 1
)

var (
	SEPARATOR = []byte("\r\n")
)

func init() {
	if err := zap.RegisterSink("fluent-bit-tcp", func(url *url.URL) (zap.Sink, error) {
		return newFluentBitTCPSink(url)
	}); err != nil {
		Fatal("RegisterSink error", err)
	}
}

func newTCPConn(address string) (net.Conn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	if err := conn.SetReadDeadline(time.Now()); err != nil {
		return nil, err
	}
	return conn, nil
}
func newFluentBitTCPSink(url *url.URL) (zap.Sink, error) {
	var err error
	//Info("newFluentBitTCPSink", Any("url", url))
	query := url.Query()
	bufferSize := DefaultBufferSize
	bufferSizeString := query.Get(bufferSizeConfigKey)
	if bufferSizeString != "" {
		bufferSize, err = strconv.Atoi(bufferSizeString)
		if err != nil {
			return nil, err
		}
	}
	ws := &fluentBitTCPSink{
		address:    url.Host,
		logC:       make(chan []byte, bufferSize*MB),
		buffer:     bytes.Buffer{},
		stopC:      make(chan struct{}),
		doneC:      make(chan struct{}),
		bufferSize: bufferSize * MB,

		oneByte: make([]byte, 1),
	}
	go ws.serve()
	return ws, nil
}

type fluentBitTCPSink struct {
	address    string
	logC       chan []byte
	buffer     bytes.Buffer
	bufferSize int

	stopC chan struct{}
	doneC chan struct{}

	oneByte []byte
}

func (s *fluentBitTCPSink) bufferFull() bool {
	return s.buffer.Len() > s.bufferSize
}

func (s *fluentBitTCPSink) Write(p []byte) (n int, err error) {
	if len(s.logC) >= s.bufferSize {
		return 0, errors.New("buffer full")
	}
	data := make([]byte, len(p))
	copy(data, p)
	s.logC <- data
	return len(p), nil
}

func (s *fluentBitTCPSink) sendData(doneC chan<- struct{}) {
	if s.buffer.Len() > 0 {
		conn, err := newTCPConn(s.address)
		if err != nil {
			Error("new tcp conn error", err)
			goto DONE
		}
		defer conn.Close()
		// Check server close wait
		// In go 1.7+, zero byte reads return immediately and will never return an error.
		// You must read at least one byte.
		if _, err := conn.Read(s.oneByte); err == io.EOF {
			Error("connection closed", nil)
			goto DONE
		}
		if _, err := conn.Write(s.buffer.Bytes()); err != nil {
			Error("write data error", err)
			goto DONE
		}
		Debug("send success")
		s.buffer.Reset()
	}
DONE:
	doneC <- struct{}{}
}

func (s *fluentBitTCPSink) trySendData() {
	Debug("try send data")
	sendDataTimer := time.NewTimer(5 * time.Second)
	defer sendDataTimer.Stop()
	doneC := make(chan struct{})
	go s.sendData(doneC)
	select {
	case <-doneC:
		break
	case <-sendDataTimer.C:
		Error("send data timeout", nil)
	}
	if s.bufferFull() {
		Debug("buffer reset")
		s.buffer.Reset()
	}
}

func (s *fluentBitTCPSink) serve() {
	//Info("fluentBitTCPSink serving")
	trySendDataTicker := time.NewTicker(5 * time.Second)
	defer trySendDataTicker.Stop()
	for {
		select {
		case <-trySendDataTicker.C:
			s.trySendData()
		case data := <-s.logC:
			s.buffer.Write(data)
			s.buffer.Write(SEPARATOR)
			if s.buffer.Len() > 512*KB {
				s.trySendData()
			}
		case <-s.stopC:
			s.trySendData()
			close(s.doneC)
			close(s.logC)
			return
		}
	}
}

func (s *fluentBitTCPSink) Sync() error {
	return s.Close()
}
func (s *fluentBitTCPSink) Close() error {
	//Info("fluentBitTCPSink closing")
	select {
	case s.stopC <- struct{}{}:
	case <-s.doneC:
		return nil
	}
	<-s.doneC
	return nil
}
