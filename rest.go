package divplan

import "github.com/gin-gonic/gin"

func GoRest() {
	r := gin.Default()
	r.GET("/v1/depot", func(c *gin.Context) {
		c.JSON(200, entries)
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
