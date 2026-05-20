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
	t.engine.fireMutex.Lock()

	if !strings.HasPrefix(addr, constants.PortPrefix) {
		addr = constants.PortPrefix + addr
	}

	t.engine.Server.Addr = addr
	if err := t.setupServer(); err != nil {
		t.engine.fireMutex.Unlock()
		return err
	}

	t.engine.fireMutex.Unlock()

	t.startTasks()

	return t.engine.Server.Serve(t.engine.Listener)
}

func (t *takibi[Bindings]) setupServer() error {
	t.engine.Server.Handler = t

	if t.engine.Listener != nil {
		return nil
	}

	ln, err := net.Listen("tcp", t.engine.Server.Addr)
	if err != nil {
		return err
	}

	if t.engine.Server.TLSConfig == nil {
		t.engine.Listener = ln
		return nil
	}

	t.engine.Listener = tls.NewListener(ln, t.engine.Server.TLSConfig)
	return nil
}

func (t *takibi[Bindings]) Finish(ctx stdContext.Context) error {
	t.engine.fireMutex.Lock()
	if err := t.engine.Server.Shutdown(ctx); err != nil {
		t.engine.fireMutex.Unlock()
		return err
	}
	t.stopTasks(ctx)
	t.cancel()
	t.engine.fireMutex.Unlock()
	return nil
}
