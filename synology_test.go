package main

import (
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

func TestGetSynoRequest(t *testing.T) {
	method := "POST"
	url := "https://www.example.com"
	payload := "a=1&b=2"
	albumCode := "k5SnJvlVW"
	cookie := http.Cookie{
		Name:  "sharing_sid",
		Value: "_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx",
		Path:  "/",
		Raw:   "sharing_sid=_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx; path=/",
	}
	want := map[string][]string{
		"Cookie":         {"sharing_sid=_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx"},
		"X-Syno-Sharing": {"k5SnJvlVW"},
	}
	got, err := getSynoRequest(method, url, payload, &cookie, albumCode)
	if !hasRequestHeaders(want, got.Header) {
		t.Fatalf(`getSynoRequest() = %v, %v, want match for %#v, nil`, got.Header, err, want)
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
