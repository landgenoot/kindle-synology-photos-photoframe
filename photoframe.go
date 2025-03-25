// Kindle Photoframe - Synology NAS & Local Folder Support
//
// This application displays a random photo on a Kindle device.
// It supports two modes:
//   1. Synology NAS Album (public sharing link)
//   2. Local folder with JPEG/PNG images
//
// To start with Synology album:
//   ./photoframe http://192.168.50.57:5000/mo/sharing/RMVJ3g6t8
//
// To start with local photo folder:
//   ./photoframe /path/to/your/photos
//
// To stop:
//   killall photoframe
//
// Notes for Synology mode:
// - Go to Synology Photos -> Select album -> Share
// - Ensure Privacy Settings are set to:
//     "Public - Anyone with the link can view"
package main

// #cgo pkg-config: MagickWand
// #include <wand/MagickWand.h>
import "C"

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
	"unsafe"

	_ "embed"
)

//go:embed assets/kindle_colors.gif
var kindle_colors []byte
var shareLink *url.URL
var cookie *http.Cookie
var baseUrl string
var albumCode string
var isLocal bool
var localPath string

func main() {
	logFile := initLogger()
	defer logFile.Close()

	if len(os.Args) < 2 {
		fmt.Println("Error: No photo source provided.")
		fmt.Println("Usage for Synology: ./photoframe http://<ip>:<port>/mo/sharing/XYZ123")
		fmt.Println("Usage for Local:     ./photoframe /path/to/photo/folder")
		return
	}

	initPowersave()

	arg := os.Args[1]
	if strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://") {
		// Synology NAS mode
		isLocal = false
		shareLink, _ = url.Parse(arg)
		baseUrl, albumCode = parseShareLink(shareLink)
		log.Printf("Running in Synology NAS mode. Album %v on %v", albumCode, shareLink.Hostname())
	} else {
		// Local folder mode
		isLocal = true
		localPath = arg
		log.Printf("Running in LOCAL mode. Folder: %s", localPath)
	}

	for {
		updatePhoto()
		checkBattery()
		seconds := nextWakeup(time.Now(), 6, 0)
		suspendToRam(seconds)
	}
}

func initLogger() *os.File {
	f, err := os.OpenFile("photoframe.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
	return f
}

func updatePhoto() {
	var photo []byte
	var err error

	if isLocal {
		photo, err = getRandomLocalPhoto(localPath)
		if err != nil {
			log.Printf("Could not load local photo: %v", err)
			return
		}
		log.Printf("Displaying random local photo")
	} else {
		connectionErr := waitForWifi(shareLink.Hostname(), shareLink.Port())
		if connectionErr != nil {
			log.Printf("Could not connect to NAS server: %v", connectionErr)
			return
		}
		cookie, _ = getSharingSidCookie(shareLink)
		album, _ := fetchSynoAlbum(baseUrl, cookie, albumCode)
		randomPhoto, _ := getRandomPhoto(album)
		photoRequest, _ := getSynoPhotoRequest(baseUrl, cookie, albumCode, randomPhoto.Id)
		photo, _ = downloadPhoto(*photoRequest)
		log.Printf("Updating to Synology NAS photo %v", randomPhoto.Id)
	}

	convertPhoto(photo, "/tmp/photoframe.jpeg")
	drawToScreen("/tmp/photoframe.jpeg")
}

func getRandomLocalPhoto(folderPath string) ([]byte, error) {
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var imageFiles []string
	for _, file := range files {
		if !file.IsDir() {
			ext := strings.ToLower(path.Ext(file.Name()))
			switch ext {
			case ".jpg", ".jpeg", ".png":
				imageFiles = append(imageFiles, path.Join(folderPath, file.Name()))
			}
		}
	}

	if len(imageFiles) == 0 {
		return nil, errors.New("no image files found in folder")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := r.Intn(len(imageFiles))
	return os.ReadFile(imageFiles[randomIndex])
}

func convertPhoto(photo []byte, filename string) {
	C.MagickWandGenesis()
	mwPhoto := C.NewMagickWand()
	mwKindleColors := C.NewMagickWand()
	pixelWand := C.NewPixelWand()

	defer func() {
		if mwPhoto != nil {
			C.DestroyMagickWand(mwPhoto)
			C.DestroyMagickWand(mwKindleColors)
			C.DestroyPixelWand(pixelWand)
		}
		C.MagickWandTerminus()
	}()

	C.MagickReadImageBlob(mwPhoto, unsafe.Pointer(&photo[0]), C.size_t(len(photo)))
	C.MagickReadImageBlob(mwKindleColors, unsafe.Pointer(&kindle_colors[0]), C.size_t(len(kindle_colors)))

	C.MagickSetImageGravity(mwPhoto, C.CenterGravity)
	mwPhoto = C.MagickTransformImage(mwPhoto, C.CString("1448x1072+0+0"), C.CString(""))
	C.MagickRotateImage(mwPhoto, pixelWand, 90)

	C.MagickTransformImageColorspace(mwPhoto, C.GRAYColorspace)
	C.MagickRemapImage(mwPhoto, mwKindleColors, C.FloydSteinbergDitherMethod)
	C.MagickSetImageCompressionQuality(mwPhoto, C.size_t(75))

	image := C.GetImageFromMagickWand(mwPhoto)
	C.BrightnessContrastImage(image, 3, 15)
	C.SetImageDepth(image, C.size_t(8))

	C.MagickWriteImage(mwPhoto, C.CString(filename))
}

// -------- Synology photo-related functions (kept unchanged for compatibility) --------

func getRandomPhoto(album synoFotoBrowseItem) (Photo, error) {
	if len(album.Data.List) < 1 {
		return Photo{}, errors.New("no photos in album")
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomIndex := r.Intn(len(album.Data.List))
	return album.Data.List[randomIndex], nil
}

func isCached(id int, cachePath string) bool {
	_, err := os.Stat(path.Join(cachePath, fmt.Sprintf("%d.png", id)))
	return err == nil
}

func downloadPhoto(req http.Request) ([]byte, error) {
	client := &http.Client{}
	res, _ := client.Do(&req)
	return ioutil.ReadAll(res.Body)
}

func waitForWifi(hostname string, port string) error {
	timeout := 30 * time.Second
	_, err := net.DialTimeout("tcp", hostname+":"+port, timeout)
	return err
}

// -------- Battery check --------

func checkBattery() {
	state := getBatteryLevel()
	level, err := parseBatteryLevel(state)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Battery level %d %%", level)
	if level <= 15 {
		drawLowBatteryIndicator()
	}
}
