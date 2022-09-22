package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	app   = tview.NewApplication()
	text  = tview.NewTextView()
	pages = tview.NewPages()
	flex  = tview.NewFlex()
	form  = tview.NewForm()
	menu  = tview.NewForm()
	ctx   = context.Background()
)

type Images struct {
	fileName string
	registry string
}

func main() {

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
	i.mainMenu()

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (i *Images) mainMenu() {

	menu.AddInputField("Input Images File: ", "", 34, nil, func(fileName string) {
		i.fileName = fileName
	}).
		AddButton("Quit", func() {
			app.Stop()
		}).
		AddButton("View File", func() {
			f, err := readFile(i.fileName)
			updateText(f, err)
		}).
		AddButton("Pull Images", func() {
			f, _ := readFile(i.fileName)
			pullImages(f)
		}).
		AddButton("Push Images", func() {
			f, _ := readFile(i.fileName)
			i.dockerPush(f)
			pages.SwitchToPage("Push")
		})
}

func listImages() []string {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	var imageId []string

	for _, image := range images {
		imageId = append(imageId, image.ID)
	}
	return imageId
}

func pullImages(s []string) {

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	for _, v := range s {
		out, err := cli.ImagePull(ctx, v, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}

		body, err := io.ReadAll(out)
		if err != nil {
			errorPage(err)
		}

		text.SetTextColor(tcell.ColorWhite).
			ScrollToEnd().
			SetScrollable(true).
			Write(body)
		defer out.Close()
	}
}

func (i *Images) dockerPush(s []string) {

	form.AddInputField("Registry: ", "", 20, nil, func(reg string) {
		i.registry = reg
	}).AddButton("Push to Registry", func() {
		i.dockerTag(s)
		for _, v := range s {
			cmdStr := "docker push " + v
			out, _ := exec.Command("/bin/sh", "-c", cmdStr).Output()
			text.Write(out)
		}
	}).AddButton("Return to Menu", func() {
		form.Clear(true)
		pages.SwitchToPage("Menu")
		app.SetFocus(menu)
	}).AddButton("List Images", func() {
		s := listImages()
		updateText(s, nil)
	})
	/*
		.AddButton("Pull Images", func() {
			f, _ := readFile(i.fileName)
			pullImages(f)
		})
	*/
}

func (i *Images) dockerTag(s []string) {

	registry := i.registry
	for _, v := range s {
		cmdStr := "docker tag " + v + registry + "/" + v
		out, _ := exec.Command("/bin/sh", "-c", cmdStr).Output()
		text.Write(out)
	}
}

func errorPage(e error) {

	text.SetTextColor(tcell.ColorRed).
		SetText(e.Error())
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

func readFile(f string) ([]string, error) {

	file, err := os.Open(f)
	if err != nil {
		errorPage(err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var images []string
	for scanner.Scan() {
		images = append(images, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return images, err
}
