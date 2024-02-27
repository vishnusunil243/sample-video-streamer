package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Static("/templates", "./templates")
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to our video streaming platform")
	})
	router.GET("/stream", func(c *gin.Context) {
		filename := "s2.mp4"
		if !strings.HasSuffix(filename, ".mp4") {
			c.String(http.StatusBadRequest, "Unsupported video format")
			return
		}
		file, err := os.Open("videos/" + filename)
		if err != nil {
			c.String(http.StatusNotFound, "Video not found")
			return
		}
		defer file.Close()

		stat, err := file.Stat()
		if err != nil {
			c.String(http.StatusInternalServerError, "Internal server error")
			return
		}

		fileSize := stat.Size()
		rangeHeader := c.GetHeader("Range")
		if rangeHeader != "" {
			ranges, err := parseRange(rangeHeader, fileSize)
			if err != nil {
				c.String(http.StatusBadRequest, "Invalid Range Header")
				return
			}
			if len(ranges) > 0 {
				c.Status(http.StatusPartialContent)
				c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", ranges[0].start, ranges[0].end, fileSize))
				c.Header("Accept-Ranges", "bytes")
				c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
				http.ServeContent(c.Writer, c.Request, "videos/"+filename, stat.ModTime(), io.NewSectionReader(file, ranges[0].start, ranges[0].end-ranges[0].start+1))
				return
			}
		}

		c.Header("Content-Type", "video/mp4")
		c.Header("Content-Length", fmt.Sprintf("%d", fileSize))
		c.File("videos/" + filename)
	})
	router.LoadHTMLFiles("./templates/temp.html")
	router.GET("/html", func(c *gin.Context) {
		c.HTML(http.StatusOK, "temp.html", gin.H{})
	})
	router.Run(":8080")
}

type rangeInfo struct {
	start int64
	end   int64
}

func parseRange(rangeHeader string, fileSize int64) ([]rangeInfo, error) {
	var ranges []rangeInfo
	parts := strings.Split(rangeHeader[6:], "-")
	fmt.Println(parts)
	if len(parts) != 2 {
		return nil, errors.New("Invalid Range Header")
	}
	start, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, errors.New("Invalid Range Header")
	}
	var end int64
	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, errors.New("Invalid Range Header")
		}
	} else {
		end = fileSize - 1
	}
	if start < 0 {
		start = 0
	}
	if end >= fileSize || end == 0 {
		end = fileSize - 1
	}
	ranges = append(ranges, rangeInfo{start: start, end: end})
	return ranges, nil
}
