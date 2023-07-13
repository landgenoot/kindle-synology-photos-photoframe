// Kindle Synology Photos Photoframe
//
// To start: ./photoframe http://192.168.50.57:5000/mo/sharing/RMVJ3g6t8
// To stop: killall photoframe
//
// URL can be obtained from:
// Synology Photos -> Pick album -> Sharing
// Privacy Settings must be "Public - Anyone with the link can view"
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

func main() {
	logFile := initLogger()
	defer logFile.Close()

	if len(os.Args) < 2 {
		fmt.Println("Error: No album url found.")
		fmt.Println("Usage: ./photoframe http://192.168.50.57:5000/mo/sharing/RMVJ3g6t8")
		return
	}

	initPowersave()
	shareLink, _ = url.Parse(os.Args[1])
	baseUrl, albumCode = parseShareLink(shareLink)
	log.Printf("Initialising album %v on %v", albumCode, shareLink.Hostname())

	for true {
		updatePhoto()
		checkBattery()
		seconds := nextWakeup(time.Now(), 6, 0)
		suspendToRam(seconds) // Loop will automatically continue after wake up
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
	connectionErr := waitForWifi(shareLink.Hostname(), shareLink.Port())
	if connectionErr != nil {
		log.Printf("Could not connect to server, connectionError = %v", connectionErr)
		return
	}
	cookie, _ = getSharingSidCookie(shareLink)
	album, _ := fetchSynoAlbum(baseUrl, cookie, albumCode)
	randomPhoto, _ := getRandomPhoto(album)
	photoRequest, _ := getSynoPhotoRequest(baseUrl, cookie, albumCode, randomPhoto.Id)
	photo, _ := downloadPhoto(*photoRequest)
	convertPhoto(photo, "/tmp/photoframe.jpeg")
	drawToScreen("/tmp/photoframe.jpeg")
	log.Printf("Updating to photo %v", randomPhoto.Id)
}

func convertPhoto(photo []byte, filename string) {
	C.MagickWandGenesis()

	// Create a wand
	mwPhoto := C.NewMagickWand()
	mwKindleColors := C.NewMagickWand()
	pixelWand := C.NewPixelWand()

	// Tidy up wands after function run
	defer func() {
		if mwPhoto != nil {
			C.DestroyMagickWand(mwPhoto)
			C.DestroyMagickWand(mwKindleColors)
			C.DestroyPixelWand(pixelWand)
		}
		C.MagickWandTerminus()
	}()

	// Read photo and Kindle Colors reference file
	C.MagickReadImageBlob(mwPhoto, unsafe.Pointer(&photo[0]), C.size_t(len(photo)))
	C.MagickReadImageBlob(mwKindleColors, unsafe.Pointer(&kindle_colors[0]), C.size_t(len(kindle_colors)))

	// Crop and resize, rotate
	C.MagickSetImageGravity(mwPhoto, C.CenterGravity)
	mwPhoto = C.MagickTransformImage(mwPhoto, C.CString("1448x1072+0+0"), C.CString(""))
	C.MagickRotateImage(mwPhoto, pixelWand, 90)

	// Convert to grayscale and apply dithering
	C.MagickTransformImageColorspace(mwPhoto, C.GRAYColorspace)
	C.MagickRemapImage(mwPhoto, mwKindleColors, C.FloydSteinbergDitherMethod)
	C.MagickSetImageCompressionQuality(mwPhoto, C.size_t(75))

	// Adjust brightness and color depth
	image := C.GetImageFromMagickWand(mwPhoto)
	C.BrightnessContrastImage(image, 3, 15)
	C.SetImageDepth(image, C.size_t(8))

	// Write the new image
	C.MagickWriteImage(mwPhoto, C.CString(*&filename))
}

func getRandomPhoto(album synoFotoBrowseItem) (Photo, error) {
	if len(album.Data.List) < 1 {
		return Photo{}, errors.New("No photos in album")
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

func max(x, y float32) float32 {
	if x > y {
		return x
	}
	return y
}

// Estabilish TCP connection to Synology NAS and time out after 30 seconds.
// This enables the Kindle to connect to wifi.
func waitForWifi(hostname string, port string) error {
	seconds := 30
	timeOut := time.Duration(seconds) * time.Second
	_, err := net.DialTimeout("tcp", hostname+":"+port, timeOut)
	return err
}

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
