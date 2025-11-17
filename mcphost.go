package main
/*
## 程式說明
這個Go程式模擬了MCP Host與多個MCP Server進行能力協商的過程，主要功能包括：

1. MCP Host初始化 ：創建一個MCP Host實例，準備連接到多個MCP Server。
2. 與MCP Server連接並進行能力協商 ：
   
   - 連接到多個MCP Server（示例中是file_server和database_server）
   - 從每個Server獲取其提供的工具列表和描述
   - 在實際應用中，這會通過HTTP請求實現
3. 處理用戶查詢 ：
   
   - 收集所有連接的Server提供的工具
   - 將用戶查詢和工具描述一起發送給LLM
   - LLM分析用戶需求，決定使用哪些工具
4. 執行LLM選擇的工具 ：
   
   - 解析LLM選擇的工具名稱，找到對應的Server
   - 向相應的Server發送請求，執行選定的工具
這個程式展示了MCP Host如何管理多個MCP工具的完整流程，包括初始化連接、能力協商、工具選擇和執行。
在實際應用中，HTTP請求和響應處理會更加複雜，但基本流程是相同的。
*/

import (
	"io"
   "fmt"
   "time"
   "net/http"
   "encoding/json"
)

// MCP Server提供的工具定義
type ToolDesc struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Parameters  map[string]string `json:"parameters,omitempty"`
}

// MCP Server的能力描述
type ServerCapabilities struct {
	Version  string `json:"version"`
	ServerID string `json:"server_id"`
	Tools    []ToolDesc `json:"tools"`
}

// MCP Server結構
type MCPServer struct {
	ID           string					`json:"id"` // Server ID
	Name         string					`json:"name"` // Server名稱			
	Capabilities ServerCapabilities	`json:"capabilities"` // Server能力描述
	Endpoint     string					`json:"endpoint,omitempty"` // Server的API端點
	IsRelatedPrompt string    			`json:"isRelatedPrompt,omitempty"` // 是否與ID服務事項相關
	ProcessPrompt string					`json:"processPrompt,omitempty"` // 處理ID服務事項的提示，若是則需要做何處理
}

// MCP Host結構
type MCPHost struct {
	ConnectedServers map[string]*MCPServer
}

// 連接到MCP Server並進行能力協商，只執行一次
func(h *MCPHost) AddCapabilities(serviceName, endpoint string) (error) {
	// 建立HTTP客戶端，設定逾時時間
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	// 發送GET請求
	resp, err := client.Get(endpoint)
	if err != nil {
		return fmt.Errorf("HTTP請求失敗: %v", err)
	}
	defer resp.Body.Close()
	// 檢查HTTP狀態碼
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP狀態碼錯誤: %d", resp.StatusCode)
	}
	// 讀取回應內容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("讀取回應內容失敗: %v", err)
	}
	// 解析JSON
	var server MCPServer
	if err := json.Unmarshal(body, &server); err != nil {
		return fmt.Errorf("JSON解析失敗: %v", err)
	}
	h.ConnectedServers[serviceName] = &server
	fmt.Printf("成功連接到MCP Server: %s，獲取到%d個工具\n", serviceName, len(server.Capabilities.Tools))
	return nil
}

// 初始化MCP Host
func NewMCPHost()(*MCPHost) {
	return &MCPHost{
		ConnectedServers: make(map[string]*MCPServer),
	}
}