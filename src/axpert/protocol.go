package axpert

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/howeyc/crc16"
)

const (
	cr        byte = 0x0d
	lf        byte = 0x0a
	leftParen byte = 0x28
)

func ProtocolId(c Connector) (id string, err error) {
	id, err = sendRequest(c, "QPI")
	return
}

func SerialNo(c Connector) (serialNo string, err error) {
	serialNo, err = sendRequest(c, "QID")
	return
}

type FirmwareVersion struct {
	Series  string
	Version string
}

func InverterFirmwareVersion(c Connector) (version *FirmwareVersion, err error) {
	resp, err := sendRequest(c, "QVFW")
	if err != nil {
		return
	}

	version, err = parseFirmwareVersion(resp, "VERFW")
	return
}

func SCC1FirmwareVersion(c Connector) (version *FirmwareVersion, err error) {
	resp, err := sendRequest(c, "QVFW2")
	if err != nil {
		return
	}

	version, err = parseFirmwareVersion(resp, "VERFW2")
	return
}

func SCC2FirmwareVersion(c Connector) (version *FirmwareVersion, err error) {
	resp, err := sendRequest(c, "QVFW3")
	if err != nil {
		return
	}

	version, err = parseFirmwareVersion(resp, "VERFW3")
	return
}

func SCC3FirmwareVersion(c Connector) (version *FirmwareVersion, err error) {
	resp, err := sendRequest(c, "QVFW4")
	if err != nil {
		return
	}

	version, err = parseFirmwareVersion(resp, "VERFW4")
	return
}

func CVModeChargingTime(c Connector) (chargingTime string, err error) {
	chargingTime, err = sendRequest(c, "QCVT")
	return
}

func ChargingStage(c Connector) (chargingStage string, err error) {
	chargingStage, err = sendRequest(c, "QCST")
	return
}

func DeviceOutputMode(c Connector) (otputMode string, err error) {
	otputMode, err = sendRequest(c, "QOPM")
	return
}

func DSPBootstraped(c Connector) (hasBootstrap string, err error) {
	hasBootstrap, err = sendRequest(c, "QBOOT")
	return
}

func MaxSolarChargingCurrent(c Connector) (charginCurrent string, err error) {
	charginCurrent, err = sendRequest(c, "QMSCHGCR")
	return
}

func MaxUtilityChargingCurrent(c Connector) (charginCurrent string, err error) {
	charginCurrent, err = sendRequest(c, "QMUCHGCR")
	return
}

func MaxTotalChargingCurrent(c Connector) (charginCurrent string, err error) {
	charginCurrent, err = sendRequest(c, "QMCHGCR")
	return
}

func DefaultSettings(c Connector) (defaultSettings string, err error) {
	defaultSettings, err = sendRequest(c, "QDI")
	return
}

//go:generate enumer -type=BatteryType -json
type BatteryType uint8

const (
	AGM BatteryType = iota
	Flooded
	User
)

//go:generate enumer -type=VoltageRange -json
type VoltageRange uint8

const (
	Appliance VoltageRange = iota
	UPS
)

//go:generate enumer -type=OutputSourcePriority -json
type OutputSourcePriority uint8

const (
	OutputUtilityFirst OutputSourcePriority = iota
	OutputSolarFirst
	OutputSBUFirst
)

//go:generate enumer -type=ChargerSourcePriority -json
type ChargerSourcePriority uint8

const (
	ChargerUtilityFirst ChargerSourcePriority = iota
	ChargerSolarFirst
	ChargerSolarAndUtility
	ChargerSolarOnly
)

//go:generate enumer -type=MachineType -json
type MachineType uint8

const (
	GridTie          MachineType = 00
	OffGrid          MachineType = 01
	Hybrid           MachineType = 10
	OffGrid2Trackers MachineType = 11
	OffGrid3Trackers MachineType = 20
)

//go:generate enumer -type=Topology -json
type Topology uint8

const (
	Transfomerless Topology = iota
	Transformer
)

//go:generate enumer -type=OutputMode -json
type OutputMode uint8

const (
	SingleMachine OutputMode = iota
	Parallel
	Phase1
	Phase2
	Phase3
)

//go:generate enumer -type=ParallelPVOK -json
type ParallelPVOK uint8

const (
	AnyInverterConnected ParallelPVOK = iota
	AllInvertersConnected
)

//go:generate enumer -type=PVPowerBalance -json
type PVPowerBalance uint8

const (
	InputCurrentIsChargedCurrent PVPowerBalance = iota
	InputPowerIsChargedPowerPlusLoadPower
)

type RatingInfo struct {
	GridRatingVoltage           float32
	GridRatingCurrent           float32
	ACOutputRatingVoltage       float32
	ACOutputRatingFrequency     float32
	ACOutputRatingCurrent       float32
	ACOutputRatingApparentPower int
	ACOutputRatingActivePower   int
	BatteryRatingVoltage        float32
	BatteryRechargeVoltage      float32
	BatteryUnderVoltage         float32
	BatteryBulkVoltage          float32
	BatteryFloatVoltage         float32
	BatteryType                 BatteryType
	MaxACChargingCurrent        int
	MaxChargingCurrent          int
	InputVoltageRange           VoltageRange
	OutputSourcePriority        OutputSourcePriority
	ChargerSourcePriority       ChargerSourcePriority
	ParallelMaxNumber           int
	MachineType                 MachineType
	Topology                    Topology
	OutputMode                  OutputMode
	BatteryRedischargeVoltage   float32
	ParallelPVOK                ParallelPVOK
	PVPowerBalance              PVPowerBalance
}

func DeviceRatingInfo(c Connector) (ratingInfo *RatingInfo, err error) {
	resp, err := sendRequest(c, "QPIRI")
	if err != nil {
		return
	}

	ratingInfo, err = parseRatingInfo(resp)
	return
}

func sendRequest(c Connector, req string) (resp string, err error) {
	reqBytes := []byte(req)
	reqBytes = append(reqBytes, crc(reqBytes)...)
	reqBytes = append(reqBytes, cr)
	log.Println("Sending ", reqBytes)
	err = c.Write(reqBytes)
	if err != nil {
		return
	}

	readBytes, err := c.ReadUntilCR()
	if err != nil {
		return
	}

	log.Println("Received ", readBytes)
	err = validateResponse(readBytes)
	if err != nil {
		return
	}

	resp = string(readBytes[1 : len(readBytes)-3])
	return
}

func validateResponse(read []byte) error {
	readLen := len(read)
	if read[0] != leftParen {
		return fmt.Errorf("invalid response start %x", read[0])
	}
	if read[readLen-1] != cr {
		return fmt.Errorf("invalid response end %x", read[readLen-1])
	}
	readCrc := read[readLen-3 : readLen-1]
	calcCrc := crc(read[:readLen-3])
	if !bytes.Equal(readCrc, calcCrc) {
		return fmt.Errorf("CRC error, received %v, expected %v", readCrc, calcCrc)
	}

	return nil
}

func crc(data []byte) []byte {
	i := crc16.Checksum(data, crc16.MakeBitsReversedTable(crc16.CCITTFalse))
	bs := []byte{uint8(i >> 8), uint8(i & 0xff)}
	for i := range bs {
		if bs[i] == lf || bs[i] == cr || bs[i] == leftParen {
			bs[i] += 1
		}
	}
	return bs
}

func parseFirmwareVersion(resp string, fwPrefix string) (*FirmwareVersion, error) {
	parts := strings.Split(resp, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid response %s", resp)
	}
	if parts[0] != fwPrefix {
		return nil, fmt.Errorf("invalid prefix %s", parts[0])
	}

	version := strings.Split(parts[1], ".")
	if len(version) != 2 {
		return nil, fmt.Errorf("invalid version %s", parts[1])
	}

	return &FirmwareVersion{version[0], version[1]}, nil
}

func parseRatingInfo(resp string) (*RatingInfo, error) {
	parts := strings.Split(resp, " ")
	if len(parts) < 25 {
		return nil, fmt.Errorf("invalid response %s : not enough fields", resp)
	}

	info := RatingInfo{}

	f, err := strconv.ParseFloat(parts[0], 32)
	if err != nil {
		return nil, err
	}
	info.GridRatingVoltage = float32(f)

	f, err = strconv.ParseFloat(parts[1], 32)
	if err != nil {
		return nil, err
	}
	info.GridRatingCurrent = float32(f)

	f, err = strconv.ParseFloat(parts[2], 32)
	if err != nil {
		return nil, err
	}
	info.ACOutputRatingVoltage = float32(f)

	f, err = strconv.ParseFloat(parts[3], 32)
	if err != nil {
		return nil, err
	}
	info.ACOutputRatingFrequency = float32(f)

	f, err = strconv.ParseFloat(parts[4], 32)
	if err != nil {
		return nil, err
	}
	info.ACOutputRatingCurrent = float32(f)

	i, err := strconv.Atoi(parts[5])
	if err != nil {
		return nil, err
	}
	info.ACOutputRatingApparentPower = i

	i, err = strconv.Atoi(parts[6])
	if err != nil {
		return nil, err
	}
	info.ACOutputRatingActivePower = i

	f, err = strconv.ParseFloat(parts[7], 32)
	if err != nil {
		return nil, err
	}
	info.BatteryRatingVoltage = float32(f)

	f, err = strconv.ParseFloat(parts[8], 32)
	if err != nil {
		return nil, err
	}
	info.BatteryRechargeVoltage = float32(f)

	f, err = strconv.ParseFloat(parts[9], 32)
	if err != nil {
		return nil, err
	}
	info.BatteryUnderVoltage = float32(f)

	f, err = strconv.ParseFloat(parts[10], 32)
	if err != nil {
		return nil, err
	}
	info.BatteryBulkVoltage = float32(f)

	f, err = strconv.ParseFloat(parts[11], 32)
	if err != nil {
		return nil, err
	}
	info.BatteryFloatVoltage = float32(f)

	b, err := strconv.ParseUint(parts[12], 10, 8)
	if err != nil {
		return nil, err
	}
	info.BatteryType = BatteryType(b)

	i, err = strconv.Atoi(parts[13])
	if err != nil {
		return nil, err
	}
	info.MaxACChargingCurrent = i

	i, err = strconv.Atoi(parts[14])
	if err != nil {
		return nil, err
	}
	info.MaxChargingCurrent = i

	b, err = strconv.ParseUint(parts[15], 10, 8)
	if err != nil {
		return nil, err
	}
	info.InputVoltageRange = VoltageRange(b)

	b, err = strconv.ParseUint(parts[16], 10, 8)
	if err != nil {
		return nil, err
	}
	info.OutputSourcePriority = OutputSourcePriority(b)

	b, err = strconv.ParseUint(parts[17], 10, 8)
	if err != nil {
		return nil, err
	}
	info.ChargerSourcePriority = ChargerSourcePriority(b)

	i, err = strconv.Atoi(parts[18])
	if err != nil {
		return nil, err
	}
	info.ParallelMaxNumber = i

	b, err = strconv.ParseUint(parts[19], 10, 8)
	if err != nil {
		return nil, err
	}
	info.MachineType = MachineType(b)

	b, err = strconv.ParseUint(parts[20], 10, 8)
	if err != nil {
		return nil, err
	}
	info.Topology = Topology(b)

	b, err = strconv.ParseUint(parts[21], 10, 8)
	if err != nil {
		return nil, err
	}
	info.OutputMode = OutputMode(b)

	f, err = strconv.ParseFloat(parts[22], 32)
	if err != nil {
		return nil, err
	}
	info.BatteryRedischargeVoltage = float32(f)

	b, err = strconv.ParseUint(parts[23], 10, 8)
	if err != nil {
		return nil, err
	}
	info.ParallelPVOK = ParallelPVOK(b)

	b, err = strconv.ParseUint(parts[24], 10, 8)
	if err != nil {
		return nil, err
	}
	info.PVPowerBalance = PVPowerBalance(b)

	return &info, nil
}
