package schedule

import (
	"reflect"
	"testing"
	"time"
)

func Test_umarshalSchedule(t *testing.T) {
	type args struct {
		msg []byte
	}

	testOne := "{\"DefaultOutputSource\": 0,\n  \"Entries\": [\n    {\n      \"Timestamp\": \"2020-01-08T06:20:00-02:00\",\n      \"OutputSource\": 1\n    }\n  ]\n}"
	time, _ := time.Parse(time.RFC3339, "2020-01-08T06:20:00-02:00")

	testOneSch := &Schedule{DefaultOutputSource: 0, Entries: []Entry{{Timestamp: time, OutputSource: 1}}}

	tests := []struct {
		name    string
		args    args
		want    *Schedule
		wantErr bool
	}{

		// TODO: Add test cases.
		{name: "Basic unmarshal", args: args{msg: []byte("{\"DefaultOutputSource:\":0}")}, want: &Schedule{DefaultOutputSource: 0}, wantErr: false},
		{name: "Simple unmarshal", args: args{msg: []byte(testOne)}, want: testOneSch, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := umarshalSchedule(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("umarshalSchedule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("umarshalSchedule() got = %v, want %v", got, tt.want)
			}
		})
	}
}
