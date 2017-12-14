package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime"
	"sync"
	"time"
)

var rName = ".txt"
var rContent = "php"
var maxSize, minSize int64
var files_ten []File

func main() {
	start := time.Now()

	channel_one := make(chan File)
	channel_two := make(chan File)

	var wg sync.WaitGroup
	var path string
	flag.StringVar(&path, "path", "", "Path to folder")
	flag.Parse()
	fmt.Println("Path=", path)

	for i := 0; i < runtime.NumCPU(); i++ {
		go check(channel_one, channel_two, &wg)
	}

	// go passTop10(channel_two, &g)

	getFolder(path, channel_one, &wg)

	wg.Wait()

	//fmt.Println("top 10" , files_ten)
	t := time.Now()
	current := t.Sub(start)
	fmt.Println(current)

}

type File struct {
	Size int64
	Name string
	Path string
}

func (this File) GetSize() int64 {
	return this.Size
}

func getFolder(path string, channel_one chan File, wg *sync.WaitGroup) {
	folder, err := ioutil.ReadDir(path)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, data := range folder {
		if data.IsDir() {
			var newFolder string = path + data.Name() + "/"
			getFolder(newFolder, channel_one, wg)
		} else {
			wg.Add(1)
			channel_one <- File{Size: data.Size(), Name: data.Name(), Path: path}
		}
	}
}

func check(channel_one chan File, channel_two chan File, wg *sync.WaitGroup) {
	for {
		file := <-channel_one
		rName := regexp.MustCompile(rName)

		maxSize = 10000
		minSize = 0

		if rName.MatchString(file.Name) {
			if file.Size <= maxSize && file.Size >= minSize {
				f, err := ioutil.ReadFile(file.Path + "/" + file.Name)

				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				rContent := regexp.MustCompile(rContent)
				if rContent.MatchString(string(f)) {
					channel_two <- file
				} else {
					wg.Done()
				}
			} else {
				wg.Done()
			}
		} else {
			wg.Done()
		}
	}
}
