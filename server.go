package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	SherryServer "github.com/asccclass/sherryserver"
	"github.com/joho/godotenv"
)

var chatService *InteractionService // 服務管理器，在 main() 中初始化
var McpHost *MCPHost                // MCPHost 用於處理 MCP Server 的能力

func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if err := godotenv.Load(currentDir + "/envfile"); err != nil {
		fmt.Println(err.Error())
		return
	}

	// Update WebSocket URL in index.html if defined in env
	if wsUrl := os.Getenv("WebSocketUrl"); wsUrl != "" {
		fmt.Printf("🚀 偵測到 WebSocketUrl 設定: %s，正在更新 Frontend 配置...\n", wsUrl)
		indexPath := currentDir + "/www/html/index.html"
		content, err := os.ReadFile(indexPath)
		if err == nil {
			newContent := strings.Replace(string(content), "const wsUrl = 'ws://localhost:9090/ws';", fmt.Sprintf("const wsUrl = '%s';", wsUrl), 1)
			if err := os.WriteFile(indexPath, []byte(newContent), 0644); err != nil {
				fmt.Printf("❌ 更新 index.html 失敗: %v\n", err)
			} else {
				fmt.Println("✅ index.html WebSocket URL 已更新")
			}
		} else {
			fmt.Printf("❌ 讀取 index.html 失敗: %v\n", err)
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	documentRoot := os.Getenv("DocumentRoot")
	if documentRoot == "" {
		documentRoot = "www/html"
	}
	templateRoot := os.Getenv("TemplateRoot")
	if templateRoot == "" {
		templateRoot = "www/template"
	}

	server, err := SherryServer.NewServer(":"+port, documentRoot, templateRoot)
	if err != nil {
		panic(err)
	}
	router := NewRouter(server, documentRoot)
	if router == nil {
		fmt.Println("router return nil")
		return
	}
	// MCP HOST 初始化
	if os.Getenv("MCPServiceName") != "" {
		var wg sync.WaitGroup // 使用 WaitGroup
		McpHost = NewMCPHost()
		serviceNames := os.Getenv("MCPServiceName")
		parts := strings.Split(serviceNames, ",")
		fmt.Printf("🚀 開始非同步處理 %d 個服務...\n", len(parts))
		for _, part := range parts {
			wg.Add(1) // 增加計數器

			go func(part string) {
				defer wg.Done()
				endpoint := "https://www.justdrink.com.tw/mcpsrv/capabilities/" + part
				if err := McpHost.AddCapabilities(part, endpoint); err != nil {
					fmt.Printf("獲取 MCP Server: %s 服務失敗: %s\n", part, err.Error())
				}
			}(part)
		}
		wg.Wait() // 等待所有 goroutine 完成
	}

	// AI
	chatService = NewInteractionService() // 服務初始化 (解決 nil pointer dereference)
	prompt := "你是一個樂於助人的助手"
	// 註冊 Agent
	agent, err := NewAgent("ollama", prompt)
	if err != nil {
		fmt.Println("Failed to create Agent:", err)
		return
	}
	chatService.RegisterAgent("chat", agent)

	// SSE 服務註冊
	sse := NewSSEService()
	sse.AddRouter(router)

	server.Server.Handler = router // server.CheckCROS(router)  // 需要自行implement, overwrite 預設的
	server.Server.WriteTimeout = 0
	server.Server.ReadTimeout = 0
	server.Start()
}
