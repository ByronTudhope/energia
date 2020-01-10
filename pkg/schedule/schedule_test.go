package schedule

import (
	"reflect"
	"testing"
	"time"
	_ "time"

	"github.com/mindworks-software/energia/pkg/axpert"
)

func Test_umarshalSchedule(t *testing.T) {
	type args struct {
		msg []byte
	}

	testOne := "{\"DefaultOutputSource\": \"Utility\",\n  \"Entries\": [\n    {\n      \"Timestamp\": \"06:20:00\",\n      \"OutputSource\": \"SUB\"\n    }\n  ]\n}"

	ts, _ := time.Parse("15:04:05", "06:20:00")
	testOneSch := &Schedule{DefaultOutputSource: axpert.Utility, Entries: []Entry{{Timestamp: "06:20:00", OutputSource: axpert.SUB, timeOnly: ts}}}

	tests := []struct {
		name    string
		args    args
		want    *Schedule
		wantErr bool
	}{

		// TODO: Add test cases.
		{name: "Basic unmarshal", args: args{msg: []byte("{\"DefaultOutputSource:\":\"Utility\"}")}, want: &Schedule{DefaultOutputSource: axpert.Utility}, wantErr: false},
		{name: "Simple unmarshal", args: args{msg: []byte(testOne)}, want: testOneSch, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := unmarshalSchedule(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("unmarshalSchedule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("unmarshalSchedule() got = %v, want %v", got, tt.want)
			}
		})
	}
}
