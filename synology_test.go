package main

import (
	"bytes"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestGetSharingSidCookie(t *testing.T) {
	response := "{}"
	headers := map[string]string{
		"Set-Cookie": "sharing_sid=_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx; path=/",
	}
	server := mockServer(200, response, headers, nil, "")
	defer server.s.Close()
	want := http.Cookie{
		Name:  "sharing_sid",
		Value: "_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx",
		Path:  "/",
		Raw:   "sharing_sid=_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx; path=/",
	}
	mockUrl, _ := url.Parse(server.s.URL)
	cookie, err := getSharingSidCookie(mockUrl)

	if !reflect.DeepEqual(*cookie, want) || err != nil {
		t.Fatalf(`fetchSynoAlbum() = %v, %v, want match for %#v, nil`, *cookie, err, want)
	}
}

func TestGetSynoAlbumRequest(t *testing.T) {
	baseUrl := "https://www.example.com"
	albumCode := "k5SnJvlVW"
	cookie := http.Cookie{
		Name:  "sharing_sid",
		Value: "_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx",
		Path:  "/",
		Raw:   "sharing_sid=_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx; path=/",
	}
	wantHeader := map[string][]string{
		"Cookie":         {"sharing_sid=_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx"},
		"X-Syno-Sharing": {"k5SnJvlVW"},
	}
	wantUrl := "https://www.example.com/webapi/entry.cgi?"
	wantMethod := "POST"
	wantPayload := `offset=0&limit=1000&api="SYNO.Foto.Browse.Item"&method="list"&version=1`

	got, err := getSynoAlbumRequest(baseUrl, &cookie, albumCode)
	buf := new(bytes.Buffer)
	buf.ReadFrom(got.Body)
	gotPayload := buf.String()

	if !hasRequestHeaders(wantHeader, got.Header) {
		t.Fatalf(`getSynoAlbumRequest() = %v, %v, want match for %#v, nil`, got.Header, err, wantHeader)
	}
	if wantUrl != got.URL.String() {
		t.Fatalf(`getSynoAlbumRequest().Method = %v, %v, want match for %#v, nil`, got.Method, err, wantMethod)
	}
	if wantMethod != got.Method {
		t.Fatalf(`getSynoAlbumRequest().Method = %v, %v, want match for %#v, nil`, got.Method, err, wantMethod)
	}
	if wantPayload != gotPayload {
		t.Fatalf(`getSynoAlbumRequest().Body = %v, %v, want match for %#v, nil`, gotPayload, err, wantPayload)
	}
}

func TestGetSynoPhotoRequest(t *testing.T) {
	baseUrl := "https://www.example.com"
	albumCode := "k5SnJvlVW"
	id := 15052
	cookie := http.Cookie{
		Name:  "sharing_sid",
		Value: "_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx",
		Path:  "/",
		Raw:   "sharing_sid=_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx; path=/",
	}
	wantHeader := map[string][]string{
		"Cookie": {"sharing_sid=_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx"},
	}
	wantUrl := `https://www.example.com/webapi/entry.cgi/20210807_144336.jpg?_sharing_id=k5SnJvlVW&api=SYNO.Foto.Thumbnail&cache_key=35336_1628372812&id=15052&method=get&passphrase=k5SnJvlVW&size=xl&type=unit&version=1`
	wantMethod := "GET"

	got, err := getSynoPhotoRequest(baseUrl, &cookie, albumCode, id)

	if !hasRequestHeaders(wantHeader, got.Header) {
		t.Fatalf(`getSynoPhotoRequest() = %v, %v, want match for %#v, nil`, got.Header, err, wantHeader)
	}
	if wantUrl != got.URL.String() {
		t.Fatalf(`getSynoAlbumRequest().Method = %v, %v, want match for %#v, nil`, got.URL.String(), err, wantUrl)
	}
	if wantMethod != got.Method {
		t.Fatalf(`getSynoPhotoRequest().Method = %v, %v, want match for %#v, nil`, got.Method, err, wantMethod)
	}
}

func TestParseSynoPhotoBrowseItem(t *testing.T) {
	json := `{
		"success":true,
		"data":{
			"list":[
				{"id":15052}, 
				{"id":10401}
			]
		}
	}`
	want := synoFotoBrowseItem{
		Success: true,
		Data: Data{
			List: []Photo{
				{
					Id: 15052,
				},
				{
					Id: 10401,
				},
			},
		},
	}
	got, err := parseSynoPhotoBrowseItem(strings.NewReader(json))
	if !reflect.DeepEqual(got, want) || err != nil {
		t.Fatalf(`parseSynoPhotoBrowseItem() = %v, %v, want match for %#v, nil`, got, err, want)
	}
}

func TestParseShareLink(t *testing.T) {
	shareLink, _ := url.Parse("https://b92.dsmdemo.synologydemo.com:5001/mo/sharing/k5SnJvlVW")
	wantBaseUrl := "https://b92.dsmdemo.synologydemo.com:5001/mo/sharing"
	wantAlbumCode := "k5SnJvlVW"
	gotBaseUrl, gotAlbumCode := parseShareLink(shareLink)
	if gotBaseUrl != wantBaseUrl || gotAlbumCode != wantAlbumCode {
		t.Fatalf(`parseShareLink() = %v, %v, want match for %#v,  %#v`, gotBaseUrl, gotAlbumCode, wantBaseUrl, wantAlbumCode)
	}
}
