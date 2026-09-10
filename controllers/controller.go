package controllers

import (
	"context"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"expressXgolang/golang/services"

	"github.com/gin-gonic/gin"
)

type coordinates struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
}

type analyzeResponse struct {
	Success     bool                      `json:"success"`
	Coordinates coordinates               `json:"coordinates"`
	Data        []services.AnalysisResult `json:"data"`
}

func GetAnalyzeLocation(c *gin.Context) {
	latParam := c.Query("lat")
	lngParam := c.Query("lng")

	if latParam == "" || lngParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "lat dan lng wajib diisi",
		})
		return
	}

	latitude, err := strconv.ParseFloat(latParam, 64)
	if err != nil || math.IsNaN(latitude) || math.IsInf(latitude, 0) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "lat harus berupa angka yang valid",
		})
		return
	}

	longitude, err := strconv.ParseFloat(lngParam, 64)
	if err != nil || math.IsNaN(longitude) || math.IsInf(longitude, 0) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "lng harus berupa angka yang valid",
		})
		return
	}

	if latitude < -90 || latitude > 90 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "lat harus berada di antara -90 dan 90",
		})
		return
	}

	if longitude < -180 || longitude > 180 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "lng harus berada di antara -180 dan 180",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	data, err := services.AnalyzeLocation(ctx, latitude, longitude)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"error": "analisis lokasi melewati batas waktu",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "gagal menganalisis lokasi",
		})
		return
	}

	c.JSON(http.StatusOK, analyzeResponse{
		Success: true,
		Coordinates: coordinates{
			Latitude:  latitude,
			Longitude: longitude,
		},
		Data: data,
	})
}

// Draft controller awal Anda disimpan di bawah ini sebagai referensi.
// import (
// 	"expressXgolang/golang/services"
// 	"github.com/gin-gonic/gin"
//     "net/http"
// 	"fmt"
// )

// func getAnalyzeLocation(c *gin.Context) {
// 	lat := c.Param("lat")
// 	lang := c.Param("lang")

// 	if (!lat || !lng){
// 		fmt.Println("Parameter belum diisi")
// 		c.Error("Parameter belum diisi")
// 	}

// }
