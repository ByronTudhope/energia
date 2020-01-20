package connector

import (
	"github.com/kristoiv/hid"
	"fmt"
	"time"
	"errors"
)

type USBConnector struct {
	deviceInfo *hid.DeviceInfo
	device     hid.Device
}

func NewUSBConnector(path string) (uc *USBConnector, err error) {
	deviceInfo, err := hid.ByPath(path)
	if err != nil {
		return nil, err
	}

  c := &USBConnector{deviceInfo: deviceInfo}
  
  err = c.Open()
  if err != nil {
    return nil, err
  }
return c, nil
}

func (uc *USBConnector) DeviceInfo() *hid.DeviceInfo {
	return uc.deviceInfo
}

func (uc *USBConnector) Path() string {
	return uc.deviceInfo.Path
}

func (uc *USBConnector) Open() error {
	// Do nothing if already open
	if uc.device != nil {
		return nil
	}

	device, err := uc.deviceInfo.Open()
	if err != nil {
		return err
	}
	uc.device = device
	return nil
}

func (uc *USBConnector) Close() {
	uc.device.Close()
	uc.device = nil
}

func (uc *USBConnector) ReadUntilCR() ([]byte, error) {
	return uc.Read(0x0d)
}

// TODO This should take timout as argument or set by config
func (uc *USBConnector) Read(terminator byte) ([]byte, error) {
	ch := uc.device.ReadCh()
	bytesRead := make([]byte, 0, 8)
	reading := true
	for reading {
		select {
			case bs := <-ch:
				for _, b := range bs {
					if b > 0 {
						bytesRead = append(bytesRead, b)
					}
					if b == terminator {
						reading = false
					}
				}
			case <-time.After(3 * time.Second):
				fmt.Println("Timeout reading HID")
				reading = false
				return nil, errors.New("Timeout reading HID")
			}
		}
	return bytesRead, nil
}

func (uc *USBConnector) Write(bytes []byte) error {
	return uc.device.Write(bytes)
}
