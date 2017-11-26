package main

import (
	"flag"
	"github.com/rwcarlsen/goexif/exif"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type AppConfig struct {
	DryRun bool
	Src    string
	Dst    string
}

var config AppConfig

func main() {
	InitConfig()
	doWork()
}

func doWork() {
	err := filepath.Walk(config.Src, processFile)
	if err != nil {
		log.Fatal(err)
	}
}

func processFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Print(err)
		return nil
	}

	if !info.IsDir() {
		moveFile(path, info)
	}

	return nil
}

func moveFile(path string, info os.FileInfo) {
	file_time := getFileTime(path, info)

	new_path := filepath.Join(config.Dst, file_time.Format("2006-01-02"), info.Name())

	log.Print(path + " -> " + new_path)

	if !config.DryRun {
		os.MkdirAll(filepath.Dir(new_path), os.ModePerm)

		old_file, err := os.Open(path)
		checkError(err)
		defer old_file.Close()

		new_file, err := os.Create(new_path)
		checkError(err)
		defer new_file.Close()

		_, err = io.Copy(new_file, old_file)
		checkError(err)

		err = new_file.Sync()
		checkError(err)
	}
}

func checkError(err error) {
	if err != nil {
		log.Printf("Error: %s", err.Error())
	}
}

func getFileTime(path string, info os.FileInfo) time.Time {
	extension := strings.ToLower(filepath.Ext(info.Name()))
	switch extension {
	case ".jpg", ".jpeg":
		return getJpegTime(path, info)
	}
	return info.ModTime()
}

func getJpegTime(path string, info os.FileInfo) time.Time {
	f, err := os.Open(path)

	checkError(err)

	if err == nil {
		x, err := exif.Decode(f)

		if err == nil {
			time, terr := x.DateTime()
			if terr == nil {
				return time
			}
		}
	}

	return info.ModTime()
}

func InitConfig() {
	config = AppConfig{}

	flag.BoolVar(&config.DryRun, "dry_run", true, "Do not perform actions")
	flag.StringVar(&config.Src, "src", "", "Path to source directory")
	flag.StringVar(&config.Dst, "dst", "", "Path to destination directory")

	flag.Parse()

	checkSrc(config.Src)
	checkDst(config.Dst)
}

func checkSrc(path string) bool {
	if config.Src == "" {
		log.Fatal("src is mandatory")
	}

	res, err := checkDirExists(path)
	if !res {
		if err == nil {
			log.Fatal("src is not directory")
		} else {
			log.Fatal("src dir does not exists")
		}
	}

	return true
}

func checkDst(path string) bool {
	if config.Src == "" {
		log.Fatal("dst is mandatory")
	}

	res, err := checkDirExists(path)
	if !res {
		if err == nil {
			log.Fatal("dst is not directory")
		} else {
			log.Fatal("dst dir does not exists")
		}
	}

	return true
}

func checkDirExists(path string) (bool, error) {
	stat, err := os.Stat(config.Src)
	if err == nil {
		if stat.IsDir() {
			return true, nil
		} else {
			return false, nil
		}
	}

	return false, err
}
