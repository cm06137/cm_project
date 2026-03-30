package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

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

	server := server2.NewMCPServer("mcp_server_demo", "0.0.1", server2.WithToolCapabilities(false))
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

	// 4. 创建一个可流式的 HTTP 服务器
	// 这个 HTTP 服务器封装了原始的 MCP 服务器
	// 默认的端点路径是 "/:serverKey/stream"，其中 :serverKey 会被替换为 mcpServer 的名称
	streamingHTTPServer := server2.NewStreamableHTTPServer(
		server,
		// 可以通过 WithEndpointPath 来自定义流的 URL 路径
		// server2.WithEndpointPath("/:serverKey/stream"), // 这是默认值
	)

	// 5. 将 HTTP 处理器注册到 Go 的标准 http 包上
	http.Handle("/", streamingHTTPServer) // 根路径将处理所有 MCP 相关的请求

	// 6. 启动 HTTP 服务器
	addr := ":8080"
	log.Printf("HTTP MCP Server starting on %s", addr)
	log.Printf("Stream endpoint will be available at: http://localhost:%s/%s/stream", addr, "mcp_server_demo")
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("HTTP server failed to start: %v", err)
	}
}

func calculateHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	log.Printf("Calculate calculator called with req: %v", req)

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
