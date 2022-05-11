package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"time"
)

type Configuration struct {
	HTTPPort    string `split_words:"true" default:"8080"`
	SortingUrls string `split_words:"true"`
}

type Sorting struct {
	ArraySize int64 `json:"arraySize"`
}

type SortingResponse struct {
	TimeSpentInMs             int64 `json:"timeSpentInMs"`
	TimeSpentCreatingListInMs int64 `json:"timeSpentCreatingListInMs"`
	TimeSpentSortingListInMs  int64 `json:"timeSpentSortingListInMs"`
}

func sorting(sorting Sorting) SortingResponse {
	start := time.Now().UnixMilli()
	uuids := make([]uuid.UUID, sorting.ArraySize)
	for i := int64(0); i < sorting.ArraySize; i++ {
		uuids[i] = uuid.New()
	}
	startSort := time.Now().UnixMilli()
	sort.Slice(uuids, func(i, j int) bool {
		return uuids[i].String() < uuids[j].String()
	})
	end := time.Now().UnixMilli()

	return SortingResponse{
		TimeSpentInMs:             startSort - start,
		TimeSpentCreatingListInMs: end - startSort,
		TimeSpentSortingListInMs:  end - start,
	}
}

func main() {
	var configuration Configuration
	envconfig.Process("", &configuration)

	sortingUrls := strings.Split(configuration.SortingUrls, ",")

	r := gin.Default()

	// Prometheus Metrics
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)

	v1 := r.Group("/v1")

	v1.POST("/sorting", func(c *gin.Context) {
		var req Sorting
		if err := c.BindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		res := sorting(req)
		c.JSON(http.StatusOK, res)
	})
	v1.POST("/delegated/sorting", func(c *gin.Context) {
		res, err := http.Post(sortingUrls[rand.Intn(len(sortingUrls))], "application/json", c.Request.Body)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer res.Body.Close()
		c.DataFromReader(http.StatusOK, res.ContentLength, "application/json", res.Body, nil)
	})

	r.Run(":" + configuration.HTTPPort)
}
