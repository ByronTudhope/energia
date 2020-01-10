package axpert

import (
	"bytes"
	"testing"
)

func TestCrc(t *testing.T) {
	data := "(NAK"
	expectedCrc := []byte{0x73, 0x73}
	crc := crc([]byte(data))

	if !bytes.Equal(expectedCrc, crc) {
		t.Error("Expected ", expectedCrc, "got ", crc)
	}
}

func TestValidateResponse(t *testing.T) {
	data := "(NAKss\r"
	err := validateResponse([]byte(data))

	if err != nil {
		t.Error("Expected no error", "got ", err)
	}
}

func TestIssue13Crc(t *testing.T) {
	data := []byte{40, 49, 32, 57, 50, 57, 51, 49, 57, 48, 55, 49, 48, 49, 53, 56, 52, 32, 66, 32, 48, 48, 32, 50, 51,
		52, 46, 51, 32, 52, 57, 46, 57, 49, 32, 50, 51, 48, 46, 49, 32, 52, 57, 46, 57, 48, 32, 49, 51, 53, 55, 32, 49,
		50, 52, 52, 32, 48, 50, 55, 32, 53, 48, 46, 53, 32, 48, 48, 48, 32, 48, 57, 53, 32, 49, 48, 55, 46, 50, 32, 48,
		48, 48, 32, 48, 49, 51, 53, 55, 32, 48, 49, 50, 52, 52, 32, 48, 50, 52, 32, 49, 48, 49, 48, 48, 48, 49, 48, 32,
		48, 32, 49, 32, 48, 54, 48, 32, 49, 52, 48, 32, 51, 48, 32, 50, 52, 32, 48, 48, 49, 208, 13}

	err := validateResponse(data)

	if err == nil {
		t.Error("Expected an error", "got ", err)
	}

}

func TestExampleCrc(t *testing.T) {
	data := []byte{40, 50, 52, 48, 46, 49, 32, 52, 57, 46, 57, 32, 50, 52, 48, 46, 49, 32, 52, 57, 46, 57, 32, 48, 50,
		52, 48, 32, 48, 49, 56, 53, 32, 48, 48, 52, 32, 52, 51, 53, 32, 53, 52, 46, 48, 48, 32, 48, 48, 48, 32, 49, 48,
		48, 32, 48, 48, 53, 53, 32, 48, 48, 48, 48, 32, 48, 48, 48, 46, 48, 32, 48, 48, 46, 48, 48, 32, 48, 48, 48, 48,
		48, 32, 48, 48, 48, 49, 48, 49, 48, 49, 32, 48, 48, 32, 48, 48, 32, 48, 48, 48, 48, 48, 32, 49, 49, 48}

	expectedCrc := []byte{219, 14}
	crc := crc(data)

	if !bytes.Equal(expectedCrc, crc) {
		t.Error("Expected ", expectedCrc, "got ", crc)
	}
}
