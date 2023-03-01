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

	C.MagickWandGenesis()

	// Create a wand
	mw := C.NewMagickWand()

	mw2 := C.NewMagickWand()

	defer func() {
		// Tidy up
		if mw != nil {
			C.DestroyMagickWand(mw)
			C.DestroyMagickWand(mw2)
		}

		C.MagickWandTerminus()
	}()

	C.MagickReadImageBlob(mw, unsafe.Pointer(&buf[0]), C.size_t(len(buf)))
	C.MagickReadImageBlob(mw2, unsafe.Pointer(&kindle_colors[0]), C.size_t(len(kindle_colors)))
	C.MagickSetImageGravity(mw, C.CenterGravity)
	mw = C.MagickTransformImage(mw, C.CString("1448x1072+0+0"), C.CString(""))
	C.MagickTransformImageColorspace(mw, C.GRAYColorspace)
	C.MagickRemapImage(mw, mw2, C.FloydSteinbergDitherMethod)
	image := C.GetImageFromMagickWand(mw)
	C.BrightnessContrastImage(image, 3, 15)
	C.MagickSetImageCompressionQuality(mw, C.size_t(75))
	C.SetImageDepth(image, C.size_t(8))

	pixelWand := C.NewPixelWand()
	C.MagickRotateImage(mw, pixelWand, 90)
	// Write the new image
	C.MagickWriteImage(mw, C.CString(*outputFile))
	path, _ := os.Getwd()
	cmd := exec.Command("/usr/sbin/eips", "-f", "-g", fmt.Sprintf(`%v/test2.png`, path))

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
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
