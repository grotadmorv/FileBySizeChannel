package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"runtime"
	"runtime/pprof"
	"sort"
	"sync"
	"time"
)

var rName = ".php"
var rContent = "php"
var maxSize, minSize int64
var files_ten []File

func main() {
	f, err := os.Create("perf_cpu.perf")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	start := time.Now()

	channelOne := make(chan File)
	channelTwo := make(chan File)

	var wg sync.WaitGroup
	var path string
	flag.StringVar(&path, "path", "", "Path to folder")
	flag.Parse()
	fmt.Println("Path=", path)

	for i := 0; i < runtime.NumCPU(); i++ {
		go check(channelOne, channelTwo, &wg)
	}
	go top10(channelTwo, &wg)

	getFolder(path, channelOne, &wg)

	wg.Wait()

	fmt.Println("top 10", files_ten)
	t := time.Now()
	current := t.Sub(start)
	fmt.Println(current)

	f, err = os.Create("mem_profile.perf")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}

	f.Close()
}

type File struct {
	Size int64
	Name string
	Path string
}

func (this File) GetSize() int64 {
	return this.Size
}

func getFolder(path string, channelOne chan File, wg *sync.WaitGroup) {
	folder, err := ioutil.ReadDir(path)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, data := range folder {
		if data.IsDir() {
			var newFolder string = path + data.Name() + "/"
			getFolder(newFolder, channelOne, wg)
		} else {
			wg.Add(1)
			channelOne <- File{Size: data.Size(), Name: data.Name(), Path: path}
		}
	}
}

func check(channelOne chan File, channelTwo chan File, wg *sync.WaitGroup) {
	for {
		file := <-channelOne
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
					channelTwo <- file
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
func sortFilesFromBiggestToLowerSize(arrayFile []File) []File {
	sort.Slice(arrayFile, func(i, j int) bool {
		return arrayFile[i].Size > arrayFile[j].Size
	})
	return arrayFile
}

func top10(channelTwo chan File, wg *sync.WaitGroup) []File {
	for {
		f := <-channelTwo
		if len(files_ten) == 10 {
			if f.Size > files_ten[0].Size || f.Size > files_ten[len(files_ten)-1].Size {
				files_ten = files_ten[:len(files_ten)-1]
				files_ten = append(files_ten, f)
				files_ten = sortFilesFromBiggestToLowerSize(files_ten)
			}
		} else {
			sortFilesFromBiggestToLowerSize(files_ten)
			files_ten = append(files_ten, f)
		}
		wg.Done()
	}
}
