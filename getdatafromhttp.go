package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// 發起請求
	client := &http.Client{Timeout: 0} // 設置為 0 表示不限制超時，避免串流中斷
	resp, err := client.Do(req)
	if err != nil {
		// resp is nil here, do not close body
		fmt.Printf("Ollama HTTP Request FAILED(PostData2HTTP): %s\n", err.Error())
		return nil, fmt.Errorf("HTTP request failed to %s: %w", url, err)
	}
	if resp.StatusCode != http.StatusOK { // 如果狀態碼不是 200 OK，則返回錯誤
		defer resp.Body.Close()
		errorBody, _ := io.ReadAll(resp.Body)
		fmt.Printf("PostData2HTTP API Error. Status: %d, Response: %s\n", resp.StatusCode, errorBody)
		return nil, fmt.Errorf("Ollama API returned non-200 status: %d", resp.StatusCode)
	}
	return resp, nil
}
