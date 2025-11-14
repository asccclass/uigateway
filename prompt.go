package main

import (
	"fmt"
)

// buildInitialPrompt 構建發送給 LLM 的完整 Prompt
// 實務上，這裡還會加入歷史紀錄、Tool Schema 等
func (app *CustomChatAgent) buildInitialPrompt(input *InteractionInput) string {
    // 這裡我們將 System Prompt 和用戶的最新查詢結合起來
    prompt := app.SystemPrompt + "\n\n"
    
    // 這裡可以加入歷史對話紀錄 (如果有的話)
    // prompt += "History: [User: ..., Agent: ...]\n"
    
    // 加入當前查詢
    prompt += fmt.Sprintf("User Query: %s", input.Query)
    
    // 如果有工具，這裡還會加入工具的 FunctionSchema 描述
    // prompt += "\n\nAvailable Tools:\n" + a.getToolSchemas()
    
    return prompt
}