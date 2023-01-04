package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app  = tview.NewApplication()
	text = tview.NewTextView().
		SetTextColor(tcell.ColorWhite).
		ScrollToEnd().
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	pages = tview.NewPages()
	flex  = tview.NewFlex()
	form  = tview.NewForm()
	menu  = tview.NewForm()
)

func runTview() {

	text.SetBorder(true)

	pages.AddPage("Menu", menu, true, true).
		AddPage("View", text, true, false).
		AddPage("Push", form, true, false).
		SetBorder(true)

	flex.AddItem(pages, 0, 4, true).
		AddItem(text, 0, 4, false)

	flex.SetDirection(tview.FlexRow)

	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 113 {
			app.Stop()
		} else if event.Rune() == 27 {
			form.Clear(true)
			pages.SwitchToPage("Menu")
			app.SetFocus(menu)
		}
		return event
	})

	i := Images{}
	mainMenu(&i)

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		log.Println(err)
		panic(err)
	}
}

func startMenu() {

	text.SetText("testing").SetTextColor(tcell.ColorWhite).
		SetBorder(true).
		SetTitle("cnvrg.io Deployment Tool").
		SetTitleColor(tcell.ColorGreen)

	initKube()

	flex.Clear()
	flex.AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(text, 0, 1, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Middle (3 x height of Top)"), 0, 3, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Bottom (5 rows)"), 5, 1, false), 0, 2, false)

}

func mainMenu(i *Images) {

	i.username = "cnvrghelm"

	menu.SetBorder(true).
		SetTitle(" cnvrg.io Deployment Tool ").
		SetTitleAlign(tview.AlignCenter).
		SetTitleColor(tcell.ColorGreen)

	text.SetText("Please enter the Docker Hub credientials provided by cnvrg.io to download the images needed.").
		SetWordWrap(true).
		SetTextColor(tcell.ColorWhite)

	menu.AddInputField("cnvrg.io Docker Username: ", "cnvrghelm", 40, nil, func(user string) {
		i.username = user
	}).AddPasswordField("cnvrg.io Docker Password: ", "", 40, 42, func(password string) {
		i.password = password
	}).AddInputField("Images File: ", "", 40, nil, func(fileName string) {
		i.fileName = fileName
	}).AddButton("Quit", func() {
		app.Stop()
	}).AddButton("View File", func() {
		f, err := readFile(i.fileName)
		updateText(f, err)
	}).AddButton("Pull Images", func() {
		f, _ := readFile(i.fileName)
		text.Clear()
		i.pullImages(f)
	}).AddButton("Push Images", func() {
		f, _ := readFile(i.fileName)
		pushMenu(i, f)
		pages.SwitchToPage("Push")
	}).AddButton("Start Menu", func() {
		startMenu()
	})
}

func pushMenu(i *Images, s []string) {
	i.server = "docker.io"

	form.SetBorder(true).
		SetTitle(" cnvrg.io Deployment Tool ").
		SetTitleAlign(tview.AlignCenter).
		SetTitleColor(tcell.ColorGreen)

	text.SetText("Please enter the private registry credientials to push images.").
		SetWordWrap(true).
		SetTextColor(tcell.ColorWhite)

	form.AddInputField("Docker Username: ", "", 40, nil, func(user string) {
		i.username = user
	}).AddPasswordField("Docker Password: ", "", 40, 42, func(password string) {
		i.password = password
	}).AddInputField("Server Address: ", "docker.io", 40, nil, func(server string) {
		i.server = server
	}).AddInputField("Registry: ", "", 40, nil, func(registry string) {
		i.registry = registry
	}).AddButton("Return to Main Menu", func() {
		form.Clear(true)
		pages.SwitchToPage("Menu")
		app.SetFocus(menu)
	}).AddButton("Tag Images", func() {
		text.Clear()
		f, _ := readFile(i.fileName)
		tagImages(i, f)
	}).AddButton("Push to Registry", func() {
		text.Clear()
		pushImages(i)
	}).AddButton("List Images", func() {
		text.Clear()
		listImages()
	})
}

// Prints to screen the text
// Define the color, options are white, red, green
func setText(s string, c string) {
	if c == "white" {
		text.SetTextColor(tcell.ColorWhite).
			SetText(s)
	}
	if c == "red" {
		text.SetTextColor(tcell.ColorRed).
			SetText(s)
	}
	if c == "green" {
		text.SetTextColor(tcell.ColorGreen).
			SetText(s)
	}

}

func updateText(s []string, e error) {

	if e != nil {
		text.SetTextColor(tcell.ColorRed).
			SetText(e.Error())
	} else {
		sString := strings.Join(s, "\n")
		text.SetTextColor(tcell.ColorWhite).
			SetText(sString)
	}
}

func handlePanic(err interface{}) {
	InfoLogger.Println("In the handlePanic function")
	text.SetTextColor(tcell.ColorRed).
		SetText(fmt.Sprint(err))
}
