package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
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
	pages    = tview.NewPages()
	flex     = tview.NewFlex()
	form     = tview.NewForm()
	menu     = tview.NewForm()
	ctx      = context.Background()
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

type Images struct {
	fileName string
	registry string
	tagged   []string
	username string
	password string
	server   string
}

func setLogger() {

	file, error := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if error != nil {
		log.Fatal(error)
	}
	log.SetOutput(file)
	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {

	text.SetBorder(true)
	setLogger()

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
		log.Println(err)
		panic(err)
	}
}

func (i *Images) mainMenu() {

	i.username = "cnvrghelm"

	menu.SetBorder(true).
		SetTitle(" cnvrg.io Deployment Tool ").
		SetTitleAlign(tview.AlignCenter).
		SetTitleColor(tcell.ColorGreen)

	text.SetText("Please enter the Docker Hub credientials provided by cnvrg.io to download the images needed.").
		SetWordWrap(true)

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

// s []string is source image.
// t string is the target which is pulled from registry input field.
func (i *Images) tagImages(s []string) ([]string, error) {

	text.SetText("Tagged Images: ")

	var target []string

	for _, v := range s {

		defer func() {
			if err := recover(); err != nil {
				text.SetText(fmt.Sprint(err)).
					SetTextColor(tcell.ColorRed)
				log.Println(err)

			}
		}()

		err := cli.ImageTag(ctx, v, i.server+"/"+i.registry+"/"+v)
		target = append(target, i.server+"/"+i.registry+"/"+v)
		if err != nil {
			panic(err)
		}
	}
	return target, err
}

// s []string is list of images to push
func (i *Images) pushImages() {

	for _, v := range i.tagged {
		i.streamPushToWriter(v)
	}
}

// takes the image as a string and streams to io.Writer
// Requires username and password to auth
func (i *Images) streamPushToWriter(image string) {

	var authConfig = types.AuthConfig{
		Username:      i.username,
		Password:      i.password,
		ServerAddress: i.server,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	go func() {
		r, err := cli.ImagePush(ctx, image, types.ImagePushOptions{RegistryAuth: authStr})
		if err != nil {
			updateText(nil, err)
			panic(err)
		}
		defer r.Close()
		io.Copy(text, r)
	}()
}

// Returns all images by Repo Tag as a []string slice
func listImages() {

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			text.SetText(fmt.Sprint(err)).
				SetTextColor(tcell.ColorRed)
		}
	}()

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	var imageId []string

	for _, image := range images {
		imageId = append(imageId, image.RepoTags...)
	}
	sString := strings.Join(imageId, "\n")
	text.SetTextColor(tcell.ColorWhite).
		SetText(sString)
}

// s []string is a slice of images
func (i *Images) pullImages(s []string) {

	if fmt.Sprint(s) == "[]" {
		log.Printf("No data was passed to the slice: %v\n", s)
		text.SetText("Please input a valid Images File and try again").
			SetTextColor(tcell.ColorRed)
	}

	for _, v := range s {
		i.streamPullToWriter(v, cli)
	}
}

// takes images as a string and streams the update to text
func (i *Images) streamPullToWriter(s string, c *client.Client) {

	var authConfig = types.AuthConfig{
		Username:      i.username,
		Password:      i.password,
		ServerAddress: i.server,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	go func() {

		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				text.SetText(fmt.Sprint(err)).
					SetTextColor(tcell.ColorRed)
			}
		}()

		out, err := c.ImagePull(ctx, s, types.ImagePullOptions{RegistryAuth: authStr})
		if err != nil {
			panic(err)
		}
		defer out.Close()
		io.Copy(text, out)
	}()
}

func (i *Images) dockerPush(s []string) {
	i.server = "docker.io"

	form.AddInputField("Docker Username: ", "", 40, nil, func(user string) {
		i.username = user
	}).AddPasswordField("Docker Password: ", "", 40, 42, func(password string) {
		i.password = password
	}).AddInputField("Server Address: ", "docker.io", 40, nil, func(server string) {
		i.server = server
	}).AddInputField("Registry: ", "", 40, nil, func(registry string) {
		i.registry = registry
	}).AddButton("Return to Menu", func() {
		form.Clear(true)
		pages.SwitchToPage("Menu")
		app.SetFocus(menu)
	}).AddButton("Push to Registry", func() {
		text.Clear()
		i.pushImages()
	}).AddButton("List Images", func() {
		text.Clear()
		listImages()
	}).AddButton("Tag Images", func() {
		text.Clear()
		f, _ := readFile(i.fileName)
		i.tagged, err = i.tagImages(f)
		updateText(i.tagged, err)
	})
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
		updateText(nil, err)
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

// Stream to byte example
/*
func streamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}
*/

/*
func (i *Images) dockerTag(s []string) {

	registry := i.registry
	for _, v := range s {
		cmdStr := "docker tag " + v + registry + "/" + v
		out, _ := exec.Command("/bin/sh", "-c", cmdStr).Output()
		text.Write(out)
	}
}
*/
