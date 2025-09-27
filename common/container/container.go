package container

import (
	"fmt"
	"go.uber.org/dig"
	"log"
	"reflect"
	"sync"
)

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
	MustFetch(expect any) any
}

type container struct {
	c            *dig.Container
	modules      map[string]IModule
	moduleScopes map[string]*moduleContainer
	store        sync.Map
}

func newContainer() *container {
	return &container{
		c:            dig.New(),
		modules:      make(map[string]IModule),
		moduleScopes: make(map[string]*moduleContainer),
		store:        sync.Map{},
	}
}

func (c *container) MustInitialize() {
	if len(c.modules) == 0 {
		panic("no module registered, call Register first")
	}
	// 调用每个模块的构造函数，实例化对象
	for _, scope := range c.moduleScopes {
		scope.instantiate()
	}
	// 初始化后每个模块的回调
	for _, m := range c.modules {
		m.AfterInstantiation(c.moduleScopes[m.Name()])
	}
}

func (c *container) Register(module IModule) {
	if _, ok := c.modules[module.Name()]; ok {
		panic(fmt.Sprintf("module already registered:%s", module.Name()))
	}
	ms := newModuleScope(module.Name(), c.c)
	module.ProvideConstructors(ms)
	ms.doRegister()

	c.modules[module.Name()] = module
	c.moduleScopes[module.Name()] = ms

	log.Printf("[container] Registere module :%s", module.Name())
}

func (c *container) MustFetch(expect any) any {
	expectType := reflect.TypeOf(expect)
	if v, ok := c.store.Load(expectType); ok {
		return v
	}
	value, err := fetchInScope(c.c, expect)
	if err != nil {
		panic(fmt.Sprintf("[container] fetch expect type: %s, error: %+v", expectType, err))
	}
	value, _ = c.store.LoadOrStore(expectType, value)
	return value
}
