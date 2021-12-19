package market

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/antihax/goesi"
	"github.com/antihax/goesi/esi"
	"github.com/antihax/goesi/optional"
	"github.com/leekchan/accounting"
	"github.com/peterbourgon/diskv/v3"
)

type Market struct {
	EVE        *goesi.APIClient
	Cache      *diskv.Diskv
	Accounting accounting.Accounting
}

type SearchOptions struct {
	RegionId     int32
	MinProfit    int32
	ShipCapacity float64
	TaxRate      float32
	MultiTrip    bool
	UseCache     bool
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

func New() *Market {
	m := &Market{
		EVE: goesi.NewAPIClient(&http.Client{}, "Personal EVE Marketer; Created by <Atticus Windstorm>"),
		Cache: diskv.New(diskv.Options{
			BasePath:     "eve-marketer-data",
			Transform:    func(s string) []string { return []string{} },
			CacheSizeMax: 1024 * 1024 * 1024, // 1GB cache
		}),
		Accounting: accounting.Accounting{Symbol: "$", Precision: 2},
	}

	return m
}

func (m *Market) FetchAllRegionOrders(so *SearchOptions, updateLabelFunc func(string)) []esi.GetMarketsRegionIdOrders200Ok {
	orders := make([]esi.GetMarketsRegionIdOrders200Ok, 0)

	if so.UseCache {
		marketCache, err := m.Cache.Read(fmt.Sprintf("market-%d", so.RegionId))
		if err == nil {
			_ = json.Unmarshal(marketCache, &orders)
		}
	}

	if len(orders) == 0 {
		page := 1
		for {
			updateLabelFunc(fmt.Sprintf("Fetching %d..", page))
			o, _, err := m.EVE.ESI.MarketApi.GetMarketsRegionIdOrders(
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
		j, _ := json.Marshal(orders)
		m.Cache.Write(fmt.Sprintf("market-%d", so.RegionId), j)
	}

	return orders
}

func (m *Market) MatchCriteria(orders []esi.GetMarketsRegionIdOrders200Ok, so *SearchOptions) []MarketMatch {
	mm := make([]MarketMatch, 0)

	for _, sellOrder := range orders {
		// If it's a sell order, continue.
		if !sellOrder.IsBuyOrder {
			for _, buyOrder := range orders {
				// If it's a buy order for the same item type, continue.
				if buyOrder.TypeId == sellOrder.TypeId && buyOrder.IsBuyOrder {
					if isMatch := m.compareSellOrder2BuyOrder(sellOrder, buyOrder, so); isMatch != nil {
						mm = append(mm, *isMatch)
					}
				}
			}
		}
	}

	return mm
}

func (m *Market) compareSellOrder2BuyOrder(sellOrder, buyOrder esi.GetMarketsRegionIdOrders200Ok, so *SearchOptions) *MarketMatch {
	// Normalize the quantity.
	normalizedQuantity := math.Min(float64(sellOrder.VolumeRemain), float64(buyOrder.VolumeRemain))
	if normalizedQuantity > so.ShipCapacity {
		normalizedQuantity = so.ShipCapacity
	}

	buyISK := normalizedQuantity * sellOrder.Price
	sellISK := normalizedQuantity * buyOrder.Price
	profit := sellISK - buyISK

	if profit > float64(so.MinProfit) {
		jumps := len(m.RouteInfo(sellOrder.SystemId, buyOrder.SystemId))

		return &MarketMatch{
			BuyOrder:        buyOrder,
			SellOrder:       sellOrder,
			BuyOrderPrice:   m.Accounting.FormatMoney(buyOrder.Price),
			SellOrderPrice:  m.Accounting.FormatMoney(sellOrder.Price),
			MoveQuantity:    normalizedQuantity,
			MoveVolumeTotal: float64(m.ItemInfo(sellOrder.TypeId).Volume) * normalizedQuantity,
			BuyISK:          m.Accounting.FormatMoney(buyISK),
			SellISK:         m.Accounting.FormatMoney(sellISK),
			Profit:          m.Accounting.FormatMoney(profit),
			Jumps:           jumps,
			ProfitPerJump:   m.Accounting.FormatMoney(profit / float64(jumps)),
		}
	}

	return nil
}

func (m *Market) ItemInfo(itemId int32) esi.GetUniverseTypesTypeIdOk {
	itemCache, err := m.Cache.Read(fmt.Sprintf("item-%d", itemId))
	if err == nil {
		var item esi.GetUniverseTypesTypeIdOk
		_ = json.Unmarshal(itemCache, &item)
		return item
	}

	item, _, err := m.EVE.ESI.UniverseApi.GetUniverseTypesTypeId(context.Background(), itemId, nil)
	if err != nil {
		panic(err)
	}

	j, _ := item.MarshalJSON()
	_ = m.Cache.Write(fmt.Sprintf("item-%d", itemId), j)

	return item
}

func (m *Market) SystemInfo(systemId int32) esi.GetUniverseSystemsSystemIdOk {
	systemCache, err := m.Cache.Read(fmt.Sprintf("system-%d", systemId))
	if err == nil {
		var sys esi.GetUniverseSystemsSystemIdOk
		_ = json.Unmarshal(systemCache, &sys)
		return sys
	}

	sys, _, err := m.EVE.ESI.UniverseApi.GetUniverseSystemsSystemId(context.Background(), systemId, nil)
	if err != nil {
		panic(err)
	}

	j, _ := sys.MarshalJSON()
	_ = m.Cache.Write(fmt.Sprintf("system-%d", systemId), j)

	return sys
}

func (m *Market) StationInfo(stationId int64) esi.GetUniverseStationsStationIdOk {
	stationCache, err := m.Cache.Read(fmt.Sprintf("station-%d", stationId))
	if err == nil {
		var station esi.GetUniverseStationsStationIdOk
		_ = json.Unmarshal(stationCache, &station)
		return station
	}

	station, _, err := m.EVE.ESI.UniverseApi.GetUniverseStationsStationId(context.Background(), int32(stationId), nil)
	if err != nil {
		panic(err)
	}

	j, _ := station.MarshalJSON()
	_ = m.Cache.Write(fmt.Sprintf("station-%d", stationId), j)

	return station
}

func (m *Market) RouteInfo(sourceId, destinationId int32) []int32 {
	routesCache, err := m.Cache.Read(fmt.Sprintf("route-%d-%d", sourceId, destinationId))
	if err == nil {
		var routes []int32
		_ = json.Unmarshal(routesCache, &routes)
		return routes
	}

	routes, _, err := m.EVE.ESI.RoutesApi.GetRouteOriginDestination(context.Background(), destinationId, sourceId, nil)
	if err != nil {
		panic(err)
	}

	j, _ := json.Marshal(routes)
	_ = m.Cache.Write(fmt.Sprintf("route-%d-%d", sourceId, destinationId), j)

	return routes
}
