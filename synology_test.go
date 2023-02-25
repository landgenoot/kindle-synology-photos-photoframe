package main

import (
	"bytes"
	"net/http"
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
	cookie, err := getSharingSidCookie(server.s.URL)

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
	wantMethod := "POST"
	wantPayload := `offset=0&limit=1000&api="SYNO.Foto.Browse.Item"&method="list"&version=1`

	got, err := getSynoAlbumRequest(baseUrl, &cookie, albumCode)
	buf := new(bytes.Buffer)
	buf.ReadFrom(got.Body)
	gotPayload := buf.String()

	if !hasRequestHeaders(wantHeader, got.Header) {
		t.Fatalf(`getSynoRequest() = %v, %v, want match for %#v, nil`, got.Header, err, wantHeader)
	}
	if wantMethod != got.Method {
		t.Fatalf(`getSynoRequest().Method = %v, %v, want match for %#v, nil`, got.Method, err, wantMethod)
	}
	if wantPayload != gotPayload {
		t.Fatalf(`getSynoRequest().Body = %v, %v, want match for %#v, nil`, gotPayload, err, wantPayload)
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
	shareLink := "https://b92.dsmdemo.synologydemo.com:5001/mo/sharing/k5SnJvlVW"
	wantBaseUrl := "https://b92.dsmdemo.synologydemo.com:5001/mo/sharing"
	wantAlbumCode := "k5SnJvlVW"
	gotBaseUrl, gotAlbumCode := parseShareLink(shareLink)
	if gotBaseUrl != wantBaseUrl || gotAlbumCode != wantAlbumCode {
		t.Fatalf(`parseShareLink() = %v, %v, want match for %#v,  %#v`, gotBaseUrl, gotAlbumCode, wantBaseUrl, wantAlbumCode)
	}
}

// func TestDownloadPhoto(t *testing.T) {
// 	request := http.Request{

// 	}
// 	body := downloadPhoto(req, )
// }

// func test
