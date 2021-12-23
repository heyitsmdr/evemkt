package search

import (
	"eve-marketer/internal/eve"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"strings"
	"sync"
)

const (
	TableBuyOrders string = "buy"
	TableSellOrders string = "sell"
)

var tables = make(map[string]*tview.Table)
var searchString = ""
var wrapper *tview.Flex
var searchModal *tview.Modal
var ui *tview.Application

func Render(app *tview.Application) (flex *tview.Flex) {
	ui = app

	wrapper = tview.
		NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(renderTableView(), 0, 1, false).
		AddItem(renderSidePanel(), 40, 0, true)

	return wrapper
}

func ShowSearch() {
	searchSplit := strings.Split(searchString, "@")
	if len(searchSplit) != 2 {
		return
	}

	//searchType := searchSplit[0]
	searchRegion := searchSplit[1]

	regionInfo := eve.GetRegionInfoByName(searchRegion)
	if regionInfo == nil {
		return
	}

	//orders := eve.FetchAllRegionOrders(&eve.SearchOptions{RegionId: regionInfo.RegionId}, nil)
}

func renderSidePanel() (flex *tview.Flex) {
	flex = tview.NewFlex()

	flex.
		SetDirection(tview.FlexRow).
		AddItem(renderSearchPanel(), 7, 0, true).
		AddItem(renderInfoPanel(), 0, 1, false)

	return
}

func renderSearchPanel() (flex *tview.Flex) {
	flex = tview.NewFlex()

	flex.
		SetDirection(tview.FlexRow).
		AddItem(itemSearchBox(), 0, 1, true).
		AddItem(regionSearchBox(), 0, 1, false).
		SetTitle("Search").
		SetBorder(true).SetBorderPadding(1, 0, 1, 1)

	return
}

func itemSearchBox() (input *tview.InputField) {
	input = tview.NewInputField()

	var mutex sync.Mutex
	items := make([]string, 0)
	itemsRequested := false

	input.
		SetLabel("Item:").
		SetFieldWidth(0).
		SetAutocompleteFunc(func(currentText string) (entries []string) {
			if strings.TrimSpace(currentText) == "" {
				return
			}

			if itemsRequested {
				return
			}

			mutex.Lock()
			defer mutex.Unlock()
			if len(items) > 0 {
				for _, item := range items {
					txtLn := len(currentText)
					if txtLn <= len(item) && strings.ToLower(currentText[0:txtLn]) == strings.ToLower(item[0:txtLn]) {
						entries = append(entries, item)
					}
				}
				return
			}

			// Items list is empty. Let's populate it asynchronously.
			itemsRequested = true

			go func() {
				r := eve.GetItems()

				mutex.Lock()
				items = r
				mutex.Unlock()

				input.Autocomplete()

				ui.Draw()

				itemsRequested = false
			}()

			return nil
		})

	return
}

func regionSearchBox() (input *tview.InputField) {
	input = tview.NewInputField()

	var mutex sync.Mutex
	regions := make([]string, 0)
	regionsRequested := false

	input.
		SetLabel("Region:").
		SetFieldWidth(0).
		SetAutocompleteFunc(func(currentText string) (entries []string) {
			if strings.TrimSpace(currentText) == "" {
				return
			}

			if regionsRequested {
				return
			}

			mutex.Lock()
			defer mutex.Unlock()
			if len(regions) > 0 {
				for _, region := range regions {
					txtLn := len(currentText)
					if txtLn <= len(region) && strings.ToLower(currentText[0:txtLn]) == strings.ToLower(region[0:txtLn]) {
						entries = append(entries, region)
					}
				}
				return
			}

			// Regions list is empty. Let's populate it asynchronously.
			regionsRequested = true

			go func() {
				r := eve.GetRegions()

				mutex.Lock()
				regions = r
				mutex.Unlock()

				input.Autocomplete()

				ui.Draw()

				regionsRequested = false
			}()

			return nil
		})

	return
}

func renderInfoPanel() (flex *tview.Flex) {
	flex = tview.NewFlex()

	flex.
		SetTitle("Info").
		SetBorder(true)

	return
}

func renderSearchForm() (form *tview.Form) {
	form = tview.NewForm()

	form.
		AddInputField("Search", "", 20, nil, func(text string) {
			searchString = text
		}).
		SetTitle("Search").
		SetBorder(true)

	return
}

func renderTableView() (flex *tview.Flex) {
	flex = tview.NewFlex()

	flex.
		SetDirection(tview.FlexColumn).
		AddItem(renderTable(TableBuyOrders, "Buy Orders"), 0, 1, false).
		AddItem(renderTable(TableSellOrders, "Sell Orders"), 0, 1, false).
		SetTitle("Orders").
		SetBorder(true)

	return
}

func renderTable(name, title string) (tbl *tview.Table) {
	tbl = tview.NewTable()

	addTableHeaders(tbl)

	tbl.
		SetBorders(true).
		SetTitle(title).
		SetBorder(true)

	tables[name] = tbl

	return
}

func addTableHeaders(tbl *tview.Table) {
	tbl.
		SetCell(0, 0, tview.NewTableCell("Item").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft)).
		SetCell(0, 1, tview.NewTableCell("Volume").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft)).
		SetCell(0, 2, tview.NewTableCell("Price").SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignLeft)).
		SetCell(0, 3, tview.NewTableCell("Location").SetTextColor(tcell.ColorBlueViolet).SetAlign(tview.AlignLeft))
}