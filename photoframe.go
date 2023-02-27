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
