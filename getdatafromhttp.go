package main

import (
	"io"
	"fmt"
	"time"
	"bytes"
   "net/http"
   "io/ioutil"
)

func GetDataFromHTTP(url string) (string, error) {
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(body), nil
}

func PostData2HTTP(url string, jsonData []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url + "/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")    
    // 發起請求
	client := &http.Client{Timeout: 30 * time.Second} // 設置超時
	resp, err := client.Do(req)
	if err != nil {
		defer resp.Body.Close()
      fmt.Printf("Ollama HTTP Request FAILED: %s\n", err.Error())
		return nil, fmt.Errorf("HTTP request failed to %s: %w", url, err)
	}
	if resp.StatusCode != http.StatusOK { // 如果狀態碼不是 200 OK，則返回錯誤
		defer resp.Body.Close()
		errorBody, _ := io.ReadAll(resp.Body)		
		// --- 【重要】檢查這裡的 Log ---
		fmt.Printf("Ollama API Error Status: %d, Body: %s\n", resp.StatusCode, errorBody)        
		return nil, fmt.Errorf("Ollama API returned non-200 status: %d", resp.StatusCode)
	}
	return resp, nil
}