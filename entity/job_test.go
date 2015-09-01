package entity

import (
	"bytes"
	"encoding/json"
	"strconv"
	"testing"
)

func TestJSONMarshal(t *testing.T) {
	var expected = []byte(`{"id":"","businessKey":"","callbackURL":"http://example.com/","data":null,"schedule":{"format":"timestamp","value":"1438948984"},"status":""}`)

	job := &Job{
		CallbackURL: "http://example.com/",
		Schedule: JobSchedule{
			Format: ScheduleFormatTimestamp,
			Value:  strconv.FormatInt(1438948984, 10),
		},
	}

	b, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("failed marshaling job: %s", err)
	}

	if !bytes.Equal(b, expected) {
		t.Fatalf("expected '%s' but got '%s'", expected, b)
	}
}

func TestNextTimestamp(t *testing.T) {
	var expected int64 = 1438948984
	job := &Job{
		CallbackURL: "http://example.com",
		Schedule: JobSchedule{
			Format: ScheduleFormatTimestamp,
			Value:  strconv.FormatInt(expected, 10),
		},
	}

	actual, err := job.Schedule.NextTimestamp()
	if actual != expected || err != nil {
		t.Fatalf("expected %d but got %d (error: %v)", expected, actual, err)
	}
}
