package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func testFunction() {

	fmt.Println("You are in the test function in my ui.go file")
}

func mainMenu(i *Images) {

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
		i.dockerPush(f)
		pages.SwitchToPage("Push")
	})

}
