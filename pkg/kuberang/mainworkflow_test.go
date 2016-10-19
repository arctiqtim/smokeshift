package kuberang

import (
	"net/http"
	"testing"
)

func TestTimeout(t *testing.T) {
	client := http.Client{
		Timeout: HTTP_Timeout,
	}
	// Simulate a timeout using httpbin.org
	if _, err := client.Get("http://httpbin.org/delay/2"); err == nil {
		t.Log(err)
		t.Errorf("Expected to timeout")
	}
	if _, err := client.Get("http://httpbin.org/delay/.1"); err != nil {
		t.Log(err)
		t.Errorf("Expected not to timeout and recieve a response")
	}

}
