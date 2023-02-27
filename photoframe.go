package main

// #cgo pkg-config: MagickWand
// #include <wand/MagickWand.h>
import "C"
import (
	"flag"
	"fmt"
	"math"
	"os"
)

var (
	inputFile  = flag.String("in", ":logo", "The file to read in")
	outputFile = flag.String("out", "logo.jpg", "The file to write out")
	scale      = flag.Float64("scale", 0.5, "how much to scale the image by")
)

func main() {
	fmt.Println("Hallo")
	flag.Parse()

	if *scale <= 0 {
		fmt.Fprintf(os.Stderr, "Scale must be greater than zero\n")
		os.Exit(1)
	}

	C.MagickWandGenesis()

	// Create a wand
	mw := C.NewMagickWand()

	defer func() {
		// Tidy up
		if mw != nil {
			C.DestroyMagickWand(mw)
		}

		C.MagickWandTerminus()
	}()

	// Read the input image
	C.MagickReadImage(mw, C.CString(*inputFile))

	// Resize it
	height := float64(C.MagickGetImageHeight(mw))
	width := float64(C.MagickGetImageWidth(mw))

	newHeight := math.Round(math.Max(height*(*scale), 1))
	newWidth := math.Round(math.Max(width*(*scale), 1))

	// Resize the image using the Lanczos filter
	// The blur factor is a "double", where > 1 is blurry, < 1 is sharp
	C.MagickResizeImage(mw, C.uint(newWidth), C.uint(newHeight), C.LanczosFilter, 1)

	// Write the new image
	C.MagickWriteImage(mw, C.CString(*outputFile))

}

// package main

// import (
// 	"errors"
// 	"fmt"
// 	"io"
// 	"math/rand"
// 	"net/http"
// 	"os"
// 	"path"
// 	"time"
// )

// func main() {
// 	fmt.Println("Hello, World!")

// 	shareLink := "http://192.168.50.57:5000/mo/sharing/RMVJ3g6t8"
// 	baseUrl, albumCode := parseShareLink(shareLink)
// 	cookie, _ := getSharingSidCookie(shareLink)
// 	album, _ := fetchSynoAlbum(baseUrl, cookie, albumCode)
// 	randomPhoto, _ := getRandomPhoto(album)
// 	photoRequest, _ := getSynoPhotoRequest(baseUrl, cookie, albumCode, randomPhoto.Id)
// 	fmt.Println(photoRequest.URL.String())
// 	fmt.Println(cookie)
// 	downloadPhoto(*photoRequest, "test.png")
// }

// func getRandomPhoto(album synoFotoBrowseItem) (Photo, error) {
// 	if len(album.Data.List) < 1 {
// 		return Photo{}, errors.New("No photos in album")
// 	}
// 	r := rand.New(rand.NewSource(time.Now().UnixNano()))
// 	randomIndex := r.Intn(len(album.Data.List))
// 	return album.Data.List[randomIndex], nil
// }

// func isCached(id int, cachePath string) bool {
// 	_, err := os.Stat(path.Join(cachePath, fmt.Sprintf("%d.png", id)))
// 	return err == nil
// }

// func downloadPhoto(req http.Request, name string) error {
// 	client := &http.Client{}
// 	res, err := client.Do(&req)
// 	if err != nil {
// 		return err
// 	}
// 	defer res.Body.Close()

// 	out, err := os.Create(name)
// 	if err != nil {
// 		return err
// 	}
// 	defer out.Close()
// 	_, err = io.Copy(out, res.Body)
// 	return err
// }
