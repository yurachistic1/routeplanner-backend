package overpass

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func init() {
	initialDelay = time.Millisecond
}

func TestQueryRetriesWhenBusy(t *testing.T) {

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++

		if got := r.Header.Get("User-Agent"); got != UserAgent {
			t.Errorf("User-Agent == %q, want %q", got, UserAgent)
		}
		if got := r.FormValue("data"); got != "[out:json];" {
			t.Errorf("data == %q, want %q", got, "[out:json];")
		}

		if calls < 3 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		fmt.Fprint(w, `{"elements":[{"type":"node","id":1,"lat":51.5,"lon":-0.1}]}`)
	}))
	defer server.Close()

	res, err := Query(context.Background(), server.URL, "[out:json];")

	if err != nil {
		t.Fatalf("Query returned error: %v", err)
	}
	if calls != 3 {
		t.Errorf("server called %d times, want 3", calls)
	}
	if len(res.Elements) != 1 || res.Elements[0].Id != 1 {
		t.Errorf("Query returned elements %+v, want one node with id 1", res.Elements)
	}
}

func TestQueryErrors(t *testing.T) {

	cases := []struct {
		name      string
		status    int
		body      string
		wantCalls int
	}{
		{"gives up after max attempts", http.StatusGatewayTimeout, "", maxAttempts},
		{"does not retry bad request", http.StatusBadRequest, "", 1},
		{"invalid json", http.StatusOK, "# not json", 1},
	}

	for _, c := range cases {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.WriteHeader(c.status)
			fmt.Fprint(w, c.body)
		}))

		_, err := Query(context.Background(), server.URL, "[out:json];")
		server.Close()

		if err == nil {
			t.Errorf("%s: Query returned no error", c.name)
		}
		if calls != c.wantCalls {
			t.Errorf("%s: server called %d times, want %d", c.name, calls, c.wantCalls)
		}
	}
}
