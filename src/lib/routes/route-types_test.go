package routetypes

import (
	"fmt"
	"net/http"
	"syscall"
	"testing"
	"time"
)

// After SIGTERM, Listen must keep serving for drainDelay (while the load
// balancer stops routing to the pod), then return.
func TestListenDrainsOnSIGTERM(t *testing.T) {
	drainDelay = 500 * time.Millisecond

	ok := func(w http.ResponseWriter, r *http.Request, route Route) { fmt.Fprint(w, "ok") }
	r := &Routes{Port: "18765", Routes: []Route{{Route: "/ok", Handler: ok}}}
	url := "http://127.0.0.1:18765/ok"

	done := make(chan struct{})
	go func() { r.Listen(); close(done) }()

	get := func() int {
		resp, err := http.Get(url)
		if err != nil {
			return 0
		}
		resp.Body.Close()
		return resp.StatusCode
	}

	// Wait for the server to come up.
	for i := 0; i < 50 && get() != 200; i++ {
		time.Sleep(20 * time.Millisecond)
	}

	syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	time.Sleep(100 * time.Millisecond)

	if code := get(); code != 200 {
		t.Fatalf("during drain: got %d, want 200", code)
	}

	select {
	case <-done:
		t.Fatal("Listen returned before drainDelay elapsed")
	case <-time.After(200 * time.Millisecond):
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Listen didn't return after drainDelay")
	}
}
