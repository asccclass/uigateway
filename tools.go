package main

import(
	"fmt"
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