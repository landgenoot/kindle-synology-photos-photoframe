package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"
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
	downloadPhoto(*photoRequest, "test.png")
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

func downloadPhoto(req http.Request, name string) error {
	client := &http.Client{}
	res, err := client.Do(&req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	out, err := os.Create(name)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, res.Body)
	return err
}
