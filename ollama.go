package main

import (
	"fmt"
	"time"
	"strings"
    "net/http"
    "encoding/json"
)

// Ollama 相關的配置和狀態
type Ollama struct {
	Name string
    URL  string    // Ollama 服務的基礎 URL，例如 "http://localhost:11434"
}

func (app *Ollama) LLMName() string { 
	return app.Name 
}

// 獲取所有可用模型
func(app *Ollama) GetModels() ([]Model, error) {
   // 這樣可以保留 DefaultTransport 內建的撥號逾時、TLS 握手逾時等合理設置
	transport := http.DefaultTransport.(*http.Transport).Clone()
   // 如果你需要設置代理或自定義 TLS 配置，在這裡添加
	// proxyURL, _ := url.Parse("http://your-proxy-server:8080")
	// transport.Proxy = http.ProxyURL(proxyURL)
	// transport.TLSHandshakeTimeout = 10 * time.Second // 已經在 DefaultTransport 中預設
   // 可以根據需要設定其他 Client 參數，例如 CheckRedirect、Timeout 等
   client := &http.Client{
      Transport: transport,
      Timeout: 60 * time.Second, // 整個請求的逾時時間
   }
   resp, err := client.Get(app.URL + "/api/tags")
   if err != nil {
      return nil, fmt.Errorf("連接到 Ollama 服務失敗: %v", err)
   }
   defer resp.Body.Close()
   if resp.StatusCode != http.StatusOK {
      return nil, fmt.Errorf("獲取模型列表失敗，狀態碼: %d", resp.StatusCode)
   }
   var listResp ModelsWrapper
   if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
      return nil, fmt.Errorf("解析回應失敗: %v", err)
   }
   return listResp.Models, nil
}

func (m *Ollama) StreamGenerate(prompt string, tools []Tool) (<-chan *LLMChunk, error) {
    output := make(chan *LLMChunk)
    
    go func() {
        defer close(output)
        // 只有當 Prompt 中包含「天氣」且不包含「Observation」時，才需要呼叫工具。
        // 如果包含 Observation，則直接輸出最終答案。
        isObservationPresent := strings.Contains(prompt, "Observation for tool")
        needsToolCall := strings.Contains(prompt, "天氣") && len(tools) > 0 && !isObservationPresent

        // 模擬文字輸出
        if needsToolCall {
            // 狀態 A: 需要呼叫工具 (因為有「天氣」且沒有 Observation)
            // 如果需要工具呼叫，文字回應會比較簡短，或是引導語句
            responseChunks := []string{"好的，", "我", "需要", "先", "查詢", "資料。"}
            for _, text := range responseChunks {
                output <- &LLMChunk{Text: text}
                time.Sleep(time.Millisecond * 30)
            }
            // 這裡發送 ToolCall Chunk
            fmt.Println("--> Mock LLM Client: Simulating a tool call for weather.")
            output <- &LLMChunk{
                ToolCall: &ToolCall{
                    Name: "get_current_weather",
                    Args: map[string]interface{}{"location": "Taipei"},
                },
            }
            // 確保 ToolCall 發送後沒有多餘的文字輸出，否則 Agent 會讀到文字並認為回應結束
        } else {
            // 狀態 B: 不需要呼叫工具 (進入最終回應階段，無論是因為有 Observation 還是沒有「天氣」)
            finalAnswer := "我已經處理了您的請求，"
            if isObservationPresent {
                // 如果有 Observation，模擬輸出最終答案
                finalAnswer = "根據我的查詢，台北的天氣目前是 25°C，陽光明媚。還有什麼可以幫助您的嗎？"
            } else {
                // 如果沒有 Observation 且沒有「天氣」，輸出普通回答
                finalAnswer = "請提供更多細節。"
            }
            for _, text := range strings.Split(finalAnswer, "") {
                output <- &LLMChunk{Text: text}
                time.Sleep(time.Millisecond * 10)
            }
        }
        time.Sleep(time.Millisecond * 50) // 確保緩衝區清空（channel 本身是同步，但I/O需要時間）
    }()
    return output, nil
}