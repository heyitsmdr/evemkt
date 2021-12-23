package eve

import (
	"context"
	"fmt"
	"math"
	"net/http"

	"github.com/antihax/goesi"
	"github.com/antihax/goesi/esi"
	"github.com/antihax/goesi/optional"
	"github.com/gregjones/httpcache"
	diskcache "github.com/gregjones/httpcache/diskcache"
	"github.com/leekchan/accounting"
	"github.com/peterbourgon/diskv/v3"
)

type EVE struct {
	EVE        *goesi.APIClient
	Cache      *diskv.Diskv
	Accounting accounting.Accounting
}

type SearchOptions struct {
	RegionId     int32
	MinProfit    int32
	ShipCapacity float64
	TaxRate      float32
	MaxTrips     int32
}

type MarketMatch struct {
	BuyOrder        esi.GetMarketsRegionIdOrders200Ok
	SellOrder       esi.GetMarketsRegionIdOrders200Ok
	BuyOrderPrice   string
	SellOrderPrice  string
	MoveQuantity    float64
	MoveVolumeTotal float64
	BuyISK          string
	SellISK         string
	Profit          string
	ProfitPerJump   string
	Jumps           int
}

var eve *EVE

func Init() {
	transport := httpcache.NewTransport(diskcache.New("cache-data"))
	transport.Transport = &http.Transport{Proxy: http.ProxyFromEnvironment}
	httpClient := &http.Client{Transport: transport}

	eve = &EVE{
		EVE: goesi.NewAPIClient(httpClient, "evemkt - An EVE EVE CLI; Created by <Atticus Windstorm>"),
		Cache: diskv.New(diskv.Options{
			BasePath:     "eve-marketer-data",
			Transform:    func(s string) []string { return []string{} },
			CacheSizeMax: 1024 * 1024 * 1024, // 1GB cache
		}),
		Accounting: accounting.Accounting{Symbol: "$", Precision: 2},
	}
}

func FetchAllRegionOrders(so *SearchOptions, updateLabelFunc func(string)) (orders []esi.GetMarketsRegionIdOrders200Ok) {
	page := 1
	for {
		updateLabelFunc(fmt.Sprintf("Fetching %d..", page))
		o, _, err := eve.EVE.ESI.MarketApi.GetMarketsRegionIdOrders(
			context.Background(),
			"",
			so.RegionId,
			&esi.GetMarketsRegionIdOrdersOpts{Page: optional.NewInt32(int32(page))},
		)
		if err != nil {
			break
		}
		orders = append(orders, o...)
		page = page + 1
	}

	return
}

func MatchCriteria(orders []esi.GetMarketsRegionIdOrders200Ok, so *SearchOptions, updateProgress func(float64, float64)) (mm []MarketMatch) {
	ordersMap := make(map[int32]map[string][]esi.GetMarketsRegionIdOrders200Ok)

	for _, o := range orders {
		if _, exists := ordersMap[o.TypeId]; !exists {
			ordersMap[o.TypeId] = map[string][]esi.GetMarketsRegionIdOrders200Ok{
				"buyOrders":  make([]esi.GetMarketsRegionIdOrders200Ok, 0),
				"sellOrders": make([]esi.GetMarketsRegionIdOrders200Ok, 0),
			}
		}

		if o.IsBuyOrder {
			ordersMap[o.TypeId]["buyOrders"] = append(ordersMap[o.TypeId]["buyOrders"], o)
		} else {
			ordersMap[o.TypeId]["sellOrders"] = append(ordersMap[o.TypeId]["sellOrders"], o)
		}
	}

	count := 1
	for _, orders := range ordersMap {
		updateProgress(float64(count), float64(len(ordersMap)))
		for _, sellOrder := range orders["sellOrders"] {
			for _, buyOrder := range orders["buyOrders"] {
				if isMatch := compareSellOrder2BuyOrder(sellOrder, buyOrder, so); isMatch != nil {
					mm = append(mm, *isMatch)
				}
			}
		}

		count = count + 1
	}

	return
}

func compareSellOrder2BuyOrder(sellOrder, buyOrder esi.GetMarketsRegionIdOrders200Ok, so *SearchOptions) *MarketMatch {
	// Normalize the quantity.
	normalizedQuantity := math.Min(float64(sellOrder.VolumeRemain), float64(buyOrder.VolumeRemain))
	typeVolume := ItemInfo(sellOrder.TypeId).Volume
	if (normalizedQuantity * float64(typeVolume)) > so.ShipCapacity {
		normalizedQuantity = math.Floor(so.ShipCapacity / float64(typeVolume))
	}

	buyISK := normalizedQuantity * sellOrder.Price
	sellISK := normalizedQuantity * buyOrder.Price
	profit := sellISK - buyISK

	if profit > float64(so.MinProfit) {
		jumps := len(RouteInfo(sellOrder.SystemId, buyOrder.SystemId))

		return &MarketMatch{
			BuyOrder:        buyOrder,
			SellOrder:       sellOrder,
			BuyOrderPrice:   eve.Accounting.FormatMoney(buyOrder.Price),
			SellOrderPrice:  eve.Accounting.FormatMoney(sellOrder.Price),
			MoveQuantity:    normalizedQuantity,
			MoveVolumeTotal: float64(ItemInfo(sellOrder.TypeId).Volume) * normalizedQuantity,
			BuyISK:          eve.Accounting.FormatMoney(buyISK),
			SellISK:         eve.Accounting.FormatMoney(sellISK),
			Profit:          eve.Accounting.FormatMoney(profit),
			Jumps:           jumps,
			ProfitPerJump:   eve.Accounting.FormatMoney(profit / float64(jumps)),
		}
	}

	return nil
}