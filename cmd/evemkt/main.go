package main

import (
	"eve-marketer/internal/eve"
	search "eve-marketer/internal/pages/search"
	"fmt"
	"math"
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var ui *tview.Application

var marketTable *tview.Table

func main() {
	eve.Init()

	ui = tview.NewApplication()

	pages := tview.
		NewPages().
		AddPage("hauling", renderHaulingPage(), true, true).
		AddPage("search", search.Render(ui), true, false)

	ui.
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyF1:
				pages.SwitchToPage("hauling")
			case tcell.KeyF2:
				pages.SwitchToPage("search")
			case tcell.KeyEscape:
				ui.Stop()
			case tcell.KeyEnter:
				currentPage, _ :=  pages.GetFrontPage()
				if currentPage == "search" {
					search.ShowSearch()
				}
			}

			return event
		})

	appFlex := tview.NewFlex()
	appFlex.
		SetDirection(tview.FlexRow).
		AddItem(pages, 0, 1, true).
		AddItem(tview.NewTextView().SetDynamicColors(true).SetText(
			"[yellow]F1[-] Hauling\t[yellow]F2[-] Search\t[red]ESC[-] Quit"),
			1, 0, false,
		)

	// Add headers to the market table, despite it being empty.
	addMarketTableHeaders()

	// Run the tview application.
	if err := ui.SetRoot(appFlex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func renderHaulingPage() *tview.Flex {
	marketTable = renderMarketTable()

	return tview.
		NewFlex().
		AddItem(renderHaulingSearchForm(), 35, 0, true).
		AddItem(marketTable, 0, 1, false)
}

func renderHaulingSearchForm() (searchForm *tview.Form) {
	so := &eve.SearchOptions{
		RegionId:     10000033,
		MinProfit:    1000000,
		ShipCapacity: 2500,
		TaxRate:      8,
		MaxTrips:     1,
	}

	searchForm = tview.NewForm()
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
		AddInputField("Ship Capacity (m??)",
			strconv.Itoa(int(so.ShipCapacity)),
			0,
			nil,
			func(text string) {
				f, _ := strconv.ParseFloat(text, 64)
				so.ShipCapacity = f
			}).
		AddInputField("Max Trips",
			strconv.Itoa(int(so.MaxTrips)),
			3,
			nil,
			func(text string) {
				i, _ := strconv.ParseInt(text, 10, 32)
				so.MaxTrips = int32(i)
			}).
		AddInputField("Tax Rate %",
			strconv.Itoa(int(so.TaxRate)),
			3,
			nil,
			func(text string) {
				f, _ := strconv.ParseFloat(text, 32)
				so.TaxRate = float32(f)
			}).
		AddButton("Search", func() {
			searchForm.GetButton(0).SetLabel("Searching..")
			ui.ForceDraw()

			orders := eve.FetchAllRegionOrders(
				so,
				func(text string) {
					searchForm.GetButton(0).SetLabel(text)
					ui.ForceDraw()
				})

			searchForm.GetButton(0).SetLabel("Matching..")
			ui.ForceDraw()

			matches := eve.MatchCriteria(orders, so, func(current, total float64) {
				percent := math.Floor((current * 100) / total)
				searchForm.GetButton(0).SetLabel(fmt.Sprintf("Matching (%.0f%%)..", percent))
				ui.ForceDraw()
			})

			updateMarketTable(matches)

			searchForm.GetButton(0).SetLabel("Search")
		}).
		SetBorder(true).
		SetTitle("EVE Search").
		SetTitleAlign(tview.AlignCenter)

	return
}

func renderMarketTable() (tbl *tview.Table) {
	tbl = tview.NewTable()
	tbl.
		SetBorders(true).
		SetTitle("Matching Orders").
		SetBorder(true).
		SetBorderPadding(1, 1, 1, 1)

	return
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
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEscape {
				ui.Stop()
			}
		}).
		SetSelectedFunc(func(row, column int) {
			marketTable.GetCell(row, column).SetTextColor(tcell.ColorRed)
		})
}
func updateMarketTable(matches []eve.MarketMatch) {
	marketTable.Clear()
	addMarketTableHeaders()

	for i, match := range matches {
		row := i + 1

		marketTable.SetCell(row, 0, tview.NewTableCell(eve.ItemInfo(match.SellOrder.TypeId).Name))
		marketTable.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%.2f m??", eve.ItemInfo(match.SellOrder.TypeId).Volume)))

		marketTable.SetCell(row, 2, tview.NewTableCell(match.SellOrderPrice).SetTextColor(tcell.ColorBlueViolet))
		marketTable.SetCell(row, 3, tview.NewTableCell(fmt.Sprintf("%d", match.SellOrder.VolumeRemain)).SetTextColor(tcell.ColorBlueViolet))
		marketTable.SetCell(row, 4, tview.NewTableCell(fmt.Sprintf("%s", eve.StationInfo(match.SellOrder.LocationId).Name)).SetTextColor(tcell.ColorBlueViolet))

		marketTable.SetCell(row, 5, tview.NewTableCell(match.BuyOrderPrice).SetTextColor(tcell.ColorSpringGreen))
		marketTable.SetCell(row, 6, tview.NewTableCell(fmt.Sprintf("%d", match.BuyOrder.VolumeRemain)).SetTextColor(tcell.ColorSpringGreen))
		marketTable.SetCell(row, 7, tview.NewTableCell(fmt.Sprintf("%s", eve.StationInfo(match.BuyOrder.LocationId).Name)).SetTextColor(tcell.ColorSpringGreen))

		marketTable.SetCell(row, 8, tview.NewTableCell(fmt.Sprintf("%.2f", match.MoveQuantity)))
		marketTable.SetCell(row, 9, tview.NewTableCell(fmt.Sprintf("%.0f m??", match.MoveVolumeTotal)))
		marketTable.SetCell(row, 10, tview.NewTableCell(fmt.Sprintf("%d", match.Jumps)))
		marketTable.SetCell(row, 11, tview.NewTableCell(match.BuyISK).SetTextColor(tcell.ColorRed))
		marketTable.SetCell(row, 12, tview.NewTableCell(match.Profit).SetTextColor(tcell.ColorGreen))
		marketTable.SetCell(row, 13, tview.NewTableCell(match.ProfitPerJump).SetTextColor(tcell.ColorGreen))
	}
}
