package jsdosbundler

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const (
	AUTHOR    = "Chris Fordham"
	EMAIL     = "chris@fordham.id.au"
	COPYRIGHT = "(c) Chris Fordham"
)

// CreateBundle creates a js-dos bundle (zip file) given game directory and output filename.
// It returns any error if encountered.
func CreateBundle(gameDir, entryPoint, outputFile string) error {
	log.WithFields(log.Fields{
		"gameDir":    gameDir,
		"entryPoint": entryPoint,
		"outputFile": outputFile,
	}).Debug("create js-dos bundle")

	// Create the output .jsdos file (which is a zip file)
	newZipFile, err := os.Create(outputFile)
	if err != nil {
		return errors.New("error creating bundle (zip) file: " + err.Error())
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add the .jsdos directory and dosbox.conf
	// Create the .jsdos directory in the zip archive
	_, err = zipWriter.Create(".jsdos/")
	if err != nil {
		return errors.New("unable to create .jsdos folder in bundle (zip) file: " + err.Error())
	}

	// Create the dosbox.conf file in the .jsdos directory
	dosboxConfWriter, err := zipWriter.Create(".jsdos/dosbox.conf")
	if err != nil {
		return errors.New("unable to create .jsdos/dosbox.conf in bundle (zip) file: " + err.Error())
	}
	// Write the dosbox.conf content. You can read this from a template file.
	// For simplicity, we'll write a string here.
	dosboxConfContent := []byte("[autoexec]\n@echo off\nmount c .\nC:\n" + entryPoint + "\n")
	log.WithFields(log.Fields{
		"dosboxConfContent": string(dosboxConfContent),
	}).Debug("create dosbox.conf")
	_, err = dosboxConfWriter.Write(dosboxConfContent)
	if err != nil {
		return errors.New("unable to write dosbox.conf: " + err.Error())
	}

	// Walk through the game directory and add files to the zip
	err = filepath.Walk(gameDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(gameDir, filePath)
		if err != nil {
			return err
		}

		zipFile, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		fsFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer fsFile.Close()

		_, err = io.Copy(zipFile, fsFile)
		return err
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully created js-dos bundle:", outputFile)

	return nil
}
