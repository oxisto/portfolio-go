package divplan

import "fmt"

func CalculateSnapshotValue(portfolio *Portfolio) Money {
	return Money{
		Amount:       0,
		CurrencyCode: "EUR",
	}
}

type SecurityPosition struct {
	Investment Security    `json:"investment"`
	Price      LatestPrice `json:"price"`
	Shares     int         `json:"shares"`

	// only for REST
	Record SecurityPerformanceRecord `json:"record"`
}

type SecurityPerformanceRecord struct {
	FifoCostPerSharesHeld  Quote `json:"fifoCostPerSharesHeld"`
	Quote                  Quote `json:"quote"`
	CapitalGainsOnHoldings Money `json:"capitalGainsOnHoldings"`
	MarketValue            Money `json:"marketValue"`
}

func Calc(portfolio *Portfolio) []*SecurityPosition {
	var entryMap map[string]*SecurityPosition = make(map[string]*SecurityPosition)

	for _, transaction := range portfolio.PortfolioTransactions {
		fmt.Printf("%s %.02f of %s for %.02f €\n",
			transaction.Type,
			float32(transaction.Shares/10000000.0),
			transaction.Security.Name,
			transaction.Amount)

		var entry *SecurityPosition
		entry, ok := entryMap[transaction.Security.ISIN]
		if !ok {
			entry = &SecurityPosition{
				Investment: *transaction.Security,
				Price:      transaction.Security.LatestPrice,
			}

			if transaction.Security.CurrencyCode == "USD" {
				// TODO(db): use currency values
				//var conversionRate float32 = 0.86

				entry.Price = LatestPrice{}

			}
		}

		if transaction.Type == "BUY" {
			entry.Shares += transaction.Shares
			// TODO: calc correctly
			//entry.BuyPrice = transaction.Amount / transaction.Shares
		} else {
			entry.Shares -= transaction.Shares
		}

		if entry.Shares <= 0 {
			// remove it from the map
			delete(entryMap, transaction.Security.ISIN)
		} else {
			/*if entry.Price != 0 {
				// TODO: why is the price 0?
				// update profit values
				entry.Profit = entry.Quantity * (entry.Price - entry.BuyPrice)
				entry.ProfitPercentage = entry.Profit / (entry.Quantity * entry.BuyPrice)
			}*/

			entryMap[transaction.Security.ISIN] = entry
		}
	}

	var positions []*SecurityPosition

	for _, position := range entryMap {
		position.Record = SecurityPerformanceRecord{
			FifoCostPerSharesHeld:  Quote{CurrencyCode: "EUR"},
			Quote:                  Quote{CurrencyCode: "EUR"},
			CapitalGainsOnHoldings: Money{CurrencyCode: "EUR"},
			MarketValue:            Money{CurrencyCode: "EUR"},
		}

		positions = append(positions, position)
	}

	return positions
}
