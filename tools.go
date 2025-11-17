package main

import(
	"fmt"
    "regexp"
    "strings"
    "encoding/json"

)

// ToolResult 結構用於 Agent 執行完工具後，將結果傳回給 LLM
type ToolResult struct {
   Observation string // 工具執行後返回的結果（文字形式）
   Metadata map[string]interface{} // 可選：額外的元數據
}

// Tool 定義 Agent 可以使用的外部工具行為
type Tool interface {
    Name() string
    Description() string    
    FunctionSchema() string   // FunctionSchema() 返回 JSON 結構，用於 LLM 瞭解呼叫參數
    Execute(args map[string]interface{}) (*ToolResult, error)
}

// ToolCall 結構體：LLM 決定要呼叫的工具及其參數
type ToolCall struct {
	Name string // 工具名稱 (e.g., "search_web")
	Args map[string]interface{} // 呼叫工具所需的參數
}

// parseIntentWithOllama 使用 Ollama 解析使用者意圖
func parseIntent(llm OllamaGenerateRequest, srv *MCPServer) (map[string]interface{}, error) {
   prompt := fmt.Sprintf("%s\n\n使用者輸入：`%s`", srv.Name, llm.Prompt)
   fmt.Println(prompt)
   response := ""
   var err error

   res, err := Send2LLM(prompt, false)  // (string, error) 
   if err != nil {
      return nil, fmt.Errorf("query ollama for intent: %s", err.Error())
   }
   response = res
  
   // 清除 <think>...</think> 標籤   
   re := regexp.MustCompile(`(?s)^.*</think>`)
	response = re.ReplaceAllString(response, "")
   // 清除 ```
   response = strings.ReplaceAll(response, "```", "")
   // 清理回應，只保留 JSON 部分
   response = strings.TrimSpace(response)
   start := strings.Index(response, "{")
   end := strings.LastIndex(response, "}") + 1
   if start >= 0 && end > start {
      response = response[start:end]
   }
   var intent map[string]interface{}
   if err = json.Unmarshal([]byte(response), &intent); err != nil { // 如果 JSON 解析失敗，預設為一般對話
      return map[string]interface{}{
         "is_related": false,
         "action":          "general_chat",
         "parameters":      map[string]interface{}{},
      }, nil 

      return nil, fmt.Errorf("Parse intent failed（Mcpsrv 回傳格式錯誤）: %s", err.Error())  
   }   
   return intent, nil
}

/* Ollama 請求體
    reqBody := OllamaGenerateRequest{
		Model:    o.Model,
		Prompt:   prompt,
		Stream:   true, // 必須開啟串流
		System:   o.SystemPrompt,
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
*/    
func RunTools(reqBody OllamaGenerateRequest, o *Ollama)(string, error){
	if len(McpHost.ConnectedServers) == 0 {  // 檢查是否有連接的 MCP Server
		return "", fmt.Errorf("no connected MCP servers")	
	}	
	for _, srv:= range McpHost.ConnectedServers {  // 遍歷所有MCP Server
      if srv.IsRelatedPrompt == "" {
	     continue  // 如果沒有相關提示，則跳過
	   }
      s, err := parseIntent(reqBody, srv) // (map[string]interface{}, error)	  *********
      if err != nil {
         continue  // 如果解析不相關，則跳過  fmt.Println("解析意圖不相關:", err.Error())
      }
      action, ok := s["action"].(string)
      if !ok || action == "general_chat" {
	     continue  // 如果沒有動作，則跳過
	  }
	  tool, err := SearchTool(srv, action)  // (string, error)
	  if err != nil {
	     continue  // 如果查找工具失敗，則跳過
	  }
      parameters, ok := s["parameters"].(map[string]interface{})
	  if !ok {
	     parameters = make(map[string]interface{})
	  }
	  return callMCPTool(tool.Name, parameters)  // 調用 MCP 工具
	}
	return "", fmt.Errorf("未找到相關的 MCP Server 或工具")
}



// 模擬氣候工具： WeatherTool)模擬一個氣候工具
type WeatherTool struct{}
func (t *WeatherTool) Name() string { return "get_current_weather" }
func (t *WeatherTool) Description() string { return "查詢指定城市的天氣" }
func (t *WeatherTool) FunctionSchema() string {
    return `{"name": "get_current_weather", "parameters": {"type": "object", "properties": {"location": {"type": "string"}}}}`
}
func (t *WeatherTool) Execute(args map[string]interface{}) (*ToolResult, error) {
    location, ok := args["location"].(string)
    if !ok {
        return nil, fmt.Errorf("missing location argument")
    }
    // 模擬實際工具運作結果
    return &ToolResult{Observation: fmt.Sprintf("The weather in %s is currently 25°C and sunny.", location)}, nil
}