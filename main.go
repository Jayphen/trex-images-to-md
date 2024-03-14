package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatal("You must provide one or more files as input")
	}

	for _, file := range args {
		mdFilename := fmt.Sprintf("%s.md", strings.TrimSuffix(file, filepath.Ext(file)))

		cmd := exec.Command("trex", "-i")
		inputFile, err := os.Open(file)
		if err != nil {
			log.Fatal(err)
		}
		defer inputFile.Close()

		cmd.Stdin = inputFile

		output, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}

		mdFile, err := os.Create(mdFilename)
		if err != nil {
			log.Fatal(err)
		}
		defer mdFile.Close()

		writer := bufio.NewWriter(mdFile)
		defer writer.Flush()

		_, err = writer.Write(output)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Created %s from %s\n", mdFilename, file)
	}
}
