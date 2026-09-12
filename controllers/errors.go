package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NotFound memastikan endpoint yang tidak terdaftar tetap mengembalikan JSON.
func NotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"error": "endpoint tidak ditemukan",
	})
}

// MethodNotAllowed menjelaskan bahwa URL tersedia tetapi metode HTTP salah.
func MethodNotAllowed(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{
		"error": "metode HTTP tidak diizinkan untuk endpoint ini",
	})
}
