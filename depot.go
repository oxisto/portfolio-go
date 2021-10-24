package divplan

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/oxisto/divplan/divvydiary"
)

var entries []*divvydiary.DepotEntry
var apiKey string

func Sync(key string) *divvydiary.Depot {
	apiKey = key

	user, err := startSession()
	if err != nil {
		log.Fatalf("Could not retrieve session: %v", err)
		return nil
	}

	fmt.Printf("%+v", user)

	depot, err := retrieveDepot(user.ID)
	if err != nil {
		log.Fatalf("Could not retrieve depot: %v", err)
		return nil
	}
	//entries = depot.Entries

	return depot
}

func startSession() (user *divvydiary.User, err error) {
	req, err := http.NewRequest("GET", "https://api.divvydiary.com/session", nil)
	if err != nil {
		return nil, fmt.Errorf("could not create HTTP request: %w", err)
	}
	req.Header.Set("X-API-Key", string(apiKey))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while sending HTTP request: %w", err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading HTTP response: %w", err)
	}

	user = new(divvydiary.User)

	json.Unmarshal(body, user)

	return user, nil
}

type Depot struct {
	XMLName    xml.Name    `xml:"client"`
	Accounts   []*Account  `xml:"accounts>account"`
	Securities []*Security `xml:"securities>security"`
}

type Account struct {
	XMLName             xml.Name             `xml:"account"`
	Name                string               `xml:"name"`
	AccountTransactions []AccountTransaction `xml:"transactions>account-transaction"`
}

type AccountTransaction struct {
	XMLName    xml.Name    `xml:"account-transaction"`
	Amount     int32       `xml:"amount"`
	CrossEntry *CrossEntry `xml:"crossEntry"`
}

type CrossEntry struct {
	Portfolio Portfolio `xml:"portfolio"`
}

type Portfolio struct {
	PortfolioTransactions []PortfolioTransaction `xml:"transactions>portfolio-transaction"`
}

type PortfolioTransactions struct {
	PortfolioTransactions []PortfolioTransaction `xml:"portfolio-transaction"`
}

type PortfolioTransaction struct {
	Type     string    `xml:"type"`
	Security *Security `xml:"security"`
	Amount   Currency  `xml:"amount"`
	Shares   Quantity  `xml:"shares"`
}

type Currency float64

func (c *Currency) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var value string
	d.DecodeElement(&value, &start)
	i, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		return err
	}
	*c = (Currency)(float64(i) / 100.0)
	return nil
}

type Quantity float64

func (a *Quantity) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var value string
	d.DecodeElement(&value, &start)
	i, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		return err
	}
	*a = (Quantity)(float64(i) / 100000000.0)
	return nil
}

type Security struct {
	Name         string      `xml:"name"`
	ISIN         string      `xml:"isin"`
	TickerSymbol string      `xml:"tickerSymbol"`
	WKN          string      `xml:"wkn"`
	LatestPrice  LatestPrice `xml:"latest"`

	Reference string `xml:"reference,attr"`
}

type LatestPrice struct {
	Time  string   `xml:"t,attr"`
	Value Currency `xml:"v,attr"`

	High   Currency `xml:"high"`
	Low    Currency `xml:"low"`
	Volume Quantity `xml:"volume"`
}

func Load() {
	usr, _ := user.Current()
	dir := usr.HomeDir

	file, err := os.Open(filepath.Join(dir, "depot.xml"))
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)

	var depot Depot
	xml.Unmarshal(byteValue, &depot)

	security := depot.Securities[0].Name
	fmt.Printf("%+v\n", security)

	crossEntry := depot.Accounts[0].AccountTransactions[1].CrossEntry

	// loop through transactions and set security (hard-coded for now)
	for _, transaction := range crossEntry.Portfolio.PortfolioTransactions {
		if transaction.Security.Reference != "" {
			// look for the security
			rr := strings.Split(transaction.Security.Reference, "/")
			id, _ := strconv.ParseInt(strings.Trim(rr[len(rr)-1], "security[]"), 10, 64)

			var security = depot.Securities[id-1]
			*transaction.Security = *security
		}
	}

	fmt.Printf("\n=== Depot ===\n")

	var entryMap map[string]*divvydiary.DepotEntry = make(map[string]*divvydiary.DepotEntry)

	for _, transaction := range crossEntry.Portfolio.PortfolioTransactions {
		fmt.Printf("%s %.02f of %s for %.02f €\n",
			transaction.Type,
			transaction.Shares,
			transaction.Security.Name,
			transaction.Amount)

		var entry *divvydiary.DepotEntry
		entry, ok := entryMap[transaction.Security.ISIN]
		if !ok {
			entry = &divvydiary.DepotEntry{
				Name:   transaction.Security.Name,
				ISIN:   transaction.Security.ISIN,
				Symbol: transaction.Security.TickerSymbol,
				WKN:    transaction.Security.WKN,
				// TODO(cb): It seems that the Currency translation is not working for XML attributes, so we need to do it here
				Price: float32(transaction.Security.LatestPrice.Value) / 100000000.0,
			}
		}

		if transaction.Type == "BUY" {
			entry.Quantity += float32(transaction.Shares)
		} else {
			entry.Quantity -= float32(transaction.Shares)
		}

		if entry.Quantity <= 0 {
			// remove it from the map
			delete(entryMap, transaction.Security.ISIN)
		} else {
			entryMap[transaction.Security.ISIN] = entry
		}

	}

	entries = make([]*divvydiary.DepotEntry, 0)

	for _, entry := range entryMap {
		if entry != nil {
			entries = append(entries, entry)
		} else {
			fmt.Printf("Somehow this entry is empty")
		}
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	fmt.Printf("done!")
}

func retrieveDepot(userID int32) (depot *divvydiary.Depot, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.divvydiary.com/users/%d/depot", userID), nil)
	if err != nil {
		return nil, fmt.Errorf("could not create HTTP request: %w", err)
	}
	req.Header.Set("X-API-Key", string(apiKey))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while sending HTTP request: %w", err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading HTTP response: %w", err)
	}

	depot = new(divvydiary.Depot)
	depot.Entries = make([]*divvydiary.DepotEntry, 0)

	json.Unmarshal(body, &depot.Entries)

	return depot, nil
}
