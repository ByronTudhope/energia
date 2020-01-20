package connector

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/goburrow/serial"
)

type SerialConnector struct {
	config  serial.Config
	port    serial.Port
	scanner *bufio.Scanner
}

func NewSerialConnector(config serial.Config) (*SerialConnector, error) {
  c := &SerialConnector{config: config}
  err := c.Open()
  if err != nil{
    return nil ,  err
  }
  return c, nil
}

func (sc *SerialConnector) Config() serial.Config {
	return sc.config
}

func (sc *SerialConnector) Port() serial.Port {
	return sc.port
}

func (sc *SerialConnector) Open() error {
	if sc.port != nil {
		return nil
	}

	port, err := serial.Open(&sc.config)
	if err != nil {
		return err
	}
	sc.port = port

	sc.scanner = bufio.NewScanner(port)
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexByte(data, '\r'); i >= 0 {
			// We have a full message
			return i + 1, data[0 : i+1], nil
		}
		// If we're at EOF, we return whatever we have
		if atEOF {
			return len(data), data, nil
		}
		// Otherwise, request more data.
		return 0, nil, nil
	}
	// Set the split function for the scanning operation.
	sc.scanner.Split(split)

	return nil
}

func (sc *SerialConnector) Close() {
	sc.port.Close()
	sc.port = nil
}

func (sc *SerialConnector) ReadUntilCR() ([]byte, error) {
	return sc.Read(0x0d)
}

func (sc *SerialConnector) Read(terminator byte) ([]byte, error) {
	if sc.scanner.Scan() {
		bytesRead := sc.scanner.Bytes()
		return bytesRead, nil
	}
	if err := sc.scanner.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}

func (sc *SerialConnector) Write(bytes []byte) error {
	n, err := sc.port.Write(bytes)
	if n != len(bytes) {
		return fmt.Errorf("write incomplete, %d of %d written", n, len(bytes))
	}
	return err
}
