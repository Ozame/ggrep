package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

func main() {
	//TODO: Custom -help flag with usage examples
	recursive := flag.Bool("r", false, "If true, search will progress recursively through all directories in the target path")
	hiddenFiles := flag.Bool("hidden", false, "If true, hidden files and directories will also be searched. Currently only supported on linux/unix systems.")
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		log.Fatalln("No search expression specified")
	}
	if len(args) < 2 {
		log.Fatalln("No search path specified")
	}
	expr, _ := regexp.Compile(args[0])
	cleanedPath := filepath.Clean(args[1])
	var wg sync.WaitGroup
	wg.Add(1)
	go checkPath(*expr, cleanedPath, *recursive, *hiddenFiles, &wg)
	wg.Wait()

}

func checkPath(expression regexp.Regexp, fileName string, isRecursive bool, hiddenFiles bool, wg *sync.WaitGroup) {
	defer wg.Done()
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		log.Fatalf("Failed to stat file with name %s, error: %v\n", fileName, err)
	}
	if fileInfo.IsDir() {
		dirs, err := os.ReadDir(fileName)
		if err != nil {
			log.Fatal("Could not read the directory")
		}
		for _, child := range dirs {
			childPath := filepath.Join(fileName, child.Name())
			if !isHidden(childPath) || hiddenFiles {
				if isRecursive && child.IsDir() {
					wg.Add(1)
					go checkPath(expression, childPath, isRecursive, hiddenFiles, wg)
				} else if !child.IsDir() {
					wg.Add(1)
					go searchFile(expression, childPath, wg)
				}
			}
		}
	} else {
		if !isHidden(fileName) || hiddenFiles {
			wg.Add(1)
			go searchFile(expression, fileName, wg)
		}
	}
}

func searchFile(expression regexp.Regexp, fileName string, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Could not find file %s\n", fileName)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var matches []string
	for scanner.Scan() {
		if text := scanner.Text(); expression.MatchString(text) {
			matches = append(matches, text)
		}
	}
	//TODO: Option to show line numbers/columns
	//TODO: Option to show filename?
	if 0 < len(matches) {
		for _, line := range matches {
			println(line)
		}
	}

}

func isHidden(filePath string) bool {
	path := filepath.Base(filePath)
	if runtime.GOOS == "windows" {
		//TODO: Windows checking requires a separate source file
		return false
	} else {
		if path == "." || strings.HasPrefix(path, "..") {
			return false
		} else {
			return path[:1] == "."
		}
	}
}
