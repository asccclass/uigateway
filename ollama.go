package main

import (
   // "io"
	"fmt"
	"time"
    "bufio"
    "bytes"
    // "strings"
    "net/http"
   "io/ioutil"
    "encoding/json"
)

// 生成回應結構
type GenerateResponse struct {
   Model              string    `json:"model"`
   CreatedAt          time.Time `json:"created_at"`
   Message            Message    `json:"message"`
	DoneReason         string    `json:"done_reason"`
   Done               bool      `json:"done"`
   TotalDuration      int64     `json:"total_duration,omitempty"`
   LoadDuration       int64     `json:"load_duration,omitempty"`
   PromptEvalCount    int64     `json:"prompt_eval_count,omitempty"`
   PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
   EvalCount          int64     `json:"eval_count,omitempty"`
   EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// OllamaFunctionCallArgs 匹配 tool_calls.function.arguments 內的 JSON 字串
type OllamaFunctionCallArgs struct {
    Arguments json.RawMessage `json:"arguments"` // 這是個 JSON 字串，需要二次解析
    Name string `json:"name"`          // 實際的工具名稱
}

// OllamaToolCallFunction 匹配 tool_calls 內的 function 物件
type OllamaToolCallFunction struct {
    Function OllamaFunctionCallArgs `json:"function"`
}

// OllamaTool 和 Function Calling 結構 (根據 Ollama 的最新 API 調整)
type OllamaTool struct {
    Type     string `json:"type"` // 必須是 "function"
    Function struct {
        Name        string                 `json:"name"`
        Description string                 `json:"description"`
        Parameters  map[string]interface{} `json:"parameters"`
    } `json:"function"`
}

// OllamaGenerateRequest 定義 Ollama /api/generate 的請求體
type OllamaGenerateRequest struct {
	Model    string   `json:"model"`
	Prompt   string   `json:"prompt"`
	Stream   bool     `json:"stream"`
	System   string   `json:"system"`
   Tools    []OllamaTool `json:"tools,omitempty"` // Ollama Tool/Function Calling 應該要移到MCP_Host內
}

// OllamaGenerateResponse 定義 Ollama 串流回應的單個 JSON 行
type OllamaGenerateResponse struct {
	Model     string     `json:"model"`
   CreatedAt time.Time  `json:"created_at"`
	Done      bool       `json:"done"`
	Response  string     `json:"response"` // 文本內容
   // Ollama 的 Function Calling 回應結構較為複雜，這裡只取關鍵部分
   TotalDuration int64  `json:"total_duration,omitempty"`
   ToolCalls []OllamaToolCallFunction `json:"tool_calls"`
   DoneReason string    `json:"done_reason,omitempty"`
}

// Ollama 相關的配置和狀態實現 LLMProviderClient 介面
type Ollama struct {
	Name string     // 客戶端名稱，實現 LLMName()
   URL  string    // Ollama 服務的基礎 URL，例如 "http://localhost:11434"
	Model      string // 使用的模型名稱 (例如: "llama3" 或 "qwen")
   SystemPrompt string // 實際 LLM 請求中的 System Prompt
}

// LLMProviderClient 介面實現
func(app *Ollama) GetURL() string { return app.URL }
func(app *Ollama) GetModel() string { return app.Model }
func(app *Ollama) GetSystemPrompt() string { return app.SystemPrompt }
func (app *Ollama) LLMName() string { return app.Name }

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

// 送出給Ollams（需要每個 LLM 自行 Implement，因為 API 網址不同）
func(app *Ollama) Send2LLM(jsonData string, isImage bool)(string, error) {
   resp, err := http.Post(app.URL+"/api/chat", "application/json", bytes.NewBuffer([]byte(jsonData))) // 發送請求給 Ollama  
   if err != nil {
      return "", fmt.Errorf("發送請求失敗: %s", err.Error())
   }
   defer resp.Body.Close()
   if resp.StatusCode != http.StatusOK { // 如果狀態碼不是 200 OK，則返回錯誤
      body, _ := ioutil.ReadAll(resp.Body)
      return "", fmt.Errorf("%s生成回應失敗，狀態碼: %d\n%s", app.URL+"/api/chat",resp.StatusCode, string(body))
   }

   // 解析回應
   var genResp GenerateResponse
   if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
      fmt.Printf("Send2LLM DECODE ERROR:解析回應失敗內容: %s\n", err.Error())  // 偵錯用 Log：觀察回應內容
      return "", fmt.Errorf("解析回應失敗: %s", err.Error())
   }
   return genResp.Message.Content, nil
}


func(o *Ollama) StreamGenerate(prompt, userPrompt string) (<-chan *LLMChunk, error) {
    output := make(chan *LLMChunk)
    o.Model = "gpt-oss:20b"  // modelName = "gpt-oss:20b"  // 預設模型名稱
    go func() {
        defer close(output) // 確保無論走哪條路徑 (MCP 或 Ollama 或 錯誤)，最後都會關閉 Channel
        // 1. 嘗試使用 MCP 工具
        if len(McpHost.ConnectedServers) > 0 {            
            toolsResult, err := CheckTools(userPrompt, o)  // 在 Goroutine 內部執行耗時的 CheckTools            
            if err == nil && toolsResult != "" {  // 若 MCP 成功處理，將結果寫入 Channel                
                output <- &LLMChunk{Text: toolsResult}                 
                return  // 任務完成，直接結束 Goroutine (不執行後續 Ollama 請求)
            }
            fmt.Println("偵錯用：MCP 無匹配或無結果，轉由一般 Ollama 處理 ->", userPrompt)   // 若 MCP 無結果，印出 Log 並繼續往下走
        }
        // 2. 執行一般 Ollama 請求 (Fallback)
        reqBody := OllamaGenerateRequest{
            Model:  o.Model,
            Prompt: prompt,
            Stream: true,
            System: o.SystemPrompt,
        }
        jsonData, err := json.Marshal(reqBody)
        if err != nil {
            fmt.Printf("Marshalling request failed: %v\n", err)
            return // 發生錯誤直接結束，defer 會關閉 channel
        }
        // 發送 HTTP 請求
        resp, err := PostData2HTTP(o.GetURL()+"/api/generate", jsonData)
        if err != nil {
            fmt.Printf("Ollama HTTP request failed: %v\n", err)
            return
        }
        defer resp.Body.Close()
        // 開始讀取 Ollama 串流
        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            line := scanner.Bytes()
            if len(line) == 0 {
                continue
            }

            var ollamaChunk OllamaGenerateResponse
            if err := json.Unmarshal(line, &ollamaChunk); err != nil {
                fmt.Printf("JSON Unmarshal Error: %s\n", err.Error())
                continue
            }

            // a. 傳輸文本塊
            if ollamaChunk.Response != "" {
                output <- &LLMChunk{Text: ollamaChunk.Response}
            }
            // b. 檢查 Ollama 原生 Tool Call (非 MCP)
            if len(ollamaChunk.ToolCalls) > 0 {
                ollamaToolCall := ollamaChunk.ToolCalls[0]
                toolName := ollamaToolCall.Function.Name
                
                if toolName != "" {
                    var args map[string]interface{}
                    if err := json.Unmarshal(ollamaToolCall.Function.Arguments, &args); err == nil {
                        output <- &LLMChunk{
                            ToolCall: &ToolCall{
                                Name: toolName,
                                Args: args,
                            },
                        }
                    }
                }
            }
            if ollamaChunk.Done {  // c. 檢查是否結束
                return
            }
        }
    }()
    return output, nil
}

func(o *Ollama) StreamGenerate_OLD(prompt, userPrompt string) (<-chan *LLMChunk, error) {
    output := make(chan *LLMChunk)
    o.Model = "gpt-oss:20b"  // modelName = "gpt-oss:20b"  // 預設模型名稱

    // 1. 構建 Ollama 請求體
    reqBody := OllamaGenerateRequest{
		Model:    o.Model,
		Prompt:   prompt,
		Stream:   true, // 必須開啟串流
		System:   o.SystemPrompt,
	}
   var err error
   toolsResult := ""
   
   isToolNeeded := 0  // 是否需要工具協助執行
   if len(McpHost.ConnectedServers) > 0 { // 先判斷使用者的問題是否需要用到MCP工具回答
      toolsResult, err = CheckTools(userPrompt, o)  // (map[string]interface, error)    // MCP 工具套用
      if toolsResult != "" && err == nil {
         isToolNeeded = 1
      } else {
         fmt.Println("偵錯用：", userPrompt) // 偵錯用
      }
   }
    
    // 2. 創建 HTTP 請求
   jsonData, err := json.Marshal(reqBody)   // 需要將我們 Agent 框架的 Tool 介面轉換成 OllamaTool 結構
	if err != nil {
		close(output)
		return nil, fmt.Errorf("Marshalling request failed: %w", err)
	}
   resp, err := PostData2HTTP(o.GetURL() + "/api/generate", jsonData)
   if err != nil {
      close(output)
      return nil, fmt.Errorf("Ollama HTTP request failed: %w", err)
   }
    
    go func() {
        defer resp.Body.Close()
        defer close(output)
        
		  scanner := bufio.NewScanner(resp.Body)  // 使用 bufio.Scanner 按行讀取串流 (JSONL 格式)
        for scanner.Scan() {
            line := scanner.Bytes()
            if len(line) == 0 {
				   continue // 跳過空行
            }
         // fmt.Printf("Ollama Stream Line: %s\n", line)  // --- 偵錯用 Log：觀察每一行原始輸出 ---
         var ollamaChunk OllamaGenerateResponse
         if err := json.Unmarshal(line, &ollamaChunk); err != nil {  // 可以在這裡發送錯誤 chunk 到 output，或記錄日誌				
            fmt.Printf("JSON Unmarshal Error on line: %s, Error: %s\n", string(line), err.Error())
				continue 
			}
         // a. 傳輸文本塊 (正常串流輸出)
			if ollamaChunk.Response != "" {
				output <- &LLMChunk{Text: ollamaChunk.Response}
			}         
         
         if isToolNeeded == 1 {  // b. 檢查 Tool Call
            // 我們假設只會有一個 Tool Call (許多模型都是這樣設計的)
            ollamaToolCall := ollamaChunk.ToolCalls[0]
               
            // 1. 檢查並處理 Tool Call 的名稱
            toolName := ollamaToolCall.Function.Name
            if toolName == "" {
               fmt.Println("Error: Tool name is empty in Ollama response.")
               continue
            }
            var args map[string]interface{}  // 2. 解析 Arguments 直接對 json.RawMessage (ollamaToolCall.Function.Arguments) 進行 Unmarshal
            if err := json.Unmarshal(ollamaToolCall.Function.Arguments, &args); err != nil {
               fmt.Printf("Tool Arguments Unmarshal Error (RawMessage): %v\n", err)
               continue
            }                
            // 將結構轉換為我們框架的 ToolCall 結構
            output <- &LLMChunk{
               ToolCall: &ToolCall{
                  Name: toolName, 
                  Args: args, 
               },
            }                
            // 發現 Tool Call，立即結束 LLM 串流讀取，觸發 Agent 執行工具
            // 雖然 Ollama 將 Tool Call 放在 done: true 的行中，但為了確保 Agent 即時反應，我們在這裡加入 return。
            // return // 這裡可以選擇是否立即 return，視乎 Ollama 是否將 Tool Call 放在 streaming 中間
         }                     
         // c. 檢查是否結束
         if ollamaChunk.Done {
            // 如果 Tool Call 發生在 Done 的同一行，這會是結束 LLM 串流的信號
            fmt.Println("Ollama stream finished.")
            return 
         }
      }
    }()
    return output, nil
}