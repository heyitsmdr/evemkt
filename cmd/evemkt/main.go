package main

import (
	"eve-marketer/internal/market"
	"fmt"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var ui *tview.Application
var marketOps *market.Market

var searchForm *tview.Form
var marketTable *tview.Table

func main() {
	marketOps = market.New()

	ui = tview.NewApplication()
	flex := tview.NewFlex()

	searchForm = renderSearchForm()
	marketTable = renderMarketTable()

	flex.AddItem(searchForm, 35, 0, true)
	flex.AddItem(marketTable, 0, 1, false)

	// Add headers to the market table, despite it being empty.
	addMarketTableHeaders()

	// Run the tview application.
	if err := ui.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func renderSearchForm() *tview.Form {
	so := &market.SearchOptions{
		RegionId:     10000033,
		MinProfit:    7000000,
		ShipCapacity: 1000,
		TaxRate:      8,
		MultiTrip:    false,
		UseCache:     true,
	}

	searchForm := tview.NewForm()
	searchForm.
		//AddDropDown("Title", []string{"Mr.", "Ms.", "Mrs.", "Dr.", "Prof."}, 0, nil).
		AddInputField("Region",
			strconv.Itoa(int(so.RegionId)),
			0,
			nil,
			func(text string) {
				i, _ := strconv.ParseInt(text, 10, 32)
				so.RegionId = int32(i)
			}).
		AddInputField("Min. Profit",
			strconv.Itoa(int(so.MinProfit)),
			0,
			nil,
			func(text string) {
				i, _ := strconv.ParseInt(text, 10, 32)
				so.MinProfit = int32(i)
			}).
		AddInputField("Ship Capacity (m³)",
			strconv.Itoa(int(so.ShipCapacity)),
			0,
			nil,
			func(text string) {
				f, _ := strconv.ParseFloat(text, 64)
				so.ShipCapacity = f
			}).
		AddCheckbox("Multi-Trip", so.MultiTrip, func(checked bool) { so.MultiTrip = checked }).
		AddInputField("Tax Rate %",
			strconv.Itoa(int(so.TaxRate)),
			0,
			nil,
			func(text string) {
				f, _ := strconv.ParseFloat(text, 32)
				so.TaxRate = float32(f)
			}).
		AddButton("Search", func() {
			searchForm.GetButton(0).SetLabel("Searching..")
			ui.ForceDraw()

			orders := marketOps.FetchAllRegionOrders(
				so,
				func(text string) {
					searchForm.GetButton(0).SetLabel(text)
					ui.ForceDraw()
				})

			searchForm.GetButton(0).SetLabel("Matching..")
			ui.ForceDraw()

			matches := marketOps.MatchCriteria(orders, so)
			updateMarketTable(matches)

			searchForm.GetButton(0).SetLabel("Search")
		}).
		AddCheckbox("Use Cache", so.UseCache, func(checked bool) { so.UseCache = checked }).
		AddButton("Quit", func() {
			ui.Stop()
		}).
		SetBorder(true).SetTitle("Market Search").SetTitleAlign(tview.AlignCenter)

	return searchForm
}

func renderMarketTable() *tview.Table {
	mktTable := tview.NewTable().SetBorders(true)

	return mktTable
}

func addMarketTableHeaders() {
	marketTable.
		SetCell(0, 0, tview.NewTableCell("Item").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft)).
		SetCell(0, 1, tview.NewTableCell("Volume").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft)).
		SetCell(0, 2, tview.NewTableCell("Buy").SetTextColor(tcell.ColorBlueViolet).SetAlign(tview.AlignLeft)).
		SetCell(0, 3, tview.NewTableCell("Amnt").SetTextColor(tcell.ColorBlueViolet).SetAlign(tview.AlignLeft)).
		SetCell(0, 4, tview.NewTableCell("Location").SetTextColor(tcell.ColorBlueViolet).SetAlign(tview.AlignLeft)).
		SetCell(0, 5, tview.NewTableCell("Sell").SetTextColor(tcell.ColorSpringGreen).SetAlign(tview.AlignLeft)).
		SetCell(0, 6, tview.NewTableCell("Amnt").SetTextColor(tcell.ColorSpringGreen).SetAlign(tview.AlignLeft)).
		SetCell(0, 7, tview.NewTableCell("Location").SetTextColor(tcell.ColorSpringGreen).SetAlign(tview.AlignLeft)).
		SetCell(0, 8, tview.NewTableCell("Move Quantity").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft)).
		SetCell(0, 9, tview.NewTableCell("Move Volume").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft)).
		SetCell(0, 10, tview.NewTableCell("Jumps").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft)).
		SetCell(0, 11, tview.NewTableCell("Upfront Cost").SetTextColor(tcell.ColorRed).SetAlign(tview.AlignLeft)).
		SetCell(0, 12, tview.NewTableCell("Profit").SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignLeft)).
		SetCell(0, 13, tview.NewTableCell("Profit/jump").SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignLeft)).
		SetSelectable(true, false).
		Select(0, 0).SetFixed(1, 1).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				ui.Stop()
			}
		}).
		SetSelectedFunc(func(row, column int) {
			marketTable.GetCell(row, column).SetTextColor(tcell.ColorRed)
		})
}
func updateMarketTable(matches []market.MarketMatch) {
	marketTable.Clear()
	addMarketTableHeaders()

	for i, match := range matches {
		row := i + 1

		marketTable.SetCell(row, 0, tview.NewTableCell(marketOps.ItemInfo(match.SellOrder.TypeId).Name))
		marketTable.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%.2f m³", marketOps.ItemInfo(match.SellOrder.TypeId).Volume)))

		marketTable.SetCell(row, 2, tview.NewTableCell(match.SellOrderPrice).SetTextColor(tcell.ColorBlueViolet))
		marketTable.SetCell(row, 3, tview.NewTableCell(fmt.Sprintf("%d", match.SellOrder.VolumeRemain)).SetTextColor(tcell.ColorBlueViolet))
		marketTable.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%s", marketOps.StationInfo(match.SellOrder.LocationId).Name)).SetTextColor(tcell.ColorBlueViolet))

		marketTable.SetCell(row, 5, tview.NewTableCell(match.BuyOrderPrice).SetTextColor(tcell.ColorSpringGreen))
		marketTable.SetCell(row, 6, tview.NewTableCell(fmt.Sprintf("%d", match.BuyOrder.VolumeRemain)).SetTextColor(tcell.ColorSpringGreen))
		marketTable.SetCell(row, 7, tview.NewTableCell(fmt.Sprintf("%s", marketOps.StationInfo(match.BuyOrder.LocationId).Name)).SetTextColor(tcell.ColorSpringGreen))

		marketTable.SetCell(row, 8, tview.NewTableCell(fmt.Sprintf("%.2f", match.MoveQuantity)))
		marketTable.SetCell(row, 9, tview.NewTableCell(fmt.Sprintf("%.0f m³", match.MoveVolumeTotal)))
		marketTable.SetCell(row, 10, tview.NewTableCell(fmt.Sprintf("%d", match.Jumps)))
		marketTable.SetCell(row, 11, tview.NewTableCell(match.BuyISK).SetTextColor(tcell.ColorRed))
		marketTable.SetCell(row, 12, tview.NewTableCell(match.Profit).SetTextColor(tcell.ColorGreen))
		marketTable.SetCell(row, 13, tview.NewTableCell(match.ProfitPerJump).SetTextColor(tcell.ColorGreen))
	}
}
