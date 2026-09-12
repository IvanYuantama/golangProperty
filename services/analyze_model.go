package services

import "encoding/json"

type analysisRow struct {
	Layer          string
	StyleValue     *string
	DistanceMeters float64
	Total          *int64
	Attributes     json.RawMessage
}

type AnalysisResult struct {
	Layer          string          `json:"layer"`
	DistanceMeters int64           `json:"distance_meters"`
	Label          string          `json:"label"`
	Color          string          `json:"color"`
	Total          *int64          `json:"total,omitempty"`
	Attributes     json.RawMessage `json:"attributes"`
}
