// 輸出聲音
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// TTSRequest 接收前端傳來的 JSON
type TTSRequest struct {
	Text  string  `json:"text"`
	Voice string  `json:"voice"` // 選填，預設 af_bella
	Speed float64 `json:"speed"` // 選填，預設 1.0
}

// DockerAPIRequest 發送給 Kokoro Docker 的格式
type DockerAPIRequest struct {
	Model          string  `json:"model"`
	Input          string  `json:"input"`
	Voice          string  `json:"voice"`
	ResponseFormat string  `json:"response_format"`
	Speed          float64 `json:"speed"`
}

// API: 語音合成，格式: POST /speak，Body: {"text": "你好", "voice": "af_bella"}
func SpeakFromWeb(w http.ResponseWriter, r *http.Request) {
	var req TTSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}
	// 設定預設值
	if req.Voice == "" {
		req.Voice = "zf_xiaobei" //"af_bella" // 預設聲音
	}
	if req.Speed == 0 {
		req.Speed = 1.0
	}
	// 2. 準備呼叫 Docker API
	apiURL := os.Getenv("TTS_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8880/v1/audio/speech"
	}

	fmt.Println("apiURL: ", apiURL)

	dockerPayload := DockerAPIRequest{
		Model:          "kokoro",
		Input:          req.Text,
		Voice:          req.Voice,
		ResponseFormat: "mp3",
		Speed:          req.Speed,
	}

	jsonData, _ := json.Marshal(dockerPayload)
	// 設定 Timeout 防止連線卡死
	client := &http.Client{Timeout: 60 * time.Second} // 3. 發送請求到 Kokoro TTS Container
	proxyReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("TTS Service unavailable: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		http.Error(w, fmt.Sprintf("TTS Error: %s", body), resp.StatusCode)
		return
	}

	// 4. 設定回應標頭 (告知瀏覽器這是音訊)
	w.Header().Set("Content-Type", "audio/mpeg")
	//w.Header().Set("Content-Disposition", "attachment; filename=speech.mp3") // "attachment" 讓瀏覽器知道這是要下載的內容
	w.Header().Set("Content-Disposition", "inline; filename=speech.mp3") // "inline" 讓瀏覽器知道這是要直接播放的內容，而不是下載

	// 5. 串流傳輸 (Streaming Copy)，直接將 Docker 的回應寫入使用者的回應，不佔用伺服器記憶體
	if _, err := io.Copy(w, resp.Body); err != nil {
		fmt.Printf("Stream error: %v\n", err)
	}
}
