package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"expressXgolang/golang/config"
	"expressXgolang/golang/db"
)

type RawQueryResult struct {
	Layer          string                 `json:"layer"`
	ZoneValue      *string                `json:"zone_value"`
	DistanceMeters float64                `json:"distance_meters"`
	Total          *int64                 `json:"total,omitempty"`
	Data           map[string]interface{} `json:"data"`
}

type AnalysisResult struct {
	Layer          string                 `json:"layer"`
	DistanceMeters int64                  `json:"distance_meters"`
	Label          string                 `json:"label"`
	Color          string                 `json:"color"`
	Total          *int64                 `json:"total,omitempty"`
	Attributes     map[string]interface{} `json:"attributes"`
}

const optimizedQuery = `
WITH coordinates AS NOT MATERIALIZED (
  SELECT
    $1::double precision AS latitude,
    $2::double precision AS longitude,
    $3::double precision AS radius_degrees
),
params AS NOT MATERIALIZED (
  SELECT
    c.*,
    ST_SetSRID(ST_Point(c.longitude, c.latitude), 4326) AS search_point
  FROM coordinates AS c
),
results AS (
  SELECT *
  FROM (
    SELECT
      1 AS sort_order,
      'flood'::text AS layer,
      t.area::text AS zone_value,
      0::double precision AS distance_meters,
      NULL::bigint AS total,
      jsonb_build_object('area', t.area) AS data
    FROM flood_zones AS t
    CROSS JOIN params AS p
    WHERE ST_Intersects(t.wkb_geometry, p.search_point)
    LIMIT 1
  ) AS flood

  UNION ALL

  SELECT *
  FROM (
    SELECT
      2 AS sort_order,
      'temperature'::text AS layer,
      t.suhu::text AS zone_value,
      0::double precision AS distance_meters,
      NULL::bigint AS total,
      jsonb_build_object('suhu', t.suhu) AS data
    FROM temperature_zones AS t
    CROSS JOIN params AS p
    WHERE ST_Intersects(t.wkb_geometry, p.search_point)
    LIMIT 1
  ) AS temperature

  UNION ALL

  SELECT *
  FROM (
    SELECT
      3 AS sort_order,
      'air_quality'::text AS layer,
      t.polusi::text AS zone_value,
      0::double precision AS distance_meters,
      NULL::bigint AS total,
      jsonb_build_object(
        'objectid', t.objectid,
        'shape_leng', t.shape_leng,
        'shape_area', t.shape_area,
        'polusi', t.polusi
      ) AS data
    FROM air_quality AS t
    CROSS JOIN params AS p
    WHERE ST_Intersects(t.wkb_geometry, p.search_point)
    LIMIT 1
  ) AS air_quality

  UNION ALL

  SELECT *
  FROM (
    SELECT
      4 AS sort_order,
      'green_spaces'::text AS layer,
      t.rth::text AS zone_value,
      0::double precision AS distance_meters,
      NULL::bigint AS total,
      jsonb_build_object(
        'id', t.id,
        'gridcode', t.gridcode,
        'rth', t.rth
      ) AS data
    FROM green_spaces AS t
    CROSS JOIN params AS p
    WHERE ST_Intersects(t.wkb_geometry, p.search_point)
    LIMIT 1
  ) AS green_spaces

  UNION ALL

  SELECT *
  FROM (
    SELECT
      5 AS sort_order,
      'population'::text AS layer,
      t.jumlah_pen::text AS zone_value,
      0::double precision AS distance_meters,
      NULL::bigint AS total,
      jsonb_build_object(
        'nama_kab', t.nama_kab,
        'nama_kec', t.nama_kec,
        'jumlah_pen', t.jumlah_pen,
        'jumlah_kk', t.jumlah_kk,
        'islam', t.islam,
        'kristen', t.kristen,
        'katholik', t.katholik,
        'hindu', t.hindu,
        'budha', t.budha,
        'konghucu', t.konghucu,
        'kepercayaa', t.kepercayaa,
        'pria', t.pria,
        'wanita', t.wanita,
        'kawin', t.kawin
      ) || jsonb_build_object(
        'belum_kawi', t.belum_kawi,
        'cerai_hidu', t.cerai_hidu,
        'cerai_mati', t.cerai_mati,
        'u0', t.u0,
        'u5', t.u5,
        'u10', t.u10,
        'u15', t.u15,
        'u20', t.u20,
        'u25', t.u25,
        'u30', t.u30,
        'u35', t.u35,
        'u40', t.u40,
        'u45', t.u45,
        'u50', t.u50
      ) || jsonb_build_object(
        'u55', t.u55,
        'u60', t.u60,
        'u65', t.u65,
        'u70', t.u70,
        'u75', t.u75,
        'tidak_blm_', t.tidak_blm_,
        'belum_tama', t.belum_tama,
        'tamat_sd', t.tamat_sd,
        'sltp', t.sltp,
        'slta', t.slta,
        'd1_dan_d2', t.d1_dan_d2,
        'd3', t.d3,
        's1', t.s1,
        's2', t.s2
      ) || jsonb_build_object(
        's3', t.s3,
        'belum_tida', t.belum_tida,
        'pensiunan', t.pensiunan,
        'mengurus_r', t.mengurus_r,
        'perdaganga', t.perdaganga,
        'perawat', t.perawat,
        'nelayan', t.nelayan,
        'pelajar_ma', t.pelajar_ma,
        'guru', t.guru,
        'wiraswasta', t.wiraswasta,
        'pengacara', t.pengacara,
        'lainnya', t.lainnya
      ) AS data
    FROM population_data AS t
    CROSS JOIN params AS p
    WHERE ST_Intersects(t.wkb_geometry, p.search_point)
    LIMIT 1
  ) AS population

  UNION ALL

  SELECT *
  FROM (
    SELECT
      6 AS sort_order,
      'elevation'::text AS layer,
      t.ketinggian::text AS zone_value,
      0::double precision AS distance_meters,
      NULL::bigint AS total,
      jsonb_build_object('ketinggian', t.ketinggian) AS data
    FROM elevation_zones AS t
    CROSS JOIN params AS p
    WHERE ST_Intersects(t.wkb_geometry, p.search_point)
    LIMIT 1
  ) AS elevation

  UNION ALL

  SELECT *
  FROM (
    SELECT
      7 AS sort_order,
      'roads_buffer'::text AS layer,
      t.remark::text AS zone_value,
      ST_Distance(t.wkb_geometry, p.search_point) * 111320 AS distance_meters,
      NULL::bigint AS total,
      jsonb_build_object(
        'objectid_1', t.objectid_1,
        'remark', t.remark,
        'shape_leng', t.shape_leng,
        'shape_area', t.shape_area
      ) AS data
    FROM roads_buffer AS t
    CROSS JOIN params AS p
    WHERE ST_DWithin(t.wkb_geometry, p.search_point, p.radius_degrees)
    ORDER BY t.wkb_geometry <-> p.search_point
    LIMIT 5
  ) AS roads_buffer

  UNION ALL

  SELECT
    8 AS sort_order,
    'public_facilities'::text AS layer,
    t.kategori::text AS zone_value,
    MAX(ST_Distance(t.wkb_geometry, p.search_point) * 111320) AS distance_meters,
    COUNT(*)::bigint AS total,
    jsonb_build_object('kategori', t.kategori) AS data
  FROM public_facilities AS t
  CROSS JOIN params AS p
  WHERE ST_DWithin(t.wkb_geometry, p.search_point, p.radius_degrees)
  GROUP BY t.kategori, p.search_point

  UNION ALL

  SELECT *
  FROM (
    SELECT
      9 AS sort_order,
      'wifi'::text AS layer,
      (t.dl_mbps::text || '|' || t.ul_mbps::text) AS zone_value,
      0::double precision AS distance_meters,
      NULL::bigint AS total,
      jsonb_build_object(
        'quadkey', t.quadkey,
        'avg_d_kbps', t.avg_d_kbps,
        'avg_u_kbps', t.avg_u_kbps,
        'avg_lat_ms', t.avg_lat_ms,
        'tests', t.tests,
        'devices', t.devices,
        'dl_mbps', t.dl_mbps,
        'ul_mbps', t.ul_mbps
      ) AS data
    FROM wifi_zones AS t
    CROSS JOIN params AS p
    WHERE ST_Intersects(t.wkb_geometry, p.search_point)
    LIMIT 1
  ) AS wifi

  UNION ALL

  SELECT *
  FROM (
    SELECT
      10 AS sort_order,
      'mobile_data'::text AS layer,
      (t.dl_mbps::text || '|' || t.ul_mbps::text) AS zone_value,
      0::double precision AS distance_meters,
      NULL::bigint AS total,
      jsonb_build_object(
        'quadkey', t.quadkey,
        'avg_d_kbps', t.avg_d_kbps,
        'avg_u_kbps', t.avg_u_kbps,
        'avg_lat_ms', t.avg_lat_ms,
        'tests', t.tests,
        'devices', t.devices,
        'dl_mbps', t.dl_mbps,
        'ul_mbps', t.ul_mbps
      ) AS data
    FROM mobile_data_zones AS t
    CROSS JOIN params AS p
    WHERE ST_Intersects(t.wkb_geometry, p.search_point)
    LIMIT 1
  ) AS mobile_data

  UNION ALL

  SELECT *
  FROM (
    SELECT
      11 AS sort_order,
      'crime'::text AS layer,
      t.crime_total::text AS zone_value,
      0::double precision AS distance_meters,
      NULL::bigint AS total,
      jsonb_build_object(
        'nama_kab', t.nama_kab,
        'jumlah_pen', t.jumlah_pen,
        'jumlah_kk', t.jumlah_kk,
        'tahun', t.tahun,
        'crime_total', t.crime_total,
        'crime_cleared', t.crime_cleared,
        'crime_rate', t.crime_rate,
        'clearance_rate', t.clearance_rate,
        'sumber', t.sumber
      ) AS data
    FROM crime_zones AS t
    CROSS JOIN params AS p
    WHERE ST_Intersects(t.wkb_geometry, p.search_point)
    LIMIT 1
  ) AS crime
)
SELECT
  layer,
  zone_value,
  ROUND(distance_meters::numeric, 2)::float8 AS distance_meters,
  total,
  data::text AS data_json
FROM results
ORDER BY sort_order, distance_meters;
`

func getDefaultStyle(layer string) config.LayerStyle {
	if style, ok := config.LayerDefaultStyles[layer]; ok {
		return style
	}
	return config.LayerDefaultStyles["default"]
}

func formatPopulation(value *string) config.LayerStyle {
	defaultStyle := getDefaultStyle("population")
	if value == nil {
		return defaultStyle
	}

	pop, err := strconv.ParseFloat(*value, 64)
	if err != nil || math.IsNaN(pop) || math.IsInf(pop, 0) {
		return defaultStyle
	}

	return config.LayerStyle{
		Color: defaultStyle.Color,
		Label: fmt.Sprintf("%d people", int64(math.Round(pop))),
	}
}

func formatNetworkSpeed(layer string, value *string) config.LayerStyle {
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

	return config.LayerStyle{
		Color: defaultStyle.Color,
		Label: fmt.Sprintf("Download %s, Upload %s", dlLabel, ulLabel),
	}
}

func formatCrime(value *string) config.LayerStyle {
	defaultStyle := getDefaultStyle("crime")
	if value == nil {
		return defaultStyle
	}

	total, err := strconv.ParseFloat(*value, 64)
	if err != nil || math.IsNaN(total) || math.IsInf(total, 0) {
		return defaultStyle
	}

	return config.LayerStyle{
		Color: defaultStyle.Color,
		Label: fmt.Sprintf("%d cases", int64(math.Round(total))),
	}
}

func getLayerStyle(layer string, value *string) config.LayerStyle {
	switch layer {
	case "population":
		return formatPopulation(value)
	case "wifi", "mobile_data":
		return formatNetworkSpeed(layer, value)
	case "crime":
		return formatCrime(value)
	}

	defaultStyle := getDefaultStyle(layer)
	valStr := ""
	if value != nil {
		valStr = *value
	}

	if subMap, ok := config.LayerValueStyles[layer]; ok {
		if style, exists := subMap[valStr]; exists {
			return style
		}
	}

	label := valStr
	if label == "" {
		label = defaultStyle.Label
	}

	return config.LayerStyle{
		Color: defaultStyle.Color,
		Label: label,
	}
}

func formatResult(row RawQueryResult) AnalysisResult {
	style := getLayerStyle(row.Layer, row.ZoneValue)
	return AnalysisResult{
		Layer:          row.Layer,
		DistanceMeters: int64(math.Round(row.DistanceMeters)),
		Label:          style.Label,
		Color:          style.Color,
		Total:          row.Total,
		Attributes:     row.Data,
	}
}

func AnalyzeLocation(ctx context.Context, latitude, longitude float64) ([]AnalysisResult, error) {
	// $1 = latitude, $2 = longitude, $3 = radius_degrees
	rows, err := db.Pool.Query(ctx, optimizedQuery, latitude, longitude, config.RadiusDegrees)
	if err != nil {
		return nil, fmt.Errorf("failed to execute spatial analysis query: %w", err)
	}
	defer rows.Close()

	results := make([]AnalysisResult, 0, 16)

	for rows.Next() {
		var r RawQueryResult
		var rawJSON string

		err := rows.Scan(
			&r.Layer,
			&r.ZoneValue,
			&r.DistanceMeters,
			&r.Total,
			&rawJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if rawJSON != "" {
			_ = json.Unmarshal([]byte(rawJSON), &r.Data)
		}

		results = append(results, formatResult(r))
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return results, nil
}
