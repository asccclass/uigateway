package main

import (
	"time"
)

// Details represents the details of a model
type Details struct {
	ParentModel       string   `json:"parent_model"`
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// Model represents a single model's information 模型資訊結構
type Model struct {
	Name        string    `json:"name"`
	Model       string    `json:"model"`
	ModifiedAt  time.Time `json:"modified_at"`
	Size        int64     `json:"size"`
	Digest      string    `json:"digest"`
	Details     Details   `json:"details"`
	Description string    `json:"description,omitempty"`
	SizeGB      string    `json:"size_gb,omitempty"`   // 轉換後的大小
	ModiTime    string    `json:"modi_time,omitempty"` // 轉換後的時間
}

// ModelsWrapper wraps the models array
type ModelsWrapper struct {
	Models []Model `json:"models"`
}

// IndexTemplateData holds data for rendering index.html
type IndexTemplateData struct {
	ShowMessageAPI bool
	ShowTTSAPI     bool
	MessageAPIURL  string
}
