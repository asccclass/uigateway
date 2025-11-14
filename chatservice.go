package main

import(
)

// 定義串流片段的結構，用於 Agent 將數據串流傳給 HTTP Handler (最終給前端)
type StreamChunk struct {
	Type  string      // message, tool_call, end, ui_patch
	Data  interface{} // 實際傳輸的數據 (例如：字串或結構化的 JSON 物件)
}

// LLMChunk 結構體：代表 LLM 串流回傳的一個片段
type LLMChunk struct {
	Text     string     // 文字片段 (例如：一個單字或標點符號)
	ToolCall *ToolCall  // 如果 LLM 決定呼叫工具，這裡會有資料
	// Error    error    // 可選：串流中的錯誤訊息
}

// LLMProviderClient 介面：一個虛擬的 LLM 客戶端介面，用於抽象化不同的 LLM SDK
type LLMProviderClient interface {
	LLMName() string
   // StreamGenerate 接受 Prompt 和 Tools，返回一個輸出串流 Channel
    StreamGenerate(prompt string, tools []Tool) (<-chan *LLMChunk, error)
}

type InteractionInput struct {
	UserID string
	Query  string
	// ... 其他客製化欄位
}
// - 它接受一個 outputChannel，用於將數據串流回 HTTP Handler。
// - 它應該在 Goroutine 中執行，並在完成時關閉 Channel。
type Agent interface {
	Name() string
	ProcessStream(input *InteractionInput, outputChannel chan *StreamChunk) error
}

// InteractionService 儲存所有 Agent (解決 undefined: InteractionService)
type InteractionService struct {
	Agents map[string]Agent
}

func (app *InteractionService) RegisterAgent(name string, agent Agent) {
	app.Agents[name] = agent
}

func NewInteractionService() (*InteractionService) {
	return &InteractionService{
		Agents: make(map[string]Agent),
	}
}