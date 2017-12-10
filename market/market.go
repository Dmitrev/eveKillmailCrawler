package market

import (
	"sync"
	"net/http"
	"fmt"
	"encoding/json"
	"time"
)

type esiMarketPrices struct {
	TypeID        int     `json:"type_id"`
	AveragePrice  float64 `json:"average_price,omitempty"`
	AdjustedPrice float64 `json:"adjusted_price"`
}

type marketPrice struct {
	TypeId int
	Price  float64
}

var marketData map[int]*marketPrice

var mu *sync.Mutex

const marketURL = "https://esi.tech.ccp.is/latest/markets/prices/?datasource=tranquility"
const timeout = 3600

var ticker *time.Ticker

func Init() {
	mu = &sync.Mutex{}
	ticker = time.NewTicker(time.Second * time.Duration(timeout))
	marketData = make(map[int]*marketPrice)

	go func() {
		requestMarketData()
		for range ticker.C {
			requestMarketData()
		}
	}()
}

func requestMarketData() {
	fmt.Println("Requesting market data")

	resp, err := http.Get(marketURL)
	if err != nil {
		fmt.Println("ERROR HTTP GET")
		return
	}
	defer resp.Body.Close()

	var marketDataJSON *[]esiMarketPrices
	json.NewDecoder(resp.Body).Decode(&marketDataJSON)
	updateMarketData(*marketDataJSON)
}

func updateMarketData(data []esiMarketPrices) {
	mu.Lock()
	defer mu.Unlock()

	for _, value := range data {
		if marketData[value.TypeID] == nil {
			marketData[value.TypeID] = &marketPrice{
				TypeId:        value.TypeID,
				Price: value.AdjustedPrice,
			}
		} else {
			marketData[value.TypeID].Price = value.AdjustedPrice
		}
	}
	fmt.Println("New market data set")
}

func GetPriceOfTypeID(id int) float64 {
	mu.Lock()
	defer mu.Unlock()

	if marketData[id] != nil {
		return marketData[id].Price
	}

	return 0
}
