package main

import (
	"github.com/cm06137/cm_project/common/container"
	"github.com/cm06137/cm_project/mcp_server/starter"
)

func main() {
	ContainerInit()
}

func ContainerInit() {
	c := container.InitOnce()
	// 注册各业务模块
	c.Register(&starter.Constructor{})
	// 实例化各容器内类型
	c.MustInitialize()
}
