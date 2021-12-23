package eve

import (
	"context"
	"github.com/antihax/goesi/esi"
	"strings"
)

func ItemInfo(itemId int32) esi.GetUniverseTypesTypeIdOk {
	item, _, err := eve.EVE.ESI.UniverseApi.GetUniverseTypesTypeId(context.Background(), itemId, nil)
	if err != nil {
		panic(err)
	}

	return item
}

func SystemInfo(systemId int32) esi.GetUniverseSystemsSystemIdOk {
	sys, _, err := eve.EVE.ESI.UniverseApi.GetUniverseSystemsSystemId(context.Background(), systemId, nil)
	if err != nil {
		panic(err)
	}

	return sys
}

func StationInfo(stationId int64) esi.GetUniverseStationsStationIdOk {
	station, _, err := eve.EVE.ESI.UniverseApi.GetUniverseStationsStationId(context.Background(), int32(stationId), nil)
	if err != nil {
		panic(err)
	}

	return station
}

func RouteInfo(sourceId, destinationId int32) []int32 {
	routes, _, err := eve.EVE.ESI.RoutesApi.GetRouteOriginDestination(context.Background(), destinationId, sourceId, nil)
	if err != nil {
		panic(err)
	}

	return routes
}

func RegionInfo(regionId int32) (regionInfo esi.GetUniverseRegionsRegionIdOk) {
	regionInfo, _, err := eve.EVE.ESI.UniverseApi.GetUniverseRegionsRegionId(context.Background(), regionId, nil)
	if err != nil {
		panic(err)
	}

	return
}

func GetRegions() (regions []string) {
	regionIds, _, err := eve.EVE.ESI.UniverseApi.GetUniverseRegions(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	for _, regionId := range regionIds {
		regionInfo := RegionInfo(regionId)
		regions = append(regions, regionInfo.Name)
	}

	return
}

func GetRegionInfoByName(name string) *esi.GetUniverseRegionsRegionIdOk {
	regionIds, _, err := eve.EVE.ESI.UniverseApi.GetUniverseRegions(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	for _, regionId := range regionIds {
		regionInfo := RegionInfo(regionId)
		if strings.ToLower(regionInfo.Name) == strings.ToLower(name) {
			return &regionInfo
		}
	}

	return nil
}

func GetItems() (items []string) {
	typeIds, _, err := eve.EVE.ESI.UniverseApi.GetUniverseTypes(context.Background(), nil)
	if err != nil {
		panic(err)
	}

	for _, typeId := range typeIds {
		itemInfo := ItemInfo(typeId)
		items = append(items, itemInfo.Name)
	}

	return
}