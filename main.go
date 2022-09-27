package main

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
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

type Images struct {
	fileName string
	registry string
	tagged   []string
	username string
	password string
	server   string
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
			text.Clear()
			pullImages(f)
		}).
		AddButton("Push Images", func() {
			f, _ := readFile(i.fileName)
			i.dockerPush(f)
			pages.SwitchToPage("Push")
		})
}

// s []string is source image.
// t string is the target which is pulled from registry input field.
func tagImages(s []string, t string) ([]string, error) {

	var target []string

	for _, v := range s {
		err := cli.ImageTag(ctx, v, t+"/"+v)
		target = append(target, t+"/"+v)
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
		ServerAddress: i.registry,
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
func listImages() []string {

	if err != nil {
		updateText(nil, err)
		panic(err)
	}

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	var imageId []string

	for _, image := range images {
		imageId = append(imageId, image.RepoTags...)
	}
	return imageId
}

// s []string is a slice of images
func pullImages(s []string) {

	if err != nil {
		panic(err)
	}
	for _, v := range s {
		streamPullToWriter(v, cli)
	}
}

// takes images as a string and streams the update to text
func streamPullToWriter(s string, c *client.Client) {
	go func() {
		out, err := c.ImagePull(ctx, s, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}
		defer out.Close()
		io.Copy(text, out)
	}()
}

func (i *Images) dockerPush(s []string) {

	form.AddInputField("Docker Username: ", "", 40, nil, func(user string) {
		i.username = user
	}).AddPasswordField("Docker Password: ", "", 40, 42, func(password string) {
		i.password = password
	}).AddInputField("Server Address: ", "https://index.docker.io/v1/", 40, nil, func(server string) {
		if server != "" {
			i.server = server
		} else {
			i.server = "https://index.docker.io/v1/"
		}
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
		s := listImages()
		updateText(s, nil)
	}).AddButton("Tag Images", func() {
		text.Clear()
		f, _ := readFile(i.fileName)
		i.tagged, err = tagImages(f, i.registry)
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
