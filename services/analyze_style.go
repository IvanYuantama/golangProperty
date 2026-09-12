package services

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type layerStyle struct {
	Color string
	Label string
}

var layerDefaultStyles = map[string]layerStyle{
	"default":     {Color: "#8E8E93", Label: "Unknown"},
	"population":  {Color: "#5AC8FA", Label: "Unknown"},
	"wifi":        {Color: "#5AC8FA", Label: "Unknown"},
	"mobile_data": {Color: "#5AC8FA", Label: "Unknown"},
	"crime":       {Color: "#FF9500", Label: "Unknown"},
}

var layerValueStyles = map[string]map[string]layerStyle{
	"flood": {
		"Low Flood Risk":  {Color: "#34C759", Label: "Low Flood Risk"},
		"High Flood Risk": {Color: "#FF3B30", Label: "High Flood Risk"},
	},
	"temperature": {
		"Cool":      {Color: "#00ff84", Label: "Cool"},
		"Moderate":  {Color: "#FFCC00", Label: "Moderate"},
		"Hot":       {Color: "#FF9500", Label: "Hot"},
		"Very Hot":  {Color: "#FF3B30", Label: "Very Hot"},
		"Very Cool": {Color: "#00ff84", Label: "Very Cool"},
	},
	"air_quality": {
		"Low":    {Color: "#34C759", Label: "Good"},
		"Medium": {Color: "#FF9500", Label: "Moderate"},
		"High":   {Color: "#FF3B30", Label: "Bad"},
	},
	"green_spaces": {
		"Dense":    {Color: "#34C759", Label: "Dense"},
		"Moderate": {Color: "#FFCC00", Label: "Moderate"},
		"Sparse":   {Color: "#FF9500", Label: "Sparse"},
	},
	"elevation": {
		"Lowland":  {Color: "#FF9500", Label: "Lowland"},
		"Midland":  {Color: "#34C759", Label: "Midland"},
		"Highland": {Color: "#5AC8FA", Label: "Highland"},
	},
}

func getDefaultStyle(layer string) layerStyle {
	if style, ok := layerDefaultStyles[layer]; ok {
		return style
	}
	return layerDefaultStyles["default"]
}

func formatPopulation(value *string) layerStyle {
	defaultStyle := getDefaultStyle("population")
	if value == nil {
		return defaultStyle
	}

	pop, err := strconv.ParseFloat(*value, 64)
	if err != nil || math.IsNaN(pop) || math.IsInf(pop, 0) {
		return defaultStyle
	}

	return layerStyle{
		Color: defaultStyle.Color,
		Label: fmt.Sprintf("%d people", int64(math.Round(pop))),
	}
}

func formatNetworkSpeed(layer string, value *string) layerStyle {
	defaultStyle := getDefaultStyle(layer)
	if value == nil || *value == "" {
		return defaultStyle
	}

	parts := strings.Split(*value, "|")
	var download, upload float64
	var errDl, errUl error

	if len(parts) > 0 {
		download, errDl = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	}
	if len(parts) > 1 {
		upload, errUl = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	}

	hasDl := errDl == nil && !math.IsNaN(download) && !math.IsInf(download, 0)
	hasUl := errUl == nil && !math.IsNaN(upload) && !math.IsInf(upload, 0)

	if !hasDl && !hasUl {
		return defaultStyle
	}

	dlLabel := "-"
	if hasDl {
		dlLabel = fmt.Sprintf("%.1f Mbps", download)
	}

	ulLabel := "-"
	if hasUl {
		ulLabel = fmt.Sprintf("%.1f Mbps", upload)
	}

	return layerStyle{
		Color: defaultStyle.Color,
		Label: fmt.Sprintf("Download %s, Upload %s", dlLabel, ulLabel),
	}
}

func formatCrime(value *string) layerStyle {
	defaultStyle := getDefaultStyle("crime")
	if value == nil {
		return defaultStyle
	}

	total, err := strconv.ParseFloat(*value, 64)
	if err != nil || math.IsNaN(total) || math.IsInf(total, 0) {
		return defaultStyle
	}

	return layerStyle{
		Color: defaultStyle.Color,
		Label: fmt.Sprintf("%d cases", int64(math.Round(total))),
	}
}

func getLayerStyle(layer string, value *string) layerStyle {
	switch layer {
	case "population":
		return formatPopulation(value)
	case "wifi", "mobile_data":
		return formatNetworkSpeed(layer, value)
	case "crime":
		return formatCrime(value)
	}

	defaultStyle := getDefaultStyle(layer)
	valueString := ""
	if value != nil {
		valueString = *value
	}

	if layerStyles, ok := layerValueStyles[layer]; ok {
		if style, exists := layerStyles[valueString]; exists {
			return style
		}
	}

	label := valueString
	if label == "" {
		label = defaultStyle.Label
	}

	return layerStyle{
		Color: defaultStyle.Color,
		Label: label,
	}
}

func formatResult(row analysisRow) AnalysisResult {
	style := getLayerStyle(row.Layer, row.StyleValue)
	return AnalysisResult{
		Layer:          row.Layer,
		DistanceMeters: int64(math.Round(row.DistanceMeters)),
		Label:          style.Label,
		Color:          style.Color,
		Total:          row.Total,
		Attributes:     row.Attributes,
	}
}
