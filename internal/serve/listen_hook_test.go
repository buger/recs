package serve

import (
	"net"
	"net/http"
	"testing"

	"github.com/buger/recs/internal/app"
)

// Verifies: SYS-REQ-260820-9W1S
// SYS-REQ-260820-9W1S
func TestListenSuccessfulBindReturns(t *testing.T) {
	old := serveHTTP
	serveHTTP = func(ln net.Listener, h http.Handler) error {
		_ = ln.Close()
		return http.ErrServerClosed
	}
	t.Cleanup(func() { serveHTTP = old })
	a := app.OpenOrCWD(t.TempDir())
	if err := a.Init(); err != nil {
		t.Fatal(err)
	}
	if err := Listen(a, 0); err != http.ErrServerClosed {
		t.Fatalf("got %v", err)
	}
}
