package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
)

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

func utilsErrorHandling(error interface{}) {
	ErrorLogger.Println(error)
	handlePanic(fmt.Sprint(error))
}

func initArchive(files []string) string {

	// Files which to include in the tar.gz archive

	defer func() {
		if err := recover(); err != nil {
			utilsErrorHandling(err)
		}
	}()

	// Create output file
	out, err := os.Create("output.tar.gz")
	if err != nil {
		ErrorLogger.Println(err)
		panic(err)
	}
	defer out.Close()

	// Create the archive and write the output to the "out" Writer
	err = createArchive(files, out)
	if err != nil {
		ErrorLogger.Println(err)
		panic(err)
	}

	success := "Archive created successfully"

	return success
}

func createArchive(files []string, buf io.Writer) error {

	/*
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				handlePanic(err)
			}
		}()
	*/

	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := addToArchive(tw, file)
		if err != nil {
			ErrorLogger.Println(err)
			panic(err)
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string) error {

	/*
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				handlePanic(err)
			}
		}()
	*/

	// Open the file which will be written into the archive
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get FileInfo about our file providing file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Use full path as name (FileInfoHeader only takes the basename)
	// If we don't do this the directory strucuture would
	// not be preserved
	// https://golang.org/src/archive/tar/common.go?#L626
	header.Name = filename

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}
