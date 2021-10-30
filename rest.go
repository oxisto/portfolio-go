package divplan

import "github.com/gin-gonic/gin"

type PortfolioWithSnapshot struct {
	*Portfolio

	SnapshotValue Money `json:"snapshotValue"`
}

type PortfolioSnapshot struct {
	PortfolioId string              `json:"portfolioId"`
	Positions   []*SecurityPosition `json:"positions"`
}

func GoRest() {
	r := gin.Default()
	r.GET("/api/v0/portfolios", ListPortfolios)
	r.GET("/api/v0/portfolios/:id/assets", ListPortfolioAsset)
	r.GET("/api/v0/securities", ListSecurities)
	r.GET("/api/v0/taxonomies", ListTaxonomies)
	r.GET("/v1/depot", func(c *gin.Context) {
		c.JSON(200, entries)
	})
	r.Run("0.0.0.0:5712") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func ListSecurities(c *gin.Context) {
	c.JSON(200, securities)
}

func ListPortfolios(c *gin.Context) {
	var portfolios []*PortfolioWithSnapshot

	for _, position := range depot.Portfolios {
		// add our snapshot value
		var withSnapshot = &PortfolioWithSnapshot{
			Portfolio:     position,
			SnapshotValue: CalculateSnapshotValue(position),
		}

		portfolios = append(portfolios, withSnapshot)
	}

	c.JSON(200, portfolios)
}

func ListPortfolioAsset(c *gin.Context) {
	var uuid = c.Param("id")

	var portfolio = depot.GetPortfolio(uuid)

	if portfolio == nil {
		c.JSON(404, map[string]interface{}{})
		return
	}

	var snapshot = PortfolioSnapshot{
		PortfolioId: portfolio.UUID,
		Positions:   Calc(portfolio),
	}

	c.JSON(200, snapshot)
}

func ListTaxonomies(c *gin.Context) {
	c.JSON(200, []interface{}{})
}
