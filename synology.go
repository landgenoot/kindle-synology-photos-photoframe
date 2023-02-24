package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

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

func getSynoRequest(method string, url string, payload string, cookie *http.Cookie, albumCode string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, strings.NewReader(payload))
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
