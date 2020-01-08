package axpert

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestParseFirmwareVersion(t *testing.T) {
	resp := string([]byte{86, 69, 82, 70, 87, 58, 48, 48, 48, 55, 50, 46, 55, 48})

	expectedFv := FirmwareVersion{"00072", "70"}

	fv, err := parseFirmwareVersion(resp, "VERFW")
	bytes, err := json.MarshalIndent(fv, "", "  ")
	fmt.Println(string(bytes))

	if err != nil {
		t.Error("expected no error, got", err)
	}
	if fv == nil {
		t.Error("expected result, got nil")
	}
	if expectedFv != *fv {
		t.Error("expected ", expectedFv, " got ", *fv)
	}
}

func TestParseRatingInfo(t *testing.T) {
	resp := string([]byte{50, 51, 48, 46, 48, 32, 50, 49, 46, 55, 32, 50, 51, 48, 46, 48, 32, 53, 48, 46, 48, 32,
		50, 49, 46, 55, 32, 53, 48, 48, 48, 32, 52, 48, 48, 48, 32, 52, 56, 46, 48, 32, 52, 56, 46, 48, 32,
		52, 55, 46, 53, 32, 53, 51, 46, 50, 32, 53, 49, 46, 57, 32, 50, 32, 51, 48, 32, 49, 50, 48, 32, 48, 32,
		48, 32, 49, 32, 57, 32, 48, 49, 32, 48, 32, 48, 32, 53, 49, 46, 48, 32, 48, 32, 49, 32, 48, 48, 48})

	expectedInfo := RatingInfo{GridRatingVoltage: 230, GridRatingCurrent: 21.7, ACOutputRatingVoltage: 230,
		ACOutputRatingFrequency: 50, ACOutputRatingCurrent: 21.7, ACOutputRatingApparentPower: 5000,
		ACOutputRatingActivePower: 4000, BatteryRatingVoltage: 48, BatteryRechargeVoltage: 48,
		BatteryUnderVoltage: 47.5, BatteryBulkVoltage: 53.2, BatteryFloatVoltage: 51.9, BatteryType: User,
		MaxACChargingCurrent: 30, MaxChargingCurrent: 120, InputVoltageRange: Appliance,
		OutputSourcePriority: Utility, ChargerSourcePriority: ChargerSolarFirst, ParallelMaxNumber: 9,
		MachineType: OffGrid, Topology: Transfomerless, OutputMode: SingleMachine, BatteryRedischargeVoltage: 51,
		ParallelPVOK: AnyInverterConnected, PVPowerBalance: InputPowerIsChargedPowerPlusLoadPower}

	info, err := parseRatingInfo(resp)
	bytes, err := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(bytes))

	if err != nil {
		t.Error("expected no error, got", err)
	}
	if info == nil {
		t.Error("expected result, got nil")
	}
	if expectedInfo != *info {
		t.Error("expected ", expectedInfo, " got ", *info)
	}

}

func TestParseDeviceFlags(t *testing.T) {
	resp := "EABJKLDUVXYZ"

	expectedFlags := map[DeviceFlag]FlagStatus{
		Buzzer:                      FlagEnabled,
		OverloadBypass:              FlagEnabled,
		PowerSaving:                 FlagEnabled,
		DisplayTimeout:              FlagEnabled,
		OverloadRestart:             FlagDisabled,
		OverTemperatureRestart:      FlagDisabled,
		BacklightOn:                 FlagDisabled,
		PrimarySourceInterruptAlarm: FlagDisabled,
		FaultCodeRecord:             FlagDisabled,
		DataLogPopUp:                FlagEnabled,
	}

	flags, err := parseDeviceFlags(resp)
	bytes, err := json.MarshalIndent(flags, "", "  ")
	fmt.Println(string(bytes))

	if err != nil {
		t.Error("expected no error, got", err)
	}
	if flags == nil {
		t.Error("expected result, got nil")
	}
	if !reflect.DeepEqual(expectedFlags, flags) {
		t.Error("expected ", expectedFlags, " got ", flags)
	}

}

func TestParseDeviceFlagsWithLowercaseResponse(t *testing.T) {
	resp := "EabjklDuvxyz"

	expectedFlags := map[DeviceFlag]FlagStatus{
		Buzzer:                      FlagEnabled,
		OverloadBypass:              FlagEnabled,
		PowerSaving:                 FlagEnabled,
		DisplayTimeout:              FlagEnabled,
		OverloadRestart:             FlagDisabled,
		OverTemperatureRestart:      FlagDisabled,
		BacklightOn:                 FlagDisabled,
		PrimarySourceInterruptAlarm: FlagDisabled,
		FaultCodeRecord:             FlagDisabled,
		DataLogPopUp:                FlagEnabled,
	}

	flags, err := parseDeviceFlags(resp)
	bytes, err := json.MarshalIndent(flags, "", "  ")
	fmt.Println(string(bytes))

	if err != nil {
		t.Error("expected no error, got", err)
	}
	if flags == nil {
		t.Error("expected result, got nil")
	}
	if !reflect.DeepEqual(expectedFlags, flags) {
		t.Error("expected ", expectedFlags, " got ", flags)
	}

}

func TestParseDeviceStatusParams(t *testing.T) {
	resp := "230.0 50.0 231.0 49.9 0300 0250 010 460 57.50 012 100 0069 0014 103.8 57.45 00000 00110110 00 07 00856 010"

	expectedParams := DeviceStatusParams{
		GridVoltage:                       230.0,
		GridFrequency:                     50.0,
		ACOutputVoltage:                   231.0,
		ACOutputFrequency:                 49.9,
		ACOutputApparentPower:             300,
		ACOutputActivePower:               250,
		OutputLoadPercent:                 10,
		BusVoltage:                        460,
		BatteryVoltage:                    57.5,
		BatteryChargingCurrent:            12,
		BatteryCapacity:                   100,
		HeatSinkTemperature:               69,
		PVInputCurrent1:                   14,
		PVInputVoltage1:                   103.8,
		BatteryVoltageSCC1:                57.45,
		BatteryDischargeCurrent:           0,
		AddSBUPriorityVersion:             false,
		ConfigStatusChanged:               false,
		SCCFirmwareVersionUpdated:         true,
		LoadOn:                            true,
		BatteryVoltageSteadyWhileCharging: false,
		ChargingOn:                        true,
		SCC1ChargingOn:                    true,
		ACChargingOn:                      false,
		FanBatteryVoltageOffset:           0,
		EEPROMVersion:                     "07",
		PVChargingPower1:                  856,
		FloatingModeCharging:              false,
		SwitchOn:                          true,
	}

	params, err := parseDeviceStatusParams(resp)
	bytes, err := json.MarshalIndent(params, "", "  ")
	fmt.Println(string(bytes))

	if err != nil {
		t.Error("expected no error, got", err)
	}
	if params == nil {
		t.Error("expected result, got nil")
	}
	if expectedParams != *params {
		t.Error("expected ", expectedParams, " got ", *params)
	}
}

func TestParseDeviceStatusParams2(t *testing.T) {
	resp := "0012 105.2 52.5 00840 11000000 0021 0900 0015 100.2 48.48 0790 01890"

	expectedParams := DeviceStatusParams{
		GridVoltage:                       0,
		GridFrequency:                     0,
		ACOutputVoltage:                   0,
		ACOutputFrequency:                 0,
		ACOutputApparentPower:             0,
		ACOutputActivePower:               0,
		OutputLoadPercent:                 0,
		BusVoltage:                        0,
		BatteryVoltage:                    0,
		BatteryChargingCurrent:            0,
		BatteryCapacity:                   0,
		HeatSinkTemperature:               0,
		PVInputCurrent1:                   0,
		PVInputVoltage1:                   0,
		BatteryVoltageSCC1:                0,
		PVInputCurrent2:                   12,
		PVInputVoltage2:                   105.2,
		BatteryVoltageSCC2:                52.5,
		PVInputCurrent3:                   15,
		PVInputVoltage3:                   100.2,
		BatteryVoltageSCC3:                48.48,
		BatteryDischargeCurrent:           0,
		AddSBUPriorityVersion:             false,
		ConfigStatusChanged:               false,
		SCCFirmwareVersionUpdated:         false,
		LoadOn:                            false,
		BatteryVoltageSteadyWhileCharging: false,
		ChargingOn:                        false,
		SCC1ChargingOn:                    false,
		SCC2ChargingOn:                    true,
		SCC3ChargingOn:                    true,
		ACChargingOn:                      false,
		FanBatteryVoltageOffset:           0,
		EEPROMVersion:                     "",
		PVChargingPower1:                  0,
		PVChargingPower2:                  840,
		PVChargingPower3:                  790,
		PVTotalChargingPower:              1890,
		FloatingModeCharging:              false,
		SwitchOn:                          false,
		ACChargingCurrent:                 21,
		ACChargingPower:                   900,
	}

	params, err := parseDeviceStatusParams2(resp, &DeviceStatusParams{})
	bytes, err := json.MarshalIndent(params, "", "  ")
	fmt.Println(string(bytes))

	if err != nil {
		t.Error("expected no error, got", err)
	}
	if params == nil {
		t.Error("expected result, got nil")
	}
	if expectedParams != *params {
		t.Error("expected ", expectedParams, " got ", *params)
	}
}

func TestParseAllStatusParams(t *testing.T) {
	resp1 := "230.0 50.0 231.0 49.9 0300 0250 010 460 57.50 012 100 0069 0014 103.8 57.45 00000 00110110 00 07 00856 010"
	resp2 := "0012 105.2 52.5 00840 11000000 0021 0900 0015 100.2 48.48 0790 01890"

	expectedParams := DeviceStatusParams{
		GridVoltage:                       230.0,
		GridFrequency:                     50.0,
		ACOutputVoltage:                   231.0,
		ACOutputFrequency:                 49.9,
		ACOutputApparentPower:             300,
		ACOutputActivePower:               250,
		OutputLoadPercent:                 10,
		BusVoltage:                        460,
		BatteryVoltage:                    57.5,
		BatteryChargingCurrent:            12,
		BatteryCapacity:                   100,
		HeatSinkTemperature:               69,
		PVInputCurrent1:                   14,
		PVInputVoltage1:                   103.8,
		BatteryVoltageSCC1:                57.45,
		PVInputCurrent2:                   12,
		PVInputVoltage2:                   105.2,
		BatteryVoltageSCC2:                52.5,
		PVInputCurrent3:                   15,
		PVInputVoltage3:                   100.2,
		BatteryVoltageSCC3:                48.48,
		BatteryDischargeCurrent:           0,
		AddSBUPriorityVersion:             false,
		ConfigStatusChanged:               false,
		SCCFirmwareVersionUpdated:         true,
		LoadOn:                            true,
		BatteryVoltageSteadyWhileCharging: false,
		ChargingOn:                        true,
		SCC1ChargingOn:                    true,
		SCC2ChargingOn:                    true,
		SCC3ChargingOn:                    true,
		ACChargingOn:                      false,
		FanBatteryVoltageOffset:           0,
		EEPROMVersion:                     "07",
		PVChargingPower1:                  856,
		PVChargingPower2:                  840,
		PVChargingPower3:                  790,
		PVTotalChargingPower:              1890,
		FloatingModeCharging:              false,
		SwitchOn:                          true,
		ACChargingCurrent:                 21,
		ACChargingPower:                   900,
	}

	params, err := parseDeviceStatusParams(resp1)
	bytes, err := json.MarshalIndent(params, "", "  ")
	fmt.Println(string(bytes))

	if err != nil {
		t.Error("expected no error, got", err)
	}
	if params == nil {
		t.Error("expected result, got nil")
	}

	params, err = parseDeviceStatusParams2(resp2, params)
	bytes, err = json.MarshalIndent(params, "", "  ")
	fmt.Println(string(bytes))

	if err != nil {
		t.Error("expected no error, got", err)
	}
	if params == nil {
		t.Error("expected result, got nil")
	}
	if expectedParams != *params {
		t.Error("expected ", expectedParams, " got ", *params)
	}
}

func TestParseWarningsAllZero(t *testing.T) {
	resp := "00000000000000000000000000000000"

	expected := make([]DeviceWarning, 0)

	warnings, err := parseWarnings(resp)
	bytes, err := json.MarshalIndent(warnings, "", "  ")
	fmt.Println(string(bytes))

	if err != nil {
		t.Error("expected no error, got", err)
	}
	if warnings == nil {
		t.Error("expected result, got nil")
	}

	if !Equal(expected, warnings) {
		t.Error("expected ", expected, " got ", warnings)
	}
}

func TestParseWarnings(t *testing.T) {
	resp := "00010000000000100000000000000000"

	expected := []DeviceWarning{WarnBusUnder, WarnBatteryShutdown}

	warnings, err := parseWarnings(resp)
	bytes, err := json.MarshalIndent(warnings, "", "  ")
	fmt.Println(string(bytes))

	if err != nil {
		t.Error("expected no error, got", err)
	}
	if warnings == nil {
		t.Error("expected result, got nil")
	}

	if !Equal(expected, warnings) {
		t.Error("expected ", expected, " got ", warnings)
	}
}

func Equal(a, b []DeviceWarning) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestFormatEnabledFlags(t *testing.T) {
	flags := []DeviceFlag{
		Buzzer,
		OverloadBypass,
		PowerSaving,
		DisplayTimeout,
		DataLogPopUp,
	}

	expectedEnable := "PEabjkl"

	enable := formatDeviceFlags(flags, FlagEnabled)
	fmt.Println(enable)

	if expectedEnable != enable {
		t.Error("expected ", expectedEnable, " got ", enable)
	}
}

func TestFormatDisabledFlags(t *testing.T) {
	flags := []DeviceFlag{
		OverloadRestart,
		OverTemperatureRestart,
		BacklightOn,
		PrimarySourceInterruptAlarm,
		FaultCodeRecord,
	}

	expectedDisable := "PDuvxyz"

	disable := formatDeviceFlags(flags, FlagDisabled)
	fmt.Println(disable)

	if expectedDisable != disable {
		t.Error("expected ", expectedDisable, " got ", disable)
	}

}
