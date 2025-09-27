package container

import (
	"fmt"
	"go.uber.org/dig"
	"log"
	"reflect"
	"sync"
)

// IModule 由每个模块实现，用于声明该模块提供的类型的构造函数
type IModule interface {
	Name() string
	// ProvideConstructors 注册该模块的构造函数
	ProvideConstructors(r IModuleRegister)
	// AfterInstantiation 在所有注册的类型实例化之后的回调，可将 IScopeContainer 设置为该模块的全局变量，用户获取不同类型的实例
	AfterInstantiation(sc IScopeContainer)
}

// IScopeContainer 专属于一个module的container，可以获取到该模块所有Provide的类型，以及其他模块Provide为export的类型
type IScopeContainer interface {
	MustFetch(expect any) any
}

// IModuleRegister 用于向 container 注册该模块提供的类型的构造方法，该接口的方法都为声明类型，不执行实际的构造操作
type IModuleRegister interface {
	// Decorate 通过传入的装饰器方法对已有的类型进行装饰或直接替换为新的实现，通常用于测试时对依赖的类型进行fake，只对当前模块生效
	// example：用一个新的 constructor 替换 IContainer：r.Decorate(func(someDependencies) exporter.IContainer{...})
	Decorate(newContainer ConstructorFunc)
	// Provide 注册/声明某个/某些类型的构造方法，默认只对本模块开放，若要暴露给其他模块获取，使用 Exporter Option
	Provide(constructor ConstructorFunc, options ...ProvideOption)
}

// ConstructorFunc 类型构造方法，必须为 func 类型，入参表示依赖，出参是构造结果
type ConstructorFunc any

type moduleContainer struct {
	name          string
	scope         *dig.Scope
	instanceStore sync.Map
	providers     []*provider
	decorators    []ConstructorFunc
}

func newModuleScope(name string, container *dig.Container) *moduleContainer {
	return &moduleContainer{
		name:          name,
		scope:         container.Scope(name),
		instanceStore: sync.Map{},
	}
}

func (m *moduleContainer) Decorate(constructor ConstructorFunc) {
	m.decorators = append(m.decorators, constructor)
}

func (m *moduleContainer) Provide(constructor ConstructorFunc, opts ...ProvideOption) {
	options := &provideOption{}
	for _, opt := range opts {
		options = opt(options)
	}
	p := &provider{
		constructor: constructor,
		options:     options,
	}

	// 记录 constructors 的出参类型：如果原来 constructor 有as参数，忽略原来的类型，as参数一定是接口类型的指针
	if len(options.as) > 0 {
		for _, as := range options.as {
			_type := reflect.TypeOf(as).Elem()
			p.outputTypes = append(p.outputTypes, _type)
		}
	} else {
		cType := reflect.TypeOf(constructor)
		for i := 0; i < cType.NumOut(); i++ {
			_type := cType.Out(i)
			if isErrType(_type) {
				continue
			}
			p.outputTypes = append(p.outputTypes, _type)
		}
	}
	// 这里不实际的向container注册，先存起来
	m.providers = append(m.providers, p)
}

func (m *moduleContainer) doRegister() {
	for _, p := range m.providers {
		// 向当前scope注册，根据export字段决定是否向外部暴露
		err := m.scope.Provide(p.constructor, dig.As(p.options.as...), dig.Export(p.options.export))
		if err != nil {
			panic(fmt.Sprintf("[module: %s] provide %T error: %s", m.name, p.constructor, err.Error()))
		}
	}
	for _, decorator := range m.decorators {
		// 如果 decorate的类型之前没有provide先provide一个，忽略重复provide报错
		_ = m.scope.Provide(decorator)
		err := m.scope.Decorate(decorator)
		if err != nil {
			panic(fmt.Sprintf("[module: %s] decorate %T error: %s", m.name, decorator, err.Error()))
		}
	}
}

func (m *moduleContainer) instantiate() {
	for _, p := range m.providers {
		funcType := reflect.FuncOf(
			p.outputTypes,
			[]reflect.Type{},
			false,
		)
		f := reflect.MakeFunc(funcType, func(args []reflect.Value) (results []reflect.Value) {
			for _, arg := range args {
				log.Printf("[module: %s] instantiate type %s, exported: %v", m.name, arg.Type().String(), p.options.export)
			}
			return
		}).Interface()
		err := m.scope.Invoke(f)
		if err != nil {
			panic(fmt.Sprintf("[module: %s] instantiate error: %+v", m.name, err))
		}
	}
}

func (m *moduleContainer) MustFetch(expect any) any {
	expectType := reflect.TypeOf(expect)
	if v, ok := m.instanceStore.Load(expectType); ok {
		return v
	}
	value, err := fetchInScope(m.scope, expectType)
	if err != nil {
		panic(fmt.Sprintf("[module: %s] fetch expect type: %s error: %+v", m.name, expectType, err))
	}
	return value
}

type scopeInvoker interface {
	Invoke(function interface{}, opts ...dig.InvokeOption) error
}

func fetchInScope(scope scopeInvoker, expect any) (any, error) {
	var value any
	funcType := reflect.FuncOf(
		[]reflect.Type{reflect.TypeOf(expect).Elem()},
		[]reflect.Type{},
		false,
	)
	f := reflect.MakeFunc(funcType, func(args []reflect.Value) (results []reflect.Value) {
		value = args[0].Interface()
		return
	}).Interface()
	err := scope.Invoke(f)
	if err != nil {
		return nil, err
	}
	return value, nil
}

var _errType = reflect.TypeOf((*error)(nil)).Elem()

func isErrType(t reflect.Type) bool {
	return t.Implements(_errType)
}

type provider struct {
	constructor ConstructorFunc
	options     *provideOption
	outputTypes []reflect.Type
}
