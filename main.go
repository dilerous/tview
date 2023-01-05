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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	ctx    = context.Background()
	cli, _ = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
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

	//initKube()
	runTview()

}

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
	setText(sString, "white")
}

// s []string is list of images to push
func pushImages(i *Images) {

	if i.tagged == nil {
		log.Printf("There are no tagged images: %v", i.tagged)
		setText("There are no tagged images. Please Tag Images and try again.", "red")
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
		setText("Success! All Images uploaded successfully.", "green")

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
	setText(sString, "white")
}

// s []string is a slice of images
func (i *Images) pullImages(s []string) {

	if fmt.Sprint(s) == "[]" {
		log.Printf("No data was passed to the slice: %v\n", s)
		setText("Please input a valid Images File and try again", "red")
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
