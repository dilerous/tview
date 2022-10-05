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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
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

func init() {

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

	//i := Images{}
	runTview()
	//mainMenu(&i)

	/*
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
	*/

}

/*
	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		log.Println(err)
		panic(err)
	}
}
*/

/*
func (i *Images) mainMenu() {

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
		i.dockerPush(f)
		pages.SwitchToPage("Push")
	})

}
*/

// s []string is source image.
// t string is the target which is pulled from registry input field.
func tagImages(i *Images, s []string) {
	InfoLogger.Printf("The value of the slice is: %v", s)

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			handlePanic(err)
		}
	}()

	var target []string

	if i.server == "" {
		i.server = "docker.io"
	}

	for _, v := range s {

		splitString := strings.Split(v, "/")
		lenString := len(splitString)
		image := splitString[lenString-1]

		err := cli.ImageTag(ctx, v, i.server+"/"+i.registry+"/"+image)
		target = append(target, i.server+"/"+i.registry+"/"+image)
		i.tagged = target
		if err != nil {
			ErrorLogger.Println(err)
			panic(err)
		}
	}

	sString := strings.Join(target, "\n")
	text.SetTextColor(tcell.ColorWhite).
		SetText(sString)
}

// s []string is list of images to push
func pushImages(i *Images) {

	if i.tagged == nil {
		log.Printf("There are no tagged images: %v", i.tagged)
		text.SetText("There are no tagged images. Please Tag Images and try again.").
			SetTextColor(tcell.ColorRed)
	}

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

		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				handlePanic(err)
			}
		}()

		r, err := cli.ImagePush(ctx, image, types.ImagePushOptions{RegistryAuth: authStr})
		if err != nil {
			updateText(nil, err)
			panic(err)
		}
		defer r.Close()
		io.Copy(text, r)
		log.Println(text)
		text.SetText("Success! All Images uploaded successfully.").
			SetTextColor(tcell.ColorGreen)
	}()
}

// Returns all images by Repo Tag as a []string slice
func listImages() {

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			handlePanic(err)
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
	/*
		text.SetText("Success! All Images downloaded successfully. Check logs for details.").
			SetTextColor(tcell.ColorGreen).
			SetWordWrap(true)
	*/
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
				InfoLogger.Println(err)
				handlePanic(err)
			}
		}()

		out, err := c.ImagePull(ctx, s, types.ImagePullOptions{RegistryAuth: authStr})
		if err != nil {
			panic(err)
		}
		defer out.Close()
		io.Copy(text, out)
		log.Println(text)
	}()
}

/*
func (i *Images) dockerPush(s []string) {
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
		listImages()
	})
}
*/

func handlePanic(err interface{}) {
	InfoLogger.Println("In the handlePanic function")
	text.SetTextColor(tcell.ColorRed).
		SetText(fmt.Sprint(err))
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
