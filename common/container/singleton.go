package container

import "sync"

var (
	globalContainer IContainer
	once            sync.Once
)

func InitOnce() IContainer {
	once.Do(func() {
		globalContainer = newContainer()
	})
	return globalContainer
}

// MustFetch 从全局容器中获取某个类型的实例，例：db := MustFetch(new(infra.DB))，无法获取某个模块未导出的类型
func MustFetch[T any](expectType *T) T {
	return MustFetchWithScope(globalContainer, expectType)
}

// MustFetchWithScope 通过某个模块的 IScopeContainer 获取对应的实例，可以获取其他模块导出的类型和本模块注册的类型
func MustFetchWithScope[T any](c IScopeContainer, expectType *T) T {
	return c.MustFetch(expectType).(T)
}
