package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	ctx    = context.Background()
	cli, _ = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
)

type Images struct {
	fileName string
	registry string
	tag      []string
	username string
	password string
	server   string
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

// TODO think about breaking out the for loop to utils.go file
// Make list images specific to UI and get images specific to Docker
// Returns all images by Repo Tag as a []string slice
func listImages() {

	var imageId []string

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

	for _, image := range images {
		imageId = append(imageId, image.RepoTags...)
	}
	sString := strings.Join(imageId, "\n")
	setText(sString, "white")
}
