package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func main() {
	fmt.Println("Hello, World!")
}

type synoFotoBrowseItem struct {
	Success bool `json:"success"`
	Data    Data `json:"data"`
}

type Data struct {
	List []Photo `json:"list"`
}

type Photo struct {
	Id int `jsoFotoBrowseItemn:"id"`
}

func fetchSynoAlbum(url string, cookie *http.Cookie, albumCode string) ([]int, error) {
	method := "POST"

	payload := strings.NewReader(`offset=0&limit=1000&api="SYNO.Foto.Browse.Item"&method="list"&version=1`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("x-syno-sharing", albumCode)
	req.Header.Add("Content-Type", "text/plain")
	req.AddCookie(cookie)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	album := synoFotoBrowseItem{}
	jsonErr := json.Unmarshal(body, &album)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	ids := make([]int, len(album.Data.List))

	for i := 0; i < len(ids); i++ {
		ids[i] = album.Data.List[i].Id
	}
	return ids, jsonErr
}

func getRandomPhotoId(ids []int) (int, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return ids[r.Intn(len(ids))], nil
}

func getSharingSidCookie(url string) (*http.Cookie, error) {
	synoClient := http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := synoClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	return res.Cookies()[0], getErr
}

func isCached(id int, cachePath string) bool {
	_, err := os.Stat(path.Join(cachePath, fmt.Sprintf("%d.png", id)))
	return err == nil
}

// func downloadPhoto(baseUrl string, albumCode string, id int, cachePath string, cookie *http.Cookie) (os.File, error) {
// 	synoClient := http.Client{}

// 	req, err := http.NewRequest(http.MethodPost, baseUrl, nil)

// 	req.AddCookie(cookie)
// 	req.Body
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	res, getErr := synoClient.Do(req)
// 	if getErr != nil {
// 		log.Fatal(getErr)
// 	}

// 	return nil, nil
// }
