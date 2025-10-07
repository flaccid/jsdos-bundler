package jsdosbundler

import (
	"archive/zip"
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
func CreateBundle(gameDir, outputFile string) error {
	log.WithFields(log.Fields{
		"gameDir":    gameDir,
		"outputFile": outputFile,
	}).Debug("create js-dos bundle")

	// Create the output .jsdos file (which is a zip file)
	newZipFile, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add the .jsdos directory and dosbox.conf
	// Create the .jsdos directory in the zip archive
	_, err = zipWriter.Create(".jsdos/")
	if err != nil {
		panic(err)
	}

	// Create the dosbox.conf file in the .jsdos directory
	dosboxConfWriter, err := zipWriter.Create(".jsdos/dosbox.conf")
	if err != nil {
		panic(err)
	}
	// Write the dosbox.conf content. You can read this from a template file.
	// For simplicity, we'll write a string here.
	dosboxConfContent := []byte("[autoexec]\n@echo off\nmount c .\nC:\nYOURGAME.EXE\n")
	_, err = dosboxConfWriter.Write(dosboxConfContent)
	if err != nil {
		panic(err)
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
