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
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

var (
	ctx         = context.Background()
	cli         = initClient()
	dockerHosts = []string{
		"unix:///var/run/docker.sock",
		"unix:///Users/bsoper/.docker/run/docker.sock",
	}
)

type Images struct {
	fileName string
	registry string
	tag      []string
	username string
	password string
	server   string
	imageId  []string
}

type AuthConfig struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	ServerAddress string `json:"server,omitempty"`
}

func initClient() *client.Client {

	for _, host := range dockerHosts {
		cli, _ := client.NewClientWithOpts(client.WithHost(host),
			client.FromEnv, client.WithAPIVersionNegotiation())

		// Check if the client is working by listing containers
		if _, err := cli.ContainerList(context.Background(), container.ListOptions{}); err != nil {
			fmt.Printf("Failed to list containers for host %s: %v\n", host, err)
			continue
		}
		return cli
	}
	return nil
}

// s []string is source images.
// the string is the target which is pulled from registry input field.
// TODO make more descriptive possibly more descriptive function names
func (i *Images) tagImages(s []string) {
	InfoLogger.Printf("The value of the slice is: %v", s)

	var target []string

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			handlePanic(err)
		}
	}()

	if i.server == "" {
		i.server = "docker.io"
	}

	for _, v := range s {

		splitString := strings.Split(v, "/")
		lenString := len(splitString)
		image := splitString[lenString-1]

		err := cli.ImageTag(ctx, v, i.server+"/"+i.registry+"/"+image)
		target = append(target, i.server+"/"+i.registry+"/"+image)
		i.tag = target
		if err != nil {
			ErrorLogger.Println(err)
			panic(err)
		}
	}

	sString := strings.Join(target, "\n")
	setText(sString, "white")
}

// Pushes the tagged images into the registry defined by the user
func (i *Images) pushImages() {

	if i.tag == nil {
		log.Printf("There are no tagged images: %v", i.tag)
		setText("There are no tagged images. Please Tag Images and try again.", "red")
	}

	for _, v := range i.tag {
		i.streamPushToWriter(v)
	}
}

// takes the image as a string and streams to io.Writer
// Requires username and password to auth
func (i *Images) streamPushToWriter(image string) {

	authConfig := AuthConfig{
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

// s []string is a slice of images
func (i *Images) pullImages(s []string) {

	if fmt.Sprint(s) == "[]" {
		log.Printf("No data was passed to the slice: %v\n", s)
		setText("Please input a valid Images File and try again", "red")
	}

	for _, v := range s {
		i.streamPullToWriter(v)
	}

}

// takes images as a string and streams the update to text
func (i *Images) streamPullToWriter(s string) {

	var authConfig = AuthConfig{
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

		out, err := cli.ImagePull(ctx, s, types.ImagePullOptions{RegistryAuth: authStr})
		if err != nil {
			ErrorLogger.Println("There is a problem with the client", err)
			panic(err)
		}
		defer out.Close()
		io.Copy(text, out)
		InfoLogger.Println(text)
	}()
}

// Make list images specific to UI and get images specific to Docker
// Returns all images as a string seperated by a new line
func (i *Images) listImages() string {

	var repoTags []string

	defer func() {
		if err := recover(); err != nil {
			InfoLogger.Println(err)
			handlePanic(err)
		}
	}()

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		repoTags = append(repoTags, image.RepoTags...)
	}

	sString := strings.Join(repoTags, "\n")
	return sString
}

//Pass in the i Images struck
//Gets a list of Images on Docker host
//Grabs the Image ID and puts that into the stuc as a slice

func (i *Images) getImageId() {
	InfoLogger.Println("In the getImageID function")

	defer func() {
		if err := recover(); err != nil {
			ErrorLogger.Println(err)
			handlePanic(err)
		}
	}()

	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		i.imageId = append(i.imageId, image.RepoTags...)
	}

}

// Save the images pulled into a TAR file
// This function requires a slice of Image IDs
func (i *Images) saveImages() string {
	InfoLogger.Println("In the docker save function")

	defer func() {
		if err := recover(); err != nil {
			ErrorLogger.Println(err)
			handlePanic(err)
		}
	}()

	i.getImageId()

	f, err := os.Create("images.tar.gz")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	setText("Creating TAR file named images.tar.gz", "white")

	save, err := cli.ImageSave(ctx, i.imageId)

	if err != nil {
		ErrorLogger.Println(err)
		panic(err)
	}
	defer save.Close()

	io.Copy(w, save)
	f.Close()
	save.Close()
	return "TAR file successfully created"

}
