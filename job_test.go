package schedula

import (
	"bytes"
	"encoding/json"
	"strconv"
	"testing"
)

func TestJSONMarshal(t *testing.T) {
	var expected = []byte(`{"id":"","businessKey":"","callbackURL":"http://example.com/","data":null,"schedule":{"format":"timestamp","value":"1438948984"},"status":""}`)
	job := aJob(1438948984)

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
	job := aJob(expected)
	actual, err := job.Schedule.NextTimestamp()
	if actual != expected || err != nil {
		t.Fatalf("expected %d but got %d (error: %v)", expected, actual, err)
	}
}

func aJob(timestamp int64) Job {
	return Job{
		CallbackURL: "http://example.com/",
		Schedule: JobSchedule{
			Format: ScheduleFormatTimestamp,
			Value:  strconv.FormatInt(timestamp, 10),
		},
	}
}
