package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"
)

func main() {
	fmt.Println("Hello, World!")

	shareLink := "https://b92.dsmdemo.synologydemo.com:5001/mo/sharing/k5SnJvlVW"
	baseUrl, albumCode := parseShareLink(shareLink)
	cookie, _ := getSharingSidCookie(shareLink)
	album, _ := fetchSynoAlbum(baseUrl, cookie, albumCode)
	fmt.Println(album)
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

func downloadPhoto(baseUrl string, albumCode string, id int, cachePath string, cookie *http.Cookie) error {
	return nil
}
