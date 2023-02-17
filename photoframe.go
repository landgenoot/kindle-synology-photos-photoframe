package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
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

func fetchSynoAlbum(url string, cookie *http.Cookie) ([]int, error) {
	synoClient := http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)

	req.AddCookie(cookie)

	if err != nil {
		log.Fatal(err)
	}

	res, getErr := synoClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
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
