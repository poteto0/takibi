package takibi

type Option[Bindings any] func(*takibi[Bindings])

func WithMaxBodyBytes[Bindings any](n int64) Option[Bindings] {
	return func(t *takibi[Bindings]) { t.maxBodyBytes = n }
}
