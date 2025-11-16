package main

import (
	"fmt"
	"strings"
)

// LLMProviderClient 介面：一個虛擬的 LLM 客戶端介面，用於抽象化不同的 LLM SDK
type LLMProviderClient interface {
	LLMName() string
   // StreamGenerate 接受 Prompt 和 Tools，返回一個輸出串流 Channel
    StreamGenerate(prompt string, tools []Tool) (<-chan *LLMChunk, error)
}

// LLM 服務的客戶端 (例如：*openai.Client, *genai.Client, 或統一的 SDK)
type CustomChatAgent struct {
	LLMClient   LLMProviderClient 
	SystemPrompt string // 定義 Agent 個性和規則的 System Prompt
	Tools       map[string]Tool // 儲存所有可用的工具，以 Name 為 Key
}

// executeTool 根據 ToolCall 執行對應的工具，並返回 ToolResult
func(app *CustomChatAgent) executeTool(toolCall *ToolCall) (*ToolResult, error) {
    tool, ok := app.Tools[toolCall.Name]   // 1. 查找工具：從 Agent 的 Tools 映射中找到具體的 Tool 實例
    if !ok {        
       return nil, fmt.Errorf("tool %s not found", toolCall.Name)  // 如果找不到工具，返回錯誤
    }    
    return tool.Execute(toolCall.Args)   // 2. 執行工具：呼叫 Tool 介面的 Execute 方法
} 

func (app *CustomChatAgent) Name() string { return app.LLMClient.LLMName() }

// getAvailableTools 解決 undefined: a.getAvailableTools
func (app *CustomChatAgent) getAvailableTools() []Tool {
    tools := make([]Tool, 0, len(app.Tools))
    for _, tool := range app.Tools {
        tools = append(tools, tool)
    }
    return tools
}

// ProcessStream Agent 核心邏輯
func (a *CustomChatAgent) ProcessStream(input *InteractionInput, outputChannel chan *StreamChunk) error {
   defer close(outputChannel)
	
   // 解決 undefined: a.buildInitialPrompt 的方法
   fullPrompt := fmt.Sprintf("%s\nUser Query: %s", a.SystemPrompt, input.Query)
   maxIterations := 5
    
   // 1. 傳送開始事件
   outputChannel <- &StreamChunk{
      Type: "start",
      Data: map[string]string{"status": "thinking"},
   }

   for currentIteration := 0; currentIteration < maxIterations; currentIteration++ {
        llmStream, err := a.LLMClient.StreamGenerate(fullPrompt, a.getAvailableTools())  // 呼叫 LLM 串流
        if err != nil {
            outputChannel <- &StreamChunk{Type: "error", Data: err.Error()}
            return fmt.Errorf("LLM stream error: %w", err)
        }
        var toolCall *ToolCall
        var textOutput strings.Builder // 僅用於單輪推理的文字緩衝

        for chunk := range llmStream {
            if chunk.Text != "" {
                textOutput.WriteString(chunk.Text)
                outputChannel <- &StreamChunk{Type: "message", Data: chunk.Text}
            }            
            if chunk.ToolCall != nil {
                toolCall = chunk.ToolCall
                break // LLM 決定呼叫工具，中斷文字輸出
            }
        }
        // 處理工具呼叫
        if toolCall != nil {
            currentIteration++            
            outputChannel <- &StreamChunk{   // 通知前端 Agent 正在呼叫工具
                Type: "tool_call",
                Data: map[string]string{"tool_name": toolCall.Name, "status": "executing"},
            }
            // 執行工具
            toolResult, err := a.executeTool(toolCall)
            if err != nil {  // 將工具執行錯誤回報給 LLM                
               toolResult = &ToolResult{Observation: fmt.Sprintf("Tool execution failed: %s", err.Error())}
            }            
            // 將工具執行結果追加到 Prompt 中，進行下一輪推理 (Re-prompting)
            toolObservation := fmt.Sprintf("\n\nObservation for tool %s: %s\n\n", toolCall.Name, toolResult.Observation)
            fullPrompt += toolObservation
            continue   // 繼續迴圈 (進行下一輪 LLM 呼叫)
        }
        // 如果沒有工具呼叫，且有文字輸出，則認為回應完成 || 達到最大輪次，強制結束
        if textOutput.Len() > 0  || currentIteration == maxIterations-1 {
            break 
        }
    }
    // 串流結束
	outputChannel <- &StreamChunk{
        Type: "end",
        Data: map[string]string{"status": "complete"},
    }
	return nil
}

func NewLLMClient(llmName string) (LLMProviderClient) {
	switch strings.ToLower(llmName) {
	case "ollama":	
        olm := &Ollama{Name: llmName}
		return olm
	}
	return nil
}

// 建立並返回一個新的 Agent 實例
func NewAgent(llmName string, systemPrompt string, tools map[string]Tool) (*CustomChatAgent, error) {   
	 llmClient := NewLLMClient(llmName)   // 設定使用的LLM服務
	 if llmClient == nil {
	    fmt.Println("Failed to create LLM client")
	    return nil, fmt.Errorf("failed to create LLM client")
	 }
    // 註冊 Agent
   return &CustomChatAgent{
        LLMClient: llmClient, // 使用指定的 LLM 客戶端
        SystemPrompt: systemPrompt,
        Tools: tools,
    }, nil
}