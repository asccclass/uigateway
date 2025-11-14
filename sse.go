package main

import (
	"fmt"
	"net/http"
    "encoding/json"	 
)

type SSEService struct{}

// AgentStreamHandler 處理來自 Agent 的 SSE 串流
func (app *SSEService) SseAgentStreamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")  // 設置 SSE 必要的 Header
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}
    ctx := r.Context()

    userID := r.URL.Query().Get("user_id")
    query := r.URL.Query().Get("query")
    agentName := "chat" 
    
    if query == "" {
        fmt.Fprintf(w, "event: error\ndata: {\"message\": \"Query cannot be empty\"}\n\n")
        flusher.Flush()
        return
    }

	// 確保 chatService 不為 nil
	if chatService == nil {
		fmt.Fprintf(w, "event: error\ndata: {\"message\": \"Service not initialized\"}\n\n")
        flusher.Flush()
		return
	}
    
	agent, ok := chatService.Agents[agentName]
	if !ok {
        fmt.Fprintf(w, "event: error\ndata: {\"message\": \"Agent not found\"}\n\n")
        flusher.Flush()
        return
	}
	outputChannel := make(chan *StreamChunk)   // 創建 Channel 作為 Agent 輸出和 SSE 寫入之間的管道
   input := &InteractionInput{
        UserID: userID,
        Query:  query,
   }

    // 在 Goroutine 中啟動 Agent 的串流處理 (非同步執行 Agent 邏輯)
	go func() {
		err := agent.ProcessStream(input, outputChannel)
		if err != nil {
			fmt.Printf("Agent process error: %v\n", err)
            outputChannel <- &StreamChunk{   // 如果 Agent 處理發生錯誤，發送錯誤事件給前端
                Type: "error",
                Data: map[string]string{"message": fmt.Sprintf("Agent internal error: %s", err.Error())},
            }
		}
	}()
L:    // 監聽 Channel 並將數據串流給客戶端
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Client disconnected.")   // 客戶端斷開連接
			return
		case chunk, ok := <-outputChannel:
			if !ok {
				break L   // Channel 已關閉，表示 Agent 處理完成
			}
			// 處理 chunk 的數據格式。注意: 這裡使用 Marshal，確保即使 Data 是字串，也會被轉為 JSON 字串
			dataBytes, _ := json.Marshal(chunk.Data)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", chunk.Type, string(dataBytes))  // 寫入 SSE 格式：event: Type \n data: Data \n\n
			flusher.Flush()
		}
	}
    fmt.Println("Agent stream finished.")
}

func(app *SSEService) AddRouter(router *http.ServeMux) {
   router.HandleFunc("/events", app.SseAgentStreamHandler)
}

func NewSSEService()(*SSEService) {
	return &SSEService{}
}