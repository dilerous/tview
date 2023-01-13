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
	pages   = tview.NewPages()
	flex    = tview.NewFlex()
	form    = tview.NewForm()
	menu    = tview.NewForm()
	start   = tview.NewForm()
	topText = tview.NewTextView()

	DEFAULT_USERNAME = "cnvrghelm"
)

func runTview() {

	text.SetBorder(true)

	pages.AddPage("Menu", menu, true, true).
		AddPage("View", text, true, false).
		AddPage("Push", form, true, false).
		AddPage("StartMenu", start, true, false).
		SetBorder(true)

	flex.AddItem(topText, 0, 1, true).
		AddItem(pages, 0, 3, true).
		AddItem(text, 0, 4, false)

	flex.SetDirection(tview.FlexRow)

	// FIXME Look to remove this code, not sure it does anything
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 113 {
			app.Stop()
		} else if event.Rune() == 27 {
			form.Clear(true)
			pages.SwitchToPage("Menu")
			app.SetFocus(menu)
			setText("You pressed ESC", "white")
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

	start.Clear(true)

	start.SetTitle(" Menu ").
		SetBorder(true).
		SetTitleColor(tcell.ColorGreen)

	start.AddButton("cnvrg Image Tool", func() {
		pages.SwitchToPage("Menu")
		app.SetFocus(menu)
	})
	initKube()

}

func mainMenu(i *Images) {

	i.username = DEFAULT_USERNAME
	startMenu()

	menu.SetBorder(true).
		SetTitle(" Menu ").
		SetTitleAlign(tview.AlignCenter).
		SetTitleColor(tcell.ColorGreen)

	text.SetText("Please enter the Docker Hub credientials provided by cnvrg.io to download the images needed.").
		SetWordWrap(true).
		SetTextColor(tcell.ColorWhite)

	topText.SetBorder(true).
		SetTitle("cnvrg.io Deployment Tool").
		SetTitleColor(tcell.ColorGreen)

	menu.AddInputField("cnvrg.io Docker Username: ", i.username, 40, nil, func(user string) {
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
	}).AddButton("Save Images to TAR", func() {
		i.saveImages()
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
		i.tagImages(f)
	}).AddButton("Push to Registry", func() {
		text.Clear()
		i.pushImages()
	}).AddButton("List Images", func() {
		text.Clear()
		setText(i.listImages(), "white")
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

// Prints to screen the text
// Define the color, options are white, red, green
func setTopText(s string, c string) {
	if c == "white" {
		topText.SetTextColor(tcell.ColorWhite).
			SetText(s)
	}
	if c == "red" {
		topText.SetTextColor(tcell.ColorRed).
			SetText(s)
	}
	if c == "green" {
		topText.SetTextColor(tcell.ColorGreen).
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

// Changes text to the color red and prints the error to the UI
func handlePanic(err interface{}) {
	InfoLogger.Println("In the handlePanic function")
	text.SetTextColor(tcell.ColorRed).
		SetText(fmt.Sprint(err))
}

// Changes text to the color red and prints the error to the UI
func handlePanicTop(err interface{}) {
	InfoLogger.Println("In the handlePanic function")
	topText.SetTextColor(tcell.ColorRed).
		SetText(fmt.Sprint(err))
}

/*
.AddButton("TAR Images", func() {
		fileToTar := []string{"tartestfile1.txt"}
		s := initArchive(fileToTar)
		if s != "" {
			setText(s, "white")
		}
.AddButton("Start Menu", func() {
		startMenu()
		pages.SwitchToPage("StartMenu")
	})
*/
