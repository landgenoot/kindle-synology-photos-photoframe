package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

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

func getSynoAlbumRequest(baseUrl string, cookie *http.Cookie, albumCode string) (*http.Request, error) {
	requestUrl := fmt.Sprintf(`%v/webapi/entry.cgi?`, baseUrl)
	method := "POST"
	payload := `offset=0&limit=1000&api="SYNO.Foto.Browse.Item"&method="list"&version=1`
	req, err := http.NewRequest(method, requestUrl, strings.NewReader(payload))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("x-syno-sharing", albumCode)
	req.AddCookie(cookie)
	return req, nil
}

func getSynoPhotoRequest(baseUrl string, cookie *http.Cookie, albumCode string, id int) (*http.Request, error) {
	requestUrl := fmt.Sprintf(`%v/webapi/entry.cgi/20210807_144336.jpg`, baseUrl)
	method := "GET"
	payload := fmt.Sprintf(`id=%v&cache_key="35336_1628372812"&type="unit"&size="xl"&passphrase="%[2]v"&api="SYNO.Foto.Thumbnail"&method="get"&version=1&_sharing_id="%[2]v"`, id, albumCode)
	req, err := http.NewRequest(method, requestUrl, strings.NewReader(payload))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("x-syno-sharing", albumCode)
	req.AddCookie(cookie)
	return req, nil
}

func parseSynoPhotoBrowseItem(response io.Reader) (synoFotoBrowseItem, error) {
	body, err := ioutil.ReadAll(response)
	if err != nil {
		fmt.Println(err)
		return synoFotoBrowseItem{}, err
	}
	album := synoFotoBrowseItem{}
	err = json.Unmarshal(body, &album)
	return album, err
}

func parseShareLink(shareLink string) (string, string) {
	shareLinkUrl, _ := url.Parse(shareLink)
	path := strings.Split(shareLinkUrl.Path, "/")
	shareLinkUrl.Path = strings.Join(path[1:3], "/")
	albumCode := path[3]
	return shareLinkUrl.String(), albumCode
}

func fetchSynoAlbum(baseUrl string, cookie *http.Cookie, albumCode string) (synoFotoBrowseItem, error) {
	req, err := getSynoAlbumRequest(baseUrl, cookie, albumCode)
	if err != nil {
		fmt.Println(err)
		return synoFotoBrowseItem{}, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return synoFotoBrowseItem{}, err
	}
	defer res.Body.Close()
	album, jsonErr := parseSynoPhotoBrowseItem(res.Body)
	return album, jsonErr
}
