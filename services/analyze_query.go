package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
)

const (
	maxLayerResultLimit = 100
	// postgreSQL membatasi attribut (40) pada jumlah jsonb_build_object, makanya dipecah lalu digabungkan.
	jsonAttributesChunk = 40
)

var sqlIdentifierPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

func buildAnalyzeQuery(layers []LayerDefinition) (string, error) {
	if len(layers) == 0 {
		return "", fmt.Errorf("minimal satu layer analisis wajib dikonfigurasi")
	}

	branches := make([]string, 0, len(layers))
	seenLayers := make(map[string]struct{}, len(layers))

	for index, layer := range layers {
		if _, exists := seenLayers[layer.Name]; exists {
			return "", fmt.Errorf("nama layer %q terdaftar lebih dari satu kali", layer.Name)
		}
		seenLayers[layer.Name] = struct{}{}

		branch, err := buildLayerQuery(index+1, layer)
		if err != nil {
			return "", err
		}
		branches = append(branches, branch)
	}

	// Semua branch digabung dengan UNION ALL
	return `
WITH params AS NOT MATERIALIZED (
  SELECT
    ST_SetSRID(
      ST_Point($2::double precision, $1::double precision),
      4326
    ) AS search_point,
    $3::double precision AS radius_degrees
),
results AS (
` + strings.Join(branches, "\n\n  UNION ALL\n\n") + `
)
SELECT
  layer,
  style_value,
  ROUND(distance_meters::numeric, 2)::float8 AS distance_meters,
  total,
  attributes
FROM results
ORDER BY sort_order, distance_meters;
`, nil
}

func buildLayerQuery(sortOrder int, layer LayerDefinition) (string, error) {
	if err := validateLayer(layer); err != nil {
		return "", err
	}

	table := sanitizeIdentifier(layer.Table)
	layerAlias := sanitizeIdentifier(layer.Name + "_result")
	value := buildValueExpression(layer.ValueColumns)
	attributes := buildAttributesExpression(layer.Attributes)

	switch layer.SearchType {
	case SearchIntersect:
		return fmt.Sprintf(`  SELECT *
  FROM (
    SELECT
      %d AS sort_order,
      '%s'::text AS layer,
      %s AS style_value,
      0::double precision AS distance_meters,
      NULL::bigint AS total,
      %s AS attributes
    FROM %s AS t
    CROSS JOIN params AS p
    WHERE ST_Intersects(t.wkb_geometry, p.search_point)
    LIMIT %d
  ) AS %s`,
			sortOrder,
			layer.Name,
			value,
			attributes,
			table,
			layer.Limit,
			layerAlias,
		), nil

	case SearchRadius:
		return fmt.Sprintf(`  SELECT *
  FROM (
    SELECT
      %d AS sort_order,
      '%s'::text AS layer,
      %s AS style_value,
      ST_Distance(t.wkb_geometry, p.search_point) * 111320 AS distance_meters,
      NULL::bigint AS total,
      %s AS attributes
    FROM %s AS t
    CROSS JOIN params AS p
    WHERE ST_DWithin(t.wkb_geometry, p.search_point, p.radius_degrees)
    ORDER BY t.wkb_geometry <-> p.search_point
    LIMIT %d
  ) AS %s`,
			sortOrder,
			layer.Name,
			value,
			attributes,
			table,
			layer.Limit,
			layerAlias,
		), nil

	case SearchGrouped:
		groupColumn := columnReference(layer.ValueColumns[0])
		return fmt.Sprintf(`  SELECT
    %d AS sort_order,
    '%s'::text AS layer,
    %s AS style_value,
    MAX(ST_Distance(t.wkb_geometry, p.search_point) * 111320) AS distance_meters,
    COUNT(*)::bigint AS total,
    %s AS attributes
  FROM %s AS t
  CROSS JOIN params AS p
  WHERE ST_DWithin(t.wkb_geometry, p.search_point, p.radius_degrees)
  GROUP BY %s, p.search_point`,
			sortOrder,
			layer.Name,
			value,
			attributes,
			table,
			groupColumn,
		), nil
	}

	return "", fmt.Errorf("layer %q menggunakan tipe pencarian yang tidak didukung: %q", layer.Name, layer.SearchType)
}

func validateLayer(layer LayerDefinition) error {
	if !validIdentifier(layer.Name) {
		return fmt.Errorf("nama layer %q bukan identifier SQL yang aman", layer.Name)
	}
	if !validIdentifier(layer.Table) {
		return fmt.Errorf("nama tabel %q untuk layer %q bukan identifier SQL yang aman", layer.Table, layer.Name)
	}
	if len(layer.ValueColumns) == 0 {
		return fmt.Errorf("layer %q wajib memiliki minimal satu kolom nilai", layer.Name)
	}

	for _, column := range layer.ValueColumns {
		if !validIdentifier(column) {
			return fmt.Errorf("kolom nilai %q pada layer %q tidak valid", column, layer.Name)
		}
	}
	for _, attribute := range layer.Attributes {
		if !validIdentifier(attribute) {
			return fmt.Errorf("atribut %q pada layer %q tidak valid", attribute, layer.Name)
		}
	}

	switch layer.SearchType {
	case SearchIntersect, SearchRadius:
		if layer.Limit < 1 || layer.Limit > maxLayerResultLimit {
			return fmt.Errorf("batas hasil layer %q harus antara 1 dan %d", layer.Name, maxLayerResultLimit)
		}

	case SearchGrouped:
		if len(layer.ValueColumns) != 1 {
			return fmt.Errorf("layer grouped %q harus memiliki tepat satu kolom nilai", layer.Name)
		}
		for _, attribute := range layer.Attributes {
			if attribute != layer.ValueColumns[0] {
				return fmt.Errorf("atribut %q pada layer grouped %q harus sama dengan kolom pengelompokan", attribute, layer.Name)
			}
		}

	default:
		return fmt.Errorf("layer %q menggunakan tipe pencarian yang tidak didukung: %q", layer.Name, layer.SearchType)
	}

	return nil
}

func validIdentifier(value string) bool {
	return sqlIdentifierPattern.MatchString(value)
}

func sanitizeIdentifier(value string) string {
	return pgx.Identifier{value}.Sanitize()
}

func columnReference(column string) string {
	return "t." + sanitizeIdentifier(column)
}

func buildValueExpression(columns []string) string {
	parts := make([]string, 0, len(columns))
	for _, column := range columns {
		parts = append(parts, columnReference(column)+"::text")
	}

	if len(parts) == 1 {
		return parts[0]
	}
	return "(" + strings.Join(parts, " || '|' || ") + ")"
}

func buildAttributesExpression(attributes []string) string {
	if len(attributes) == 0 {
		return "'{}'::jsonb"
	}

	objects := make([]string, 0, (len(attributes)+jsonAttributesChunk-1)/jsonAttributesChunk)
	for start := 0; start < len(attributes); start += jsonAttributesChunk {
		end := min(start+jsonAttributesChunk, len(attributes))
		pairs := make([]string, 0, (end-start)*2)

		for _, attribute := range attributes[start:end] {
			pairs = append(pairs, "'"+attribute+"'", columnReference(attribute))
		}

		objects = append(objects, "jsonb_build_object("+strings.Join(pairs, ", ")+")")
	}

	return strings.Join(objects, " || ")
}
