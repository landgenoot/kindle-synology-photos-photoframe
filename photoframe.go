package main

// #cgo pkg-config: MagickWand
// #include <wand/MagickWand.h>
import "C"

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"
	"unsafe"

	_ "embed"
)

//go:embed bin/kindle_colors.gif
var kindle_colors []byte

var (
	inputFile  = flag.String("in", "test.png", "The file to read in")
	outputFile = flag.String("out", "test2.png", "The file to write out")
	scale      = flag.Float64("scale", 0.5, "how much to scale the image by")
)

func main() {
	fmt.Println("Hello, World!")

	shareLink := "http://192.168.50.57:5000/mo/sharing/RMVJ3g6t8"
	baseUrl, albumCode := parseShareLink(shareLink)
	cookie, _ := getSharingSidCookie(shareLink)
	album, _ := fetchSynoAlbum(baseUrl, cookie, albumCode)
	randomPhoto, _ := getRandomPhoto(album)
	photoRequest, _ := getSynoPhotoRequest(baseUrl, cookie, albumCode, randomPhoto.Id)
	fmt.Println(photoRequest.URL.String())
	fmt.Println(cookie)
	buf, _ := downloadPhoto(*photoRequest)
	convertPhoto(buf, "a.jpeg")
	path, _ := os.Getwd()
	cmd := exec.Command("/usr/sbin/eips", "-f", "-g", fmt.Sprintf(`%v/test2.png`, path))

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
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
