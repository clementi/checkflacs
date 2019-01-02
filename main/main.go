package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
)

// Config is program configuration
type Config struct {
	PathToFlac string `json:"pathToFlac"`
}

func main() {
	app := cli.NewApp()
	app.Name = "checkflacs"
	app.Usage = "check FLAC files for errors"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config",
			Value: "config.json",
			Usage: "path to config file",
		},
	}

	app.Action = func(c *cli.Context) error {
		root := c.Args().First()

		var config Config
		configFile, err := os.Open("config.json")
		defer configFile.Close()
		if err != nil {
			log.Fatal(err)
		}
		jsonParser := json.NewDecoder(configFile)
		jsonParser.Decode(&config)

		paths := make(chan string)
		results := make(chan string)

		go getPaths(root, paths)

		go checkFiles(config, paths, results)

		handleResults(results)

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func getPaths(root string, paths chan string) {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".flac") {
			paths <- path
		}
		return err
	})

	if err != nil {
		log.Fatal(err)
	}
}

func checkFiles(config Config, paths chan string, results chan string) {
	for path := range paths {
		cmd := exec.Command(config.PathToFlac, "-t", path)
		go run(cmd, results)
	}
}

func run(cmd *exec.Cmd, results chan string) {
	err := cmd.Run()

	if err != nil {
		results <- fmt.Sprintf("%s: %s", err, cmd.Args[len(cmd.Args)-1])
	} else {
		results <- fmt.Sprintf("OK: %s", cmd.Args[len(cmd.Args)-1])
	}
}

func handleResults(results chan string) {
	for result := range results {
		log.Println(result)
	}
}
