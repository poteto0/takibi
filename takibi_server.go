//go:build !wasm

package takibi
import (
	"crypto/tls"
	"net"
	"strings"
	stdContext "context"

	"github.com/poteto0/takibi/constants"
)


func (t *takibi[Bindings]) Fire(addr string) error {
	t.fireMutex.Lock()

	if !strings.HasPrefix(addr, constants.PortPrefix) {
		addr = constants.PortPrefix + addr
	}

	t.Server.Addr = addr
	if err := t.setupServer(); err != nil {
		t.fireMutex.Unlock()
		return err
	}

	t.fireMutex.Unlock()

	t.startTasks()

	return t.Server.Serve(t.Listener)
}

func (t *takibi[Bindings]) setupServer() error {
	t.Server.Handler = t

	if t.Listener != nil {
		return nil
	}

	ln, err := net.Listen("tcp", t.Server.Addr)
	if err != nil {
		return err
	}

	if t.Server.TLSConfig == nil {
		t.Listener = ln
		return nil
	}

	t.Listener = tls.NewListener(ln, t.Server.TLSConfig)
	return nil
}

func (t *takibi[Bindings]) Finish(ctx stdContext.Context) error {
	t.fireMutex.Lock()
	if err := t.Server.Shutdown(ctx); err != nil {
		t.fireMutex.Unlock()
		return err
	}
	t.stopTasks(ctx)
	t.cancel()
	t.fireMutex.Unlock()
	return nil
}
