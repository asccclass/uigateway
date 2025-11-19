package main

import(
	"fmt"
    "regexp"
    "strings"
    "encoding/json"

)

func Prompt2String(reqBody GenerateRequest, role, prompt string)(string, error) {  //req GenerateRequest, role, prompt, base64Image string)(string, error) {
/*   
   if base64Image != "" {
      req.Images = []string{base64Image}
      req.Messages = append(req.Messages, Message{Role: role, Content: prompt, Images: []string{base64Image}})  // 如果沒有工具套用，則使用原始提示
   } else {
      req.Messages = append(req.Messages, Message{Role: role, Content: prompt})
   }
*/      
   reqBody.Messages = append(reqBody.Messages, Message{Role: role, Content: prompt})
   jData, err := json.Marshal(reqBody)  // 將請求轉為 JSON
   if err != nil {
      return "", fmt.Errorf("prompt to string's json marshal failed, 序列化請求失敗: %s", err.Error())
   }
   return string(jData), nil
}

// parseIntentWithOllama 使用 Ollama 解析使用者意圖
func parseIntent(reqBody GenerateRequest,pt string, srv *MCPServer, llmClient LLMProviderClient) (map[string]interface{}, error) {
   // reqBody.System = srv.IsRelatedPrompt
   prompt := fmt.Sprintf("%s\n\n使用者輸入：`%s`", srv.IsRelatedPrompt, pt)   // 組合判斷是否需要使用工具的prompt

   response := ""
   var err error
   
   jData, err := Prompt2String(reqBody, "user", prompt)  // (string, error)
   if err != nil {
      return nil, fmt.Errorf("prepare prompt for ollama: %s", err.Error())
   }
   res, err := llmClient.Send2LLM(jData, false)  // (string, error) 回傳：{"is_related": false, "action": "general_chat", "parameters": {}}
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
// 判斷是否需要使用 MCP 工具，並執行相應工具
func CheckTools(prompt string, llmClient LLMProviderClient)(string, error){
	if len(McpHost.ConnectedServers) == 0 {  // 檢查是否有連接的 MCP Server
		return "", fmt.Errorf("no connected MCP servers")	
	}

   reqBody := GenerateRequest {  // 初始化
      Model:  llmClient.GetModel(),
      Messages: []Message{
         {
            Role:    "system",
            Content: "請務必使用繁體中文進行所有回覆，不可使用簡體中文。", // 設定系統指令
        },
      },   // role, content
      Stream: false,          // 因為要一次取得json，故不使用串流模式
   }
	for _, srv:= range McpHost.ConnectedServers {  // 遍歷所有MCP Server
      if srv.IsRelatedPrompt == "" {
	     continue  // 如果沒有相關提示，則跳過判斷
	   }
      s, err := parseIntent(reqBody, prompt, srv, llmClient) // (map[string]interface{}, error)	
      if err != nil {
         continue  // 如果解析不相關，則跳過  fmt.Println("解析意圖不相關:", err.Error())
      }
      isRelate, ok := s["is_related"].(bool)
      if !ok || !isRelate {
	     continue  // 如果解析不相關，則跳過
	  }
      action, ok := s["action"].(string)
      if !ok || action == "" {
	     continue  // 如果沒有動作，則跳過
	  }
	  tool, err := srv.SearchTool(action)  // (string, error)
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