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
	"os/exec"
	"path"
	"runtime"
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

	shareLink, _ = url.Parse(os.Args[1])
	baseUrl, albumCode = parseShareLink(shareLink)
	log.Printf("Initialising album %v on %v", albumCode, shareLink.Hostname())

	for true {
		updatePhoto()
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

// Count seconds till next wake up time. Formatted as clock
// time in 24H format. E.g. 6, 30 means 6:30 AM.
func nextWakeup(now time.Time, hour int, minutes int) int {
	yyyy, mm, dd := now.Date()
	if now.Hour() > hour || now.Hour() == hour && now.Minute() >= minutes {
		dd++ // Jump to tomorrow, if wakeup time has already passed.
	}
	tomorrow := time.Date(yyyy, mm, dd, hour, minutes, 0, 0, now.Location())
	return int(tomorrow.Sub(now).Seconds())
}

func drawToScreen(imagePath string) {
	if runtime.GOARCH != "arm" {
		return // Skip if not on Kindle
	}
	cmd := exec.Command("/usr/sbin/eips", "-f", "-g", imagePath)
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// Suspend device and use real time clock alarm to wake it up.
// If our wake up time is more or less 24 hours away, we can put it to
// sleep immediately. Otherwise, we will wait another 30 seconds, which enables us
// to abort the process.
func suspendToRam(duration int) {
	if runtime.GOARCH != "arm" {
		return // Skip if not on Kindle
	}
	cmd1 := exec.Command("sh", "-c", "echo \"\" > /sys/class/rtc/rtc1/wakealarm")
	err1 := cmd1.Run()
	if err1 != nil {
		log.Fatal(err1)
	}
	cmd2 := exec.Command("sh", "-c", fmt.Sprintf("echo \"+%d\" > /sys/class/rtc/rtc1/wakealarm", duration))
	err2 := cmd2.Run()
	if err2 != nil {
		log.Fatal(err2)
	}

	// Check if we are waken up manually, give us time to abort the process
	if duration < 3600*24-60 {
		log.Println("Waiting 30 seconds before going back to sleep")
		time.Sleep(30 * time.Second)
	}

	log.Println("Suspending to RAM")

	cmd3 := exec.Command("sh", "-c", "echo \"mem\" > /sys/power/state")
	err3 := cmd3.Run()
	if err3 != nil {
		log.Fatal(err3)
	}
}

// Estabilish TCP connection to Synology NAS and time out after 30 seconds.
// This enables the Kindle to connect to wifi.
func waitForWifi(hostname string, port string) error {
	seconds := 30
	timeOut := time.Duration(seconds) * time.Second
	_, err := net.DialTimeout("tcp", hostname+":"+port, timeOut)
	return err
}
