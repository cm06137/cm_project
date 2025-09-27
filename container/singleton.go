package container

// IContainer: 主容器，会被初始化为全局单例，用于管理所有模块的注册和各类型的实例化，同时可以从主容器中获取到各模块对外暴露的类型
//
// 注意：一般实例都是通过各自模块专属的 IScopeContainer 获取，主 container 主要用于流程管理，仅可以从主 container 中获取各模块Export的类型实例
//
//使用流程：
//	1. 初始化煮容器
//	c := InitOnce()
//	2. 注册所有模块
//	c.Register(module1)
//	c.Register(module2)
//	...
//	3. 初始化所有模块，将各模块 Provide 的类型实例化
//	c.MustInitialize()
//	该方法调用后，会回调各module的AfterInstantiation，并传入模块专属的IScopedContainer，在回调方法中，各模块可以将专属container设置为全局变量以动态获取实例
//	4. 业务模块获取实例
//	moduleContainer.Fetch(new(xxxType).(xxxType))
//	也可以用全局的泛型函数， MustFetchWithScope(moduleContainer, new(xxxType)) 省去类型断言的操作

type IContainer interface {
	MustInitialize()
	Register(module IModule)
	MustFetch()
}

type IModule interface{}
