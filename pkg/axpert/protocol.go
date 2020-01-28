package axpert

import (
	"bytes"
	"fmt"
	"github.com/howeyc/crc16"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/mindworks-software/energia/pkg/connector"
)

const (
	cr        byte = 0x0d
	lf        byte = 0x0a
	leftParen byte = 0x28
)

func ProtocolId(c connector.Connector) (id string, err error) {
	id, err = sendRequest(c, "QPI")
	return
}

func SerialNo(c connector.Connector) (serialNo string, err error) {
	serialNo, err = sendRequest(c, "QID")
	return
}

type FirmwareVersion struct {
	Series  string
	Version string
}

func InverterFirmwareVersion(c connector.Connector) (version *FirmwareVersion, err error) {
	resp, err := sendRequest(c, "QVFW")
	if err != nil {
		return
	}

	version, err = parseFirmwareVersion(resp, "VERFW")
	return
}

func SCC1FirmwareVersion(c connector.Connector) (version *FirmwareVersion, err error) {
	resp, err := sendRequest(c, "QVFW2")
	if err != nil {
		return
	}

	version, err = parseFirmwareVersion(resp, "VERFW2")
	return
}

func SCC2FirmwareVersion(c connector.Connector) (version *FirmwareVersion, err error) {
	resp, err := sendRequest(c, "QVFW3")
	if err != nil {
		return
	}

	version, err = parseFirmwareVersion(resp, "VERFW3")
	return
}

func SCC3FirmwareVersion(c connector.Connector) (version *FirmwareVersion, err error) {
	resp, err := sendRequest(c, "QVFW4")
	if err != nil {
		return
	}

	version, err = parseFirmwareVersion(resp, "VERFW4")
	return
}

func CVModeChargingTime(c connector.Connector) (chargingTime uint8, err error) {
	const query = "QCVT"
	resp, err := sendRequest(c, query)
	if err != nil {
		return
	}
	if resp == "NAK" {
		err = fmt.Errorf("query not supported, %v", query)
		return
	}
	b, err := strconv.ParseUint(resp, 10, 8)
	if err != nil {
		return
	}
	chargingTime = uint8(b)
	return
}

func DeviceChargingStage(c connector.Connector) (chargingStage ChargingStage, err error) {
	const query = "QCST"
	resp, err := sendRequest(c, query)
	if err != nil {
		return
	}
	if resp == "NAK" {
		err = fmt.Errorf("query not supported, %v", query)
		return
	}
	b, err := strconv.ParseUint(resp, 10, 8)
	if err != nil {
		return
	}
	chargingStage = ChargingStage(b)
	return
}

func DeviceOutputMode(c connector.Connector) (outputMode OutputMode, err error) {
	query := "QOPM"
	resp, err := sendRequest(c, query)
	if err != nil {
		return
	}
	if resp == "NAK" {
		err = fmt.Errorf("query not supported, %v", query)
		return
	}
	b, err := strconv.ParseUint(resp, 10, 8)
	if err != nil {
		return
	}
	outputMode = OutputMode(b)
	return
}

func DSPBootstrapped(c connector.Connector) (hasBootstrap bool, err error) {
	bootstrap, err := sendRequest(c, "QBOOT")
	if err != nil {
		return
	}
	hasBootstrap = bootstrap != "0"
	return
}

func MaxSolarChargingCurrent(c connector.Connector) (chargingCurrent string, err error) {
	chargingCurrent, err = sendRequest(c, "QMSCHGCR")
	return
}

func MaxUtilityChargingCurrent(c connector.Connector) (chargingCurrent string, err error) {
	chargingCurrent, err = sendRequest(c, "QMUCHGCR")
	return
}

func MaxTotalChargingCurrent(c connector.Connector) (chargingCurrent string, err error) {
	chargingCurrent, err = sendRequest(c, "QMCHGCR")
	return
}

func DefaultSettings(c connector.Connector) (defaultSettings string, err error) {
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
	Utility OutputSourcePriority = iota
	SUB
	SBU
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

//go:generate enumer -type=ChargingStage -json
type ChargingStage uint8

const (
	Auto ChargingStage = iota
	TwoStage
	ThreeStage
)

//go:generate enumer -type=ParallelPVOK -json -text
type ParallelPVOK uint8

const (
	AnyInverterConnected ParallelPVOK = iota
	AllInvertersConnected
)

//go:generate enumer -type=PVPowerBalance -json -text
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

func DeviceRatingInfo(c connector.Connector) (ratingInfo *RatingInfo, err error) {
	resp, err := sendRequest(c, "QPIRI")
	if err != nil {
		return
	}

	ratingInfo, err = parseRatingInfo(resp)
	return
}

//go:generate enumer -type=FlagStatus -json -text
type FlagStatus byte

const (
	FlagDisabled FlagStatus = iota
	FlagEnabled
)

func (s FlagStatus) char() byte {
	switch s {
	case FlagDisabled:
		return 'D'
	case FlagEnabled:
		return 'E'
	}
	return 0
}

//go:generate enumer -type=DeviceFlag -json -text
type DeviceFlag byte

const (
	Buzzer DeviceFlag = iota
	OverloadBypass
	PowerSaving
	DisplayTimeout
	OverloadRestart
	OverTemperatureRestart
	BacklightOn
	PrimarySourceInterruptAlarm
	FaultCodeRecord
	DataLogPopUp
)

func (f DeviceFlag) char() byte {
	switch f {
	case Buzzer:
		return 'a'
	case OverloadBypass:
		return 'b'
	case PowerSaving:
		return 'j'
	case DisplayTimeout:
		return 'k'
	case OverloadRestart:
		return 'u'
	case OverTemperatureRestart:
		return 'v'
	case BacklightOn:
		return 'x'
	case PrimarySourceInterruptAlarm:
		return 'y'
	case FaultCodeRecord:
		return 'z'
	case DataLogPopUp:
		return 'l'
	}
	return 0
}

func DeviceFlagStatus(c connector.Connector) (flags map[DeviceFlag]FlagStatus, err error) {
	resp, err := sendRequest(c, "QFLAG")
	if err != nil {
		return
	}

	flags, err = parseDeviceFlags(resp)
	return
}

type DeviceStatusParams struct {
	GridVoltage                       float32
	GridFrequency                     float32
	ACOutputVoltage                   float32
	ACOutputFrequency                 float32
	ACOutputApparentPower             int
	ACOutputActivePower               int
	OutputLoadPercent                 int
	BusVoltage                        int
	BatteryVoltage                    float32
	BatteryChargingCurrent            int
	BatteryCapacity                   int
	HeatSinkTemperature               int
	PVInputCurrent1                   int
	PVInputVoltage1                   float32
	BatteryVoltageSCC1                float32
	PVInputCurrent2                   int
	PVInputVoltage2                   float32
	BatteryVoltageSCC2                float32
	PVInputCurrent3                   int
	PVInputVoltage3                   float32
	BatteryVoltageSCC3                float32
	BatteryDischargeCurrent           int
	AddSBUPriorityVersion             bool
	ConfigStatusChanged               bool
	SCCFirmwareVersionUpdated         bool
	LoadOn                            bool
	BatteryVoltageSteadyWhileCharging bool
	ChargingOn                        bool
	SCC1ChargingOn                    bool
	SCC2ChargingOn                    bool
	SCC3ChargingOn                    bool
	ACChargingOn                      bool
	FanBatteryVoltageOffset           int
	EEPROMVersion                     string
	PVChargingPower1                  int
	PVChargingPower2                  int
	PVChargingPower3                  int
	PVTotalChargingPower              int
	FloatingModeCharging              bool
	SwitchOn                          bool
	ACChargingCurrent                 int
	ACChargingPower                   int
}

func DeviceGeneralStatus(c connector.Connector) (params *DeviceStatusParams, err error) {
	resp, err := sendRequest(c, "QPIGS")
	if err != nil {
		return
	}

	params, err = parseDeviceStatusParams(resp)

	return
}

func DeviceGeneralStatus2(c connector.Connector, p *DeviceStatusParams) (params *DeviceStatusParams, err error) {
	resp, err := sendRequest(c, "QPIGS2")
	if err != nil {
		return
	}

	if p != nil {
		params = p
	}

	params, err = parseDeviceStatusParams2(resp, params)

	return
}

//go:generate enumer -type=BatteryStatus -json -text
type BatteryStatus uint8

const (
	BatteryNormal BatteryStatus = iota
	BatteryUnder
	BatteryOpen
)

type ParallelInfo struct {
	DeviceIndex                int
	DeviceExists               bool
	SerialNumber               string
	DeviceMode                 string
	FaultCode                  uint8
	GridVoltage                float32
	GridFrequency              float32
	ACOutputVoltage            float32
	ACOutputFrequency          float32
	ACOutputApparentPower      int
	ACOutputActivePower        int
	OutputLoadPercent          int
	BatteryVoltage             float32
	BatteryChargingCurrent     int
	BatteryCapacity            int
	TotalChargingCurrent       int
	TotalACOutputApparentPower int
	TotalOutputActivePower     int
	TotalACOutputPercent       int
	ACCharging                 bool
	LineLoss                   bool
	LoadOn                     bool
	ConfigurationChanged       bool
	BatteryStatus              BatteryStatus
	OutputMode                 OutputMode
	ChargerSourcePriority      ChargerSourcePriority
	MaxChargerCurrent          int
	MaxChargerRange            int
	MaxACChargerCurrent        int
	BatteryDischargeCurrent    int
	PV1InputCurrent            int
	PV1InputVoltage            float32
	PV1ChargingPower           int
	PV2InputCurrent            int
	PV2InputVoltage            float32
	PV2ChargingPower           int
	PV3InputCurrent            int
	PV3InputVoltage            float32
	PV3ChargingPower           int
	SCC1OK                     bool
	SCC1Charging               bool
	SCC2OK                     bool
	SCC2Charging               bool
	SCC3OK                     bool
	SCC3Charging               bool
}

func ParallelDeviceInfo(c connector.Connector, inverterIndex int) (info *ParallelInfo, err error) {
	resp, err := sendRequest(c, fmt.Sprintf("QPGS%d", inverterIndex))
	if err != nil {
		return
	}
	info, err = parseParallelInfo(resp)
	if err != nil {
		return
	}

	// This always fails/blocks indefinitely/times out
	//resp, err = sendRequest(c, fmt.Sprintf("QP2GS%d", inverterIndex))
	//if err != nil {
	//	return
	//}
	//info, err = parseParallelPVInfo(resp, info)
	//if err != nil {
	//	return
	//}

	info.DeviceIndex = inverterIndex
	return
}

func DeviceMode(c connector.Connector) (mode string, err error) {
	mode, err = sendRequest(c, "QMOD")
	return
}

//go:generate enumer -type=DeviceWarning -json
type DeviceWarning uint8

const (
	WarnReserved DeviceWarning = iota
	WarnInverterFault
	WarnBusOver
	WarnBusUnder
	WarnBusSoftFail
	WarnLineFail
	WarnOPVShort
	WarnInverterVoltageLow
	WarnInverterVoltageHigh
	WarnOverTemperature
	WarnFanLocked
	WarnBatteryVoltageHigh
	WarnBatteryLowAlarm
	WarnReservedOvercharge
	WarnBatteryShutdown
	WarnReservedBatteryDerating
	WarnOverload
	WarnEEPROMFault
	WarnInverterOverCurrent
	WarnInverterSoftFail
	WarnSelfTestFail
	WarnOPDCVoltageOver
	WarnBatteryOpen
	WarnCurrentSensorFail
	WarnBatteryShort
	WarnPowerLimit
	WarnPVVoltageHigh
	WarnMPPTOverloadFault
	WarnMPPTOverloadWarning
	WarnBatteryTooLowToCharge
	WarnPVVoltageHigh2
	WarnMPPTOverloadFault2
	WarnMPPTOverloadWarning2
	WarnBatteryTooLowToCharge2
	WarnPVVoltageHigh3
	WarnMPPTOverloadFault3
	WarnMPPTOverloadWarning3
	WarnBatteryTooLowToCharge3
)

func WarningStatus(c connector.Connector) (warnings []DeviceWarning, err error) {
	status, err := sendRequest(c, "QPIWS")
	if err != nil {
		return
	}

	warnings, err = parseWarnings(status)
	return
}

func EnableDeviceFlags(c connector.Connector, flags []DeviceFlag) error {
	command := formatDeviceFlags(flags, FlagEnabled)
	resp, err := sendRequest(c, command)
	if err != nil {
		return err
	}
	if resp == "NAK" {
		return fmt.Errorf("command not acknowledged, %v", command)
	}
	return nil
}

func DisableDeviceFlags(c connector.Connector, flags []DeviceFlag) error {
	command := formatDeviceFlags(flags, FlagDisabled)
	return sendCommand(c, command)
}

func formatDeviceFlags(flags []DeviceFlag, status FlagStatus) string {
	cmdBuilder := new(strings.Builder)
	cmdBuilder.WriteByte('P')
	cmdBuilder.WriteByte(status.char())
	for _, flag := range flags {
		cmdBuilder.WriteByte(flag.char())
	}

	return cmdBuilder.String()
}

func SetOutputSourcePriority(c connector.Connector, priority OutputSourcePriority) error {
	command := fmt.Sprintf("POP%02d", priority)
	return sendCommand(c, command)
}

func SetDefaultSettings(c connector.Connector) error {
	command := "PF"
	return sendCommand(c, command)
}

func SetMaxTotalChargingCurrent(c connector.Connector, current uint8, parallelNumber uint8) error {
	command := fmt.Sprintf("MCHGC%1d%03d", parallelNumber, current)
	return sendCommand(c, command)
}

func SetParallelMaxTotalChargingCurrent(c connector.Connector, current uint8) error {
	command := fmt.Sprintf("MNCHGC%03d", current)
	return sendCommand(c, command)

}

func SetMaxUtilityChargingCurrent(c connector.Connector, current uint8) error {
	command := fmt.Sprintf("MUCHGC%03d", current)
	return sendCommand(c, command)
}

func SetMaxSolarChargingCurrent(c connector.Connector, current uint8) error {
	command := fmt.Sprintf("MSCHGC%03d", current)
	return sendCommand(c, command)
}

func SetOutputRatingFrequency(c connector.Connector, frequency uint8) error {
	command := fmt.Sprintf("F%02d", frequency)
	return sendCommand(c, command)
}

// Valid values are
// 12V unit: 11V/11.3V/11.5V/11.8V/12V/12.3V/12.5V/12.8V
// 24V unit: 22V/22.5V/23V/23.5V/24V/24.5V/25V/25.5V
// 48V unit: 44V/45V/46V/47V/48V/49V/50V/51V
func SetBatteryRechargeVoltage(c connector.Connector, voltage float32) error {
	command := fmt.Sprintf("PBCV%.1f", voltage)
	return sendCommand(c, command)
}

// Valid values are
// 12V unit: 00.0V/12V/12.3V/12.5V/12.8V/13V/13.3V/13.5V/13.8V/14V/14.3V/14.5
// 24V unit: 00.0V/24V/24.5V/25V/25.5V/26V/26.5V/27V/27.5V/28V/28.5V/29V
// 48V unit: 00.0/V48V/49V/50V/51V/52V/53V/54V/55V/56V/57V/58V
// 00.0V means battery is full(charging in float mode).
func SetBatteryRedischargeVoltage(c connector.Connector, voltage float32) error {
	command := fmt.Sprintf("PBDV%.1f", voltage)
	return sendCommand(c, command)
}

func SetChargerSourcePriority(c connector.Connector, priority ChargerSourcePriority) error {
	command := fmt.Sprintf("PCP%02d", priority)
	return sendCommand(c, command)
}

func SetGridWorkingRange(c connector.Connector, voltageRange VoltageRange) error {
	command := fmt.Sprintf("PGR%02d", voltageRange)
	return sendCommand(c, command)
}

func SetBatteryType(c connector.Connector, batteryType BatteryType) error {
	command := fmt.Sprintf("PBT%02d", batteryType)
	return sendCommand(c, command)
}

func SetDeviceOutputMode(c connector.Connector, mode OutputMode) error {
	command := fmt.Sprintf("POPM%02d", mode)
	return sendCommand(c, command)
}

func SetDeviceOutputVoltage(c connector.Connector, voltage uint8) error {
	command := fmt.Sprintf("POPV%03d", voltage)
	return sendCommand(c, command)
}

func SetParallelChargerSourcePriority(c connector.Connector, priority ChargerSourcePriority, parallelNumber uint8) error {
	command := fmt.Sprintf("PPCP%1d%02d", parallelNumber, priority)
	return sendCommand(c, command)
}

// Valid range is 40.0V ~ 48.0V for 48V unit
func SetBatteryCutoffVoltage(c connector.Connector, voltage float32) error {
	command := fmt.Sprintf("PSDV%.1f", voltage)
	return sendCommand(c, command)
}

// Valid range is 48.0V ~ 58.4V for 48V unit
func SetCVModeChargingVoltage(c connector.Connector, voltage float32) error {
	command := fmt.Sprintf("PCVV%.1f", voltage)
	return sendCommand(c, command)
}

// Valid range is 48.0V ~ 58.4V for 48V unit
func SetFloatChargingVoltage(c connector.Connector, voltage float32) error {
	command := fmt.Sprintf("PBFT%.1f", voltage)
	return sendCommand(c, command)
}

func SetDeviceChargingStage(c connector.Connector, mode OutputMode) error {
	command := fmt.Sprintf("PCST%02d", mode)
	return sendCommand(c, command)
}

// Valid times are
// 0, 10, 20, 40, 60, 90, 120, 150, 180, 210, 240, 255, in minutes
// 255 is a special value that makes the actual time automatically determined
func SetCVModeChargingTime(c connector.Connector, chargingTime uint8) error {
	command := fmt.Sprintf("PCVT%03d", chargingTime)
	return sendCommand(c, command)
}

func SetParallelPVOK(c connector.Connector, pvok ParallelPVOK) error {
	command := fmt.Sprintf("PPVOKC%1d", pvok)
	return sendCommand(c, command)
}

func SetPVPowerBalance(c connector.Connector, balance PVPowerBalance) error {
	command := fmt.Sprintf("PSPB%1d", balance)
	return sendCommand(c, command)
}

func sendCommand(c connector.Connector, command string) error {
	resp, err := sendRequest(c, command)
	if err != nil {
		return err
	}
	if resp == "NAK" {
		return fmt.Errorf("command not acknowledged, %v", command)
	}
	return nil
}

func sendRequest(c connector.Connector, req string) (resp string, err error) {
	defer timeTrack(time.Now(), "timing : "+req)
	log.Println("Sending request", req)
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
	log.Println("Received response: ", resp)
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
	i := crc16.Checksum(data, crc16.CCITTFalseTable)
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

func parseDeviceFlags(resp string) (map[DeviceFlag]FlagStatus, error) {
	flags := make(map[DeviceFlag]FlagStatus)

	if len(resp) < 2 {
		return nil, fmt.Errorf("response too short: %s", resp)
	}
	if strings.HasPrefix(resp, "E") {
		value := FlagEnabled
		for i := 1; i < len(resp); i++ {
			switch resp[i] {
			case 'A', 'a':
				flags[Buzzer] = value
			case 'B', 'b':
				flags[OverloadBypass] = value
			case 'J', 'j':
				flags[PowerSaving] = value
			case 'K', 'k':
				flags[DisplayTimeout] = value
			case 'L', 'l':
				flags[DataLogPopUp] = value
			case 'U', 'u':
				flags[OverloadRestart] = value
			case 'V', 'v':
				flags[OverTemperatureRestart] = value
			case 'X', 'x':
				flags[BacklightOn] = value
			case 'Y', 'y':
				flags[PrimarySourceInterruptAlarm] = value
			case 'Z', 'z':
				flags[FaultCodeRecord] = value
			case 'D':
				value = FlagDisabled
			default:
				fmt.Printf("WARNING: unknown flag %v\n", resp[i])
			}
		}
	}
	return flags, nil
}

func parseDeviceStatusParams(resp string) (*DeviceStatusParams, error) {
	parts := strings.Split(resp, " ")
	if len(parts) < 21 {
		return nil, fmt.Errorf("response too short: %s", resp)
	}

	params := DeviceStatusParams{}

	f, err := strconv.ParseFloat(parts[0], 32)
	if err != nil {
		return nil, err
	}
	params.GridVoltage = float32(f)

	f, err = strconv.ParseFloat(parts[1], 32)
	if err != nil {
		return nil, err
	}
	params.GridFrequency = float32(f)

	f, err = strconv.ParseFloat(parts[2], 32)
	if err != nil {
		return nil, err
	}
	params.ACOutputVoltage = float32(f)

	f, err = strconv.ParseFloat(parts[3], 32)
	if err != nil {
		return nil, err
	}
	params.ACOutputFrequency = float32(f)

	i, err := strconv.Atoi(parts[4])
	if err != nil {
		return nil, err
	}
	params.ACOutputApparentPower = i

	i, err = strconv.Atoi(parts[5])
	if err != nil {
		return nil, err
	}
	params.ACOutputActivePower = i

	i, err = strconv.Atoi(parts[6])
	if err != nil {
		return nil, err
	}
	params.OutputLoadPercent = i

	i, err = strconv.Atoi(parts[7])
	if err != nil {
		return nil, err
	}
	params.BusVoltage = i

	f, err = strconv.ParseFloat(parts[8], 32)
	if err != nil {
		return nil, err
	}
	params.BatteryVoltage = float32(f)

	i, err = strconv.Atoi(parts[9])
	if err != nil {
		return nil, err
	}
	params.BatteryChargingCurrent = i

	i, err = strconv.Atoi(parts[10])
	if err != nil {
		return nil, err
	}
	params.BatteryCapacity = i

	i, err = strconv.Atoi(parts[11])
	if err != nil {
		return nil, err
	}
	params.HeatSinkTemperature = i

	i, err = strconv.Atoi(parts[12])
	if err != nil {
		return nil, err
	}
	params.PVInputCurrent1 = i

	f, err = strconv.ParseFloat(parts[13], 32)
	if err != nil {
		return nil, err
	}
	params.PVInputVoltage1 = float32(f)

	f, err = strconv.ParseFloat(parts[14], 32)
	if err != nil {
		return nil, err
	}
	params.BatteryVoltageSCC1 = float32(f)

	i, err = strconv.Atoi(parts[15])
	if err != nil {
		return nil, err
	}
	params.BatteryDischargeCurrent = i

	b, err := strconv.ParseUint(parts[16], 2, 8)
	if err != nil {
		return nil, err
	}

	sflags := uint8(b)
	params.AddSBUPriorityVersion = sflags&0x80 == 0x80
	params.ConfigStatusChanged = sflags&0x40 == 0x40
	params.SCCFirmwareVersionUpdated = sflags&0x20 == 0x20
	params.LoadOn = sflags&0x10 == 0x10
	params.BatteryVoltageSteadyWhileCharging = sflags&0x08 == 0x08
	params.ChargingOn = sflags&0x04 == 0x04
	params.SCC1ChargingOn = sflags&0x02 == 0x02
	params.ACChargingOn = sflags&0x01 == 0x01

	i, err = strconv.Atoi(parts[17])
	if err != nil {
		return nil, err
	}
	params.FanBatteryVoltageOffset = i

	params.EEPROMVersion = parts[18]

	i, err = strconv.Atoi(parts[19])
	if err != nil {
		return nil, err
	}
	params.PVChargingPower1 = i

	if len(parts[20]) < 3 {
		return nil, fmt.Errorf("invalid status %s", parts[20])
	}
	params.FloatingModeCharging = parts[20][0] == '1'
	params.SwitchOn = parts[20][1] == '1'

	return &params, nil
}

func parseDeviceStatusParams2(resp string, params *DeviceStatusParams) (*DeviceStatusParams, error) {
	parts := strings.Split(resp, " ")
	if len(parts) < 12 {
		return params, fmt.Errorf("response too short: %s", resp)
	}

	i, err := strconv.Atoi(parts[0])
	if err != nil {
		return params, err
	}
	params.PVInputCurrent2 = i

	f, err := strconv.ParseFloat(parts[1], 32)
	if err != nil {
		return params, err
	}
	params.PVInputVoltage2 = float32(f)

	f, err = strconv.ParseFloat(parts[2], 32)
	if err != nil {
		return params, err
	}
	params.BatteryVoltageSCC2 = float32(f)

	i, err = strconv.Atoi(parts[3])
	if err != nil {
		return params, err
	}
	params.PVChargingPower2 = i

	b, err := strconv.ParseUint(parts[4], 2, 8)
	if err != nil {
		return params, err
	}

	sflags := uint8(b)
	params.SCC2ChargingOn = sflags&0x80 == 0x80
	params.SCC3ChargingOn = sflags&0x40 == 0x40

	i, err = strconv.Atoi(parts[5])
	if err != nil {
		return params, err
	}
	params.ACChargingCurrent = i

	i, err = strconv.Atoi(parts[6])
	if err != nil {
		return params, err
	}
	params.ACChargingPower = i

	i, err = strconv.Atoi(parts[7])
	if err != nil {
		return params, err
	}
	params.PVInputCurrent3 = i

	f, err = strconv.ParseFloat(parts[8], 32)
	if err != nil {
		return params, err
	}
	params.PVInputVoltage3 = float32(f)

	f, err = strconv.ParseFloat(parts[9], 32)
	if err != nil {
		return params, err
	}
	params.BatteryVoltageSCC3 = float32(f)

	i, err = strconv.Atoi(parts[10])
	if err != nil {
		return params, err
	}
	params.PVChargingPower3 = i

	i, err = strconv.Atoi(parts[11])
	if err != nil {
		return params, err
	}
	params.PVTotalChargingPower = i

	return params, nil
}

func parseWarnings(status string) ([]DeviceWarning, error) {
	if len(status) < 32 {
		return nil, fmt.Errorf("not enough status flags, %d", len(status))
	}

	if len(status) > 38 {
		return nil, fmt.Errorf("too many status flags, %d", len(status))
	}

	warnings := make([]DeviceWarning, 0)
	for i, c := range status {
		switch c {
		case '1':
			warnings = append(warnings, DeviceWarning(i))
		default:
			continue
		}
	}

	return warnings, nil
}

func parseParallelInfo(resp string) (*ParallelInfo, error) {
	parts := strings.Split(resp, " ")
	if len(parts) < 27 {
		return nil, fmt.Errorf("response too short: %s", resp)
	}

	info := ParallelInfo{}

	b, err := strconv.ParseUint(parts[0], 10, 8)
	if err != nil {
		return nil, err
	}
	info.DeviceExists = b == 1

	info.SerialNumber = parts[1]

	info.DeviceMode = parts[2]

	b, err = strconv.ParseUint(parts[3], 10, 8)
	if err != nil {
		return nil, err
	}
	info.FaultCode = uint8(b)

	f, err := strconv.ParseFloat(parts[4], 32)
	if err != nil {
		return nil, err
	}
	info.GridVoltage = float32(f)

	f, err = strconv.ParseFloat(parts[5], 32)
	if err != nil {
		return nil, err
	}
	info.GridFrequency = float32(f)

	f, err = strconv.ParseFloat(parts[6], 32)
	if err != nil {
		return nil, err
	}
	info.ACOutputVoltage = float32(f)

	f, err = strconv.ParseFloat(parts[7], 32)
	if err != nil {
		return nil, err
	}
	info.ACOutputFrequency = float32(f)

	i, err := strconv.Atoi(parts[8])
	if err != nil {
		return nil, err
	}
	info.ACOutputApparentPower = i

	i, err = strconv.Atoi(parts[9])
	if err != nil {
		return nil, err
	}
	info.ACOutputActivePower = i

	i, err = strconv.Atoi(parts[10])
	if err != nil {
		return nil, err
	}
	info.OutputLoadPercent = i

	f, err = strconv.ParseFloat(parts[11], 32)
	if err != nil {
		return nil, err
	}
	info.BatteryVoltage = float32(f)

	i, err = strconv.Atoi(parts[12])
	if err != nil {
		return nil, err
	}
	info.BatteryChargingCurrent = i

	i, err = strconv.Atoi(parts[13])
	if err != nil {
		return nil, err
	}
	info.BatteryCapacity = i

	f, err = strconv.ParseFloat(parts[14], 32)
	if err != nil {
		return nil, err
	}
	info.PV1InputVoltage = float32(f)

	i, err = strconv.Atoi(parts[15])
	if err != nil {
		return nil, err
	}
	info.TotalChargingCurrent = i

	i, err = strconv.Atoi(parts[16])
	if err != nil {
		return nil, err
	}
	info.TotalACOutputApparentPower = i

	i, err = strconv.Atoi(parts[17])
	if err != nil {
		return nil, err
	}
	info.TotalOutputActivePower = i

	i, err = strconv.Atoi(parts[18])
	if err != nil {
		return nil, err
	}
	info.TotalACOutputPercent = i

	b, err = strconv.ParseUint(parts[19], 2, 8)
	if err != nil {
		return nil, err
	}

	sflags := uint8(b)
	info.SCC1OK = sflags&0x80 == 0x80
	info.ACCharging = sflags&0x40 == 0x40
	info.SCC1Charging = sflags&0x20 == 0x20

	info.LineLoss = sflags&0x04 == 0x04
	info.LoadOn = sflags&0x02 == 0x02
	info.ConfigurationChanged = sflags&0x01 == 0x01

	bs := sflags >> 3 & 0x03
	info.BatteryStatus = BatteryStatus(bs)

	b, err = strconv.ParseUint(parts[20], 10, 8)
	if err != nil {
		return nil, err
	}
	info.OutputMode = OutputMode(b)

	b, err = strconv.ParseUint(parts[21], 10, 8)
	if err != nil {
		return nil, err
	}
	info.ChargerSourcePriority = ChargerSourcePriority(b)

	i, err = strconv.Atoi(parts[22])
	if err != nil {
		return nil, err
	}
	info.MaxChargerCurrent = i

	i, err = strconv.Atoi(parts[23])
	if err != nil {
		return nil, err
	}
	info.MaxChargerRange = i

	i, err = strconv.Atoi(parts[24])
	if err != nil {
		return nil, err
	}
	info.MaxACChargerCurrent = i

	i, err = strconv.Atoi(parts[25])
	if err != nil {
		return nil, err
	}
	info.PV1InputCurrent = i

	i, err = strconv.Atoi(parts[26])
	if err != nil {
		return nil, err
	}
	info.BatteryDischargeCurrent = i

	return &info, nil

}

func parseParallelPVInfo(resp string, info *ParallelInfo) (*ParallelInfo, error) {
	parts := strings.Split(resp, " ")
	if len(parts) < 9 {
		return info, fmt.Errorf("response too short: %s", resp)
	}

	i, err := strconv.Atoi(parts[1])
	if err != nil {
		return info, err
	}
	info.PV1ChargingPower = i

	// TODO: Someday, maybe, parse PV2 & PV3 values

	return info, nil
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}
