package takibi

import (
	"net/http"

	"github.com/poteto0/takibi/cloudflareenv"
	"github.com/poteto0/takibi/interfaces"
)

// defaultEnvResolver returns the struct-tag resolver for Bindings, or nil when
// Bindings carries no `env` / `cfbinding` tag. An invalid tag is a programming
// error, so it panics at app construction rather than on the first request.
func defaultEnvResolver[Bindings any](bindings *Bindings) interfaces.EnvResolverFunc[Bindings] {
	resolver, err := cloudflareenv.Resolver(bindings)
	if err != nil {
		panic(err)
	}
	return resolver
}

// OnEnv registers a resolver invoked once per request to build ctx.Env().
func (
	t *takibi[Bindings],
) OnEnv(
	resolver interfaces.EnvResolverFunc[Bindings],
) {
	t.envResolver = resolver
}

// resolveEnv builds the Bindings for one request or one Blow task. Without a
// resolver every caller shares the app-wide Bindings, as before.
func (
	t *takibi[Bindings],
) resolveEnv(
	r *http.Request,
) *Bindings {
	if t.envResolver == nil {
		return t.env
	}
	return t.envResolver(r)
}

// initializeContext takes a context out of the pool (or builds a new one) and
// gives it this request's env. The env is always re-assigned so a pooled
// context never carries the previous request's Bindings.
func (
	t *takibi[Bindings],
) initializeContext(
	w http.ResponseWriter,
	r *http.Request,
) interfaces.IContext[Bindings] {
	env := t.resolveEnv(r)

	ctx, ok := t.cache.Get().(interfaces.IContext[Bindings])
	if !ok {
		return NewContext(w, r, env, &t.option)
	}

	ctx.Reset(w, r)
	ctx.SetEnv(env)
	return ctx
}

// releaseContext returns ctx to the pool, dropping its env first. The env is a
// per-request value now, so an idle pooled context holding on to it would keep
// that request's Bindings — and on wasm the JS objects a `cfbinding` field
// wraps — alive until the context is picked up again.
func (
	t *takibi[Bindings],
) releaseContext(
	ctx interfaces.IContext[Bindings],
) {
	ctx.SetEnv(nil)
	t.cache.Put(ctx)
}

// newTaskContext builds the context handed to a Blow task. Like a request, a
// task gets its env from the resolver, so tagged Bindings are populated there
// too.
func (
	t *takibi[Bindings],
) newTaskContext(
	r *http.Request,
) interfaces.IContext[Bindings] {
	return NewContext(nil, r, t.resolveEnv(r), &t.option)
}
