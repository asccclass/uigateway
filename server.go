package main

import (
   "os"
   "fmt"
   "strings"
   "github.com/joho/godotenv"
   "github.com/asccclass/sherryserver"
)

var chatService *InteractionService // 服務管理器，在 main() 中初始化
var McpHost *MCPHost           // MCPHost 用於處理 MCP Server 的能力

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

   server, err := SherryServer.NewServer(":" + port, documentRoot, templateRoot)
   if err != nil {
      panic(err)
   }
   router := NewRouter(server, documentRoot)
   if router == nil {
      fmt.Println("router return nil")
      return
   }
	sse := NewSSEService()
	sse.AddRouter(router)
   // MCP HOST 初始化
   if os.Getenv("MCPServiceName") != "" {
      McpHost = NewMCPHost()
      serviceNames := os.Getenv("MCPServiceName")
      parts := strings.Split(serviceNames, ",")
      for _, part := range parts {
         endpoint := "https://www.justdrink.com.tw/mcpsrv/capabilities/" + part
         if err := McpHost.AddCapabilities(part, endpoint); err != nil {
            fmt.Printf("獲取 MCP Server %s 服務失敗: %s\n", serviceNames,err.Error())        
         }
      }
   }
   
	// AI
	chatService = NewInteractionService()  // 服務初始化 (解決 nil pointer dereference)
    // 註冊工具
    tools := map[string]Tool{
        "get_current_weather": &WeatherTool{},
    }
	 prompt := "你是一個樂於助人的助手。如果你看到用戶問及天氣，請務必使用 get_current_weather 工具。"
	 // 註冊 Agent
	 agent, err := NewAgent("ollama", prompt, tools)
	 if err != nil {
	    fmt.Println("Failed to create Agent:", err)
	    return
	 }
	chatService.RegisterAgent("chat", agent)

   server.Server.Handler = router  // server.CheckCROS(router)  // 需要自行implement, overwrite 預設的
   server.Start()
}