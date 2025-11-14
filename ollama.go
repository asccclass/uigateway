package main

import (
	"fmt"
	"time"
	"strings"
)

type Ollama struct {
	Name string
	// 這裡可以放置 Ollama 相關的配置和狀態
}

func (app *Ollama) LLMName() string { 
	return app.Name 
}

func (m *Ollama) StreamGenerate(prompt string, tools []Tool) (<-chan *LLMChunk, error) {
    output := make(chan *LLMChunk)
    
    go func() {
        defer close(output)
        
        // 模擬文字輸出
        responseChunks := []string{"好的，", "我", "正在", "處理", "您的", "請求。"}
        for _, text := range responseChunks {
            output <- &LLMChunk{Text: text}
            time.Sleep(time.Millisecond * 30)
        }
        
        // 模擬工具呼叫：如果 Prompt 包含 "天氣" 且有工具可用，則模擬呼叫工具
        if strings.Contains(prompt, "天氣") && len(tools) > 0 {
            fmt.Println("--> Mock LLM Client: Simulating a tool call for weather.")
            output <- &LLMChunk{
                ToolCall: &ToolCall{
                    Name: "get_current_weather",
                    Args: map[string]interface{}{"location": "Taipei"},
                },
            }
        }
    }()
    return output, nil
}

func NewOllamaClient(name string) *Ollama {
	return &Ollama{Name: name}
}
