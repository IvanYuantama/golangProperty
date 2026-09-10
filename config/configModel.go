package config

const RadiusDegrees = 0.027

type LayerStyle struct {
	Color string `json:"color"`
	Label string `json:"label"`
}

type AnalysisQuery struct {
	Table string `json:"table"`
	Layer string `json:"layer"`
	ValueColumn string `json:"valueColumn"`
	SearchType string `json:"searchType"`
}

type LayerValueStylesMap map[string]map[string]LayerStyle

var LayerDefaultStyles = map[string]LayerStyle{
	"default": {
		Color: "#8E8E93",
		Label: "Unknown",
	},
	"population": {
		Color: "#5AC8FA",
		Label: "Unknown",
	},
	"wifi": {
		Color: "#5AC8FA",
		Label: "Unknown",
	},
	"mobile_data": {
		Color: "#5AC8FA",
		Label: "Unknown",
	},
	"crime": {
		Color: "#FF9500",
		Label: "Unknown",
	},
}

var LayerValueStyles = LayerValueStylesMap {
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
	
var AnalysisQueriesList = []AnalysisQuery{
	{
		Table:       "flood_zones",
		Layer:       "flood",
		ValueColumn: "area",
		SearchType:  "intersect",
	},
	{
		Table:       "temperature_zones",
		Layer:       "temperature",
		ValueColumn: "suhu",
		SearchType:  "intersect",
	},
	{
		Table:       "air_quality",
		Layer:       "air_quality",
		ValueColumn: "polusi",
		SearchType:  "intersect",
	},
	{
		Table:       "green_spaces",
		Layer:       "green_spaces",
		ValueColumn: "rth",
		SearchType:  "intersect",
	},
	{
		Table:       "population_data",
		Layer:       "population",
		ValueColumn: "jumlah_pen::text",
		SearchType:  "intersect",
	},
	{
		Table:       "elevation_zones",
		Layer:       "elevation",
		ValueColumn: "ketinggian::text",
		SearchType:  "intersect",
	},
	{
		Table:       "roads_buffer",
		Layer:       "roads_buffer",
		ValueColumn: "remark",
		SearchType:  "radius",
	},
	{
		Table:       "public_facilities",
		Layer:       "public_facilities",
		ValueColumn: "kategori",
		SearchType:  "grouped",
	},
	{
		Table:       "wifi_zones",
		Layer:       "wifi",
		ValueColumn: "(dl_mbps::text || '|' || ul_mbps::text)",
		SearchType:  "intersect",
	},
	{
		Table:       "mobile_data_zones",
		Layer:       "mobile_data",
		ValueColumn: "(dl_mbps::text || '|' || ul_mbps::text)",
		SearchType:  "intersect",
	},
	{
		Table:       "crime_zones",
		Layer:       "crime",
		ValueColumn: "crime_total::text",
		SearchType:  "intersect",
	},
}