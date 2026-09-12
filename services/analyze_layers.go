package services

const analyzeRadiusDegrees = 0.027

type SearchType string

const (
	SearchIntersect SearchType = "intersect"
	SearchRadius    SearchType = "radius"
	SearchGrouped   SearchType = "grouped"
)

type LayerDefinition struct {
	Name         string
	Table        string
	ValueColumns []string
	SearchType   SearchType
	Attributes   []string
	Limit        int
}

// analyzeLayers adalah whitelist tabel dan kolom yang boleh masuk ke query.
// Untuk menambah layer, cukup tambahkan definisi di sini; jangan mengambil nama
// tabel atau kolom langsung dari parameter pengguna.
var analyzeLayers = []LayerDefinition{
	{
		Name:         "flood",
		Table:        "flood_zones",
		ValueColumns: []string{"area"},
		SearchType:   SearchIntersect,
		Attributes:   []string{"area"},
		Limit:        1,
	},
	{
		Name:         "temperature",
		Table:        "temperature_zones",
		ValueColumns: []string{"suhu"},
		SearchType:   SearchIntersect,
		Attributes:   []string{"suhu"},
		Limit:        1,
	},
	{
		Name:         "air_quality",
		Table:        "air_quality",
		ValueColumns: []string{"polusi"},
		SearchType:   SearchIntersect,
		Attributes:   []string{"shape_leng", "shape_area", "polusi"},
		Limit:        1,
	},
	{
		Name:         "green_spaces",
		Table:        "green_spaces",
		ValueColumns: []string{"rth"},
		SearchType:   SearchIntersect,
		Attributes:   []string{"rth"},
		Limit:        1,
	},
	{
		Name:         "population",
		Table:        "population_data",
		ValueColumns: []string{"jumlah_pen"},
		SearchType:   SearchIntersect,
		Attributes: []string{
			"nama_kab", "nama_kec", "jumlah_pen", "jumlah_kk",
			"islam", "kristen", "katholik", "hindu", "budha",
			"konghucu", "kepercayaa", "pria", "wanita", "kawin",
			"belum_kawi", "cerai_hidu", "cerai_mati",
			"u0", "u5", "u10", "u15", "u20", "u25", "u30",
			"u35", "u40", "u45", "u50", "u55", "u60", "u65",
			"u70", "u75", "tidak_blm_", "belum_tama", "tamat_sd",
			"sltp", "slta", "d1_dan_d2", "d3", "s1", "s2", "s3",
			"belum_tida", "pensiunan", "mengurus_r", "perdaganga",
			"perawat", "nelayan", "pelajar_ma", "guru", "wiraswasta",
			"pengacara", "lainnya",
		},
		Limit: 1,
	},
	{
		Name:         "elevation",
		Table:        "elevation_zones",
		ValueColumns: []string{"ketinggian"},
		SearchType:   SearchIntersect,
		Attributes:   []string{"ketinggian"},
		Limit:        1,
	},
	{
		Name:         "roads_buffer",
		Table:        "roads_buffer",
		ValueColumns: []string{"remark"},
		SearchType:   SearchRadius,
		Attributes:   []string{"remark", "shape_leng", "shape_area"},
		Limit:        5,
	},
	{
		Name:         "public_facilities",
		Table:        "public_facilities",
		ValueColumns: []string{"kategori"},
		SearchType:   SearchGrouped,
		Attributes:   []string{"kategori"},
	},
	{
		Name:         "wifi",
		Table:        "wifi_zones",
		ValueColumns: []string{"dl_mbps", "ul_mbps"},
		SearchType:   SearchIntersect,
		Attributes:   []string{"avg_lat_ms", "tests", "dl_mbps", "ul_mbps"},
		Limit:        1,
	},
	{
		Name:         "mobile_data",
		Table:        "mobile_data_zones",
		ValueColumns: []string{"dl_mbps", "ul_mbps"},
		SearchType:   SearchIntersect,
		Attributes:   []string{"avg_lat_ms", "tests", "dl_mbps", "ul_mbps"},
		Limit:        1,
	},
	{
		Name:         "crime",
		Table:        "crime_zones",
		ValueColumns: []string{"crime_total"},
		SearchType:   SearchIntersect,
		Attributes: []string{
			"nama_kab", "jumlah_pen", "jumlah_kk", "tahun",
			"crime_total", "crime_cleared", "crime_rate",
			"clearance_rate", "sumber",
		},
		Limit: 1,
	},
}
