package httpclient

import (
	"fmt"
	"testing"
)

func TestPostWithJson(t *testing.T) {

}

func TestPostFile(t *testing.T) {
	b, err := PostFormMultipart(
		"http://0.0.0.0:8880/demo/upload",
		"file",
		"/Users/switch/Downloads/LoadTest.exe",
		map[string]string{
			"aa": "hello",
		},
	)
	if err != nil {
		t.Fail()
	}

	fmt.Println(string(b))
}
