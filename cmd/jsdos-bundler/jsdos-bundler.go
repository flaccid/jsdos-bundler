package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime/debug"
	"strings"
	"text/tabwriter"
	"time"

	jsdosbundler "github.com/flaccid/jsdos-bundler"
	"github.com/icza/bitio"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const (
	AUTHOR    = jsdosbundler.AUTHOR
	EMAIL     = jsdosbundler.EMAIL
	COPYRIGHT = jsdosbundler.COPYRIGHT
)

type Module struct {
	Path    string  // module path
	Version string  // module version
	Sum     string  // checksum
	Replace *Module // replaced by this module
}

var (
	version string
)

func beforeApp(c *cli.Context) error {
	log.Debugf("initialize jsdosbundler %s", version)

	switch c.String("log-level") {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "":
		log.SetLevel(log.InfoLevel)
	default:
		log.Fatalf("%s is an invalid log level", c.String("log-level"))
	}

	log.Debug("using log level " + log.GetLevel().String())

	if c.Bool("module-info") {
		_ = bitio.NewReader
		bi, ok := debug.ReadBuildInfo()
		if !ok {
			log.Fatal("failed to read build info")
		}
		writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
		fmt.Fprintln(writer, "VERSION\tCHECKSUM\tPATH\tREPLACED BY")
		for _, dep := range bi.Deps {
			// handle no module replace existing
			var rPath string
			r := reflect.ValueOf(dep.Replace)
			if r.IsNil() {
				rPath = "-"
			} else {
				rPath = dep.Replace.Path
			}
			line := fmt.Sprintf("%s\t%s\t%s\t%s",
				dep.Version,
				dep.Sum,
				dep.Path,
				rPath)
			fmt.Fprintln(writer, line)
		}
		writer.Flush()
		os.Exit(0)
	}

	return nil
}

// findEntryPoint scans a given directory for the first file ending in .exe, .com, or .bat.
// It returns the name of the first matching file found.
// If no suitable file is found, it returns the default string "ECHO".
// It also returns an error if the directory cannot be read.
func findEntryPoint(gameDir string) (string, error) {
	// Read all entries in the given directory
	entries, err := os.ReadDir(gameDir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory '%s': %w", gameDir, err)
	}

	// Define the valid executable extensions
	validExtensions := map[string]bool{
		".exe": true,
		".com": true,
		".bat": true,
	}

	// Loop through each entry in the directory
	for _, entry := range entries {
		// Skip subdirectories
		if entry.IsDir() {
			continue
		}

		// Get the file extension and convert it to lowercase for a case-insensitive check
		fileName := entry.Name()
		ext := strings.ToLower(filepath.Ext(fileName))

		// Check if the extension is in our set of valid extensions
		if validExtensions[ext] {
			// A valid entry point was found, return its name immediately
			return fileName, nil
		}
	}

	// If the loop completes, no suitable file was found in the directory
	return "ECHO", nil
}

func main() {
	app := &cli.App{
		Name:     "jsdosbundler",
		Version:  version,
		Compiled: time.Now(),
		Authors: []*cli.Author{
			{
				Name:  AUTHOR,
				Email: EMAIL,
			},
		},
		Copyright: COPYRIGHT,
		HelpName:  "jsdosbundler",
		Usage:     "create js-dos bundles",
		UsageText: "jsdosbundler [OPTIONS] COMMAND",
		Before:    beforeApp,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				Value:   "info",
				Usage:   "log level to use (debug,warn,error,info)",
				EnvVars: []string{"LOG_LEVEL"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "creates a js-dos bundle",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "game-dir",
						Aliases: []string{"g"},
						Value:   ".",
						Usage:   "game files directory",
						EnvVars: []string{"GAME_DIR"},
					},
					&cli.StringFlag{
						Name:    "entrypoint",
						Aliases: []string{"e"},
						Usage:   "final command in autoexec",
						EnvVars: []string{"ENTRYPOINT"},
					},
					&cli.StringFlag{
						Name:    "output-file",
						Aliases: []string{"o"},
						Usage:   "js-dos bundle output file path",
						EnvVars: []string{"OUTPUT"},
					},
				},
				Action: func(cCtx *cli.Context) error {
					if len(cCtx.Args().First()) < 1 {
						fmt.Println("Error: please provide a name for the bundle")
						os.Exit(1)
					}
					gameDir := cCtx.String("game-dir")
					entryPoint := cCtx.String("entrypoint")

					var err error
					if entryPoint == "" {
						entryPoint, err = findEntryPoint(gameDir)
						if err != nil {
							fmt.Printf("Error: %v\n", err)
							os.Exit(1)
						}
					}

					outputFile := cCtx.Args().First() + ".jsdos"
					if cCtx.String("output-file") != "" {
						outputFile = cCtx.String("output-file")
					}

					jsdosbundler.CreateBundle(gameDir, entryPoint, outputFile)

					return nil
				},
			},
			{
				Name:  "module-info",
				Usage: "shows compiled in go modules",
				Action: func(cCtx *cli.Context) error {
					_ = bitio.NewReader
					bi, ok := debug.ReadBuildInfo()
					if !ok {
						log.Fatal("failed to read build info")
					}
					writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
					fmt.Fprintln(writer, "VERSION\tCHECKSUM\tPATH\tREPLACED BY")
					for _, dep := range bi.Deps {
						// handle no module replace existing
						var rPath string
						r := reflect.ValueOf(dep.Replace)
						if r.IsNil() {
							rPath = "-"
						} else {
							rPath = dep.Replace.Path
						}
						line := fmt.Sprintf("%s\t%s\t%s\t%s",
							dep.Version,
							dep.Sum,
							dep.Path,
							rPath)
						fmt.Fprintln(writer, line)
					}
					writer.Flush()
					os.Exit(0)

					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
