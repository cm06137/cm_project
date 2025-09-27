package container

type ProvideOption func(option *provideOption) *provideOption

type provideOption struct {
	as     []any
	export bool
}
