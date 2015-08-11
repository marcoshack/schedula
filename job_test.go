package schedula

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"
)

func TestJSONMarshal(t *testing.T) {
	var expected = []byte(`{"id":"","businessKey":"","callbackURL":"http://example.com/","data":null,"timeout":{"format":"timestamp","value":"1438948984"}}`)
	j := &Job{CallbackURL: "http://example.com/", Timeout: JobTimeout{Format: "timestamp", Value: "1438948984"}}

	b, err := json.Marshal(j)
	if err != nil {
		log.Printf("TestJSONMarshal: %s", err)
		t.FailNow()
	}

	if !bytes.Equal(b, expected) {
		log.Printf("TestJSONMarshal: expected '%s' but got '%s'", expected, b)
		t.FailNow()
	}
}
