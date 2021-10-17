package divvydiary

type User struct {
	ID       int32  `json:"id"`
	EMail    string `json:"email"`
	Forename string `json:"forename"`
}

type Depot struct {
	Entries []DepotEntry
}

type DepotEntry struct {
	Name     string  `json:"name"`
	ISIN     string  `json:"isin"`
	WKN      string  `json:"wkn"`
	Quantity int32   `json:"quantity"`
	Price    float32 `json:"price"`
	Symbol   string  `json:"symbol"`
}
