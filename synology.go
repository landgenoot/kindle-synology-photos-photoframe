package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
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

func getSharingSidCookie(url *url.URL) (*http.Cookie, error) {
	synoClient := http.Client{}
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
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
	requestUrl := fmt.Sprintf(`%v/webapi/entry.cgi/20210807_144336.jpg?`, baseUrl)
	method := "GET"

	u, err := url.Parse(requestUrl)
	if err != nil {
		log.Fatal(err)
	}
	q := u.Query()
	q.Set("id", strconv.Itoa(id))
	q.Set("cache_key", "35336_1628372812")
	q.Set("type", "unit")
	q.Set("size", "xl")
	q.Set("passphrase", albumCode)
	q.Set("api", "SYNO.Foto.Thumbnail")
	q.Set("method", "get")
	q.Set("version", "1")
	q.Set("_sharing_id", albumCode)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
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

func parseShareLink(shareLink *url.URL) (string, string) {
	path := strings.Split(shareLink.Path, "/")
	shareLink.Path = strings.Join(path[1:3], "/")
	albumCode := path[3]
	return shareLink.String(), albumCode
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
