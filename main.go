package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/cm06137/cm_project/common/container"
	"github.com/cm06137/cm_project/mcp_server/starter"
	"github.com/mark3labs/mcp-go/mcp"

	server2 "github.com/mark3labs/mcp-go/server"
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

	server := server2.NewMCPServer("mcp server demo", "0.0.1", server2.WithToolCapabilities(false))
	//srv := server2.NewStreamableHTTPServer(server,
	//	server2.WithEndpointPath("/:serverKey/stream"))

	// 注册计算器工具
	calculatorTool := mcp.NewTool("calculate",
		mcp.WithDescription("基础算术运算工具"),
		mcp.WithString("operation"),
		mcp.WithNumber("x",
			mcp.Required(),
			mcp.Description("第一个运算数"),
		),
		mcp.WithNumber("y",
			mcp.Required(),
			mcp.Description("第二个运算数"),
		),
	)

	// 添加请求处理逻辑
	server.AddTool(calculatorTool, calculateHandler)

	if err := server2.ServeStdio(server); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}

func calculateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	op := req.Params.Arguments.(map[string]any)["operation"].(string)
	x := req.Params.Arguments.(map[string]any)["x"].(float64)
	y := req.Params.Arguments.(map[string]any)["y"].(float64)

	var result float64
	switch op {
	case "add":
		result = x + y
	case "subtract":
		result = x - y
	case "multiply":
		result = x * y
	case "divide":
		if y == 0 {
			return nil, errors.New("除数不能为零")
		}
		result = x / y
	default:
		return nil, fmt.Errorf("不支持的操作: %s", op)
	}
	return mcp.NewToolResultText(fmt.Sprintf("%.2f", result)), nil
}
