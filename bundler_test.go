package jsdosbundler

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestCreateBundle(t *testing.T) {
	if os.Getenv("LOG_LEVEL") == "debug" {
		log.SetLevel(log.DebugLevel)
	}

	entryPoint := "CAT.EXE"           // final autoexec
	gameDir := "test/game"            // game files
	outputFile := "test/bundle.jsdos" // output bundle file

	err := CreateBundle(gameDir, entryPoint, outputFile)
	if err != nil {
		t.Errorf("got %e, wanted nil", err)
	}
}
