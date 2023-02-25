package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
)

type countingServer struct {
	s          *httptest.Server
	successful int
	failed     []string
}

func mockServer(code int, body string, headers map[string]string, requestHeaders map[string][]string, requestBody string) *countingServer {
	server := &countingServer{}
	server.s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.successful++
		if !hasRequestHeaders(requestHeaders, r.Header) && !hasRequestBody(requestBody, r.Body) {
			server.failed = append(server.failed, r.URL.RawQuery)
			http.Error(w, "{}", 999)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		for key, element := range headers {
			w.Header().Set(key, element)
		}
		w.WriteHeader(code)
		fmt.Fprintln(w, body)
	}))

	return server
}

func hasRequestBody(want string, got io.ReadCloser) bool {
	if want != "" && got != nil {
		var bodyBytes []byte
		bodyBytes, _ = ioutil.ReadAll(got)
		if string(bodyBytes) != want {
			return false
		}
	}
	return true
}

func hasRequestHeaders(want map[string][]string, got map[string][]string) bool {
	if want != nil {
		for key, element := range want {
			if !reflect.DeepEqual(got[key], element) {
				return false
			}
		}
	}
	return true
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func TestGetRandomPhoto(t *testing.T) {
	album := synoFotoBrowseItem{
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
	ids := []int{15052, 10401}
	photo, err := getRandomPhoto(album)
	if !contains(ids, photo.Id) || err != nil {
		t.Fatalf(`getPhoto() = %q, %v, want match for %#q, nil`, photo, err, ids)
	}
}

// func TestFetchSynoAlbum(t *testing.T) {
// 	response := `{
// 		"success":true,
// 		"data":{
// 			"list":[
// 				{
// 					"id":15052,
// 					"filename":"20160824_195115.jpg",
// 					"filesize":8393648,
// 					"time":1472068275,
// 					"indexed_time":1625238601571,
// 					"owner_user_id":2,
// 					"folder_id":1309,
// 					"type":"photo",
// 					"additional":{
// 						"resolution":{"width":2592, "height":1944},
// 						"orientation":1,
// 						"orientation_original":1,
// 						"thumbnail":{
// 							"m":"ready",
// 							"xl":"ready",
// 							"preview":"broken",
// 							"sm":"ready",
// 							"cache_key":"15052_1625238462",
// 							"unit_id":15052},
// 						"provider_user_id":2}
// 				},{
// 					"id":10401,
// 					"filename":"20160910_164128.jpg",
// 					"filesize":9655210,
// 					"time":1473525688,
// 					"indexed_time":1625237897811,
// 					"owner_user_id":2,
// 					"folder_id":1031,
// 					"type":"photo",
// 					"additional":{
// 						"resolution":{"width":4032,"height":3024},
// 						"orientation":1,
// 						"orientation_original":1,
// 						"thumbnail":{
// 							"m":"ready",
// 							"xl":"ready",
// 							"preview":"broken",
// 							"sm":"ready",
// 							"cache_key":"10401_1625238153",
// 							"unit_id":10401},
// 						"provider_user_id":2}
// 				}]}}`

// 	cookie := http.Cookie{
// 		Name:  "sharing_sid",
// 		Value: "_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx",
// 		Path:  "/",
// 		Raw:   "sharing_sid=_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx; path=/",
// 	}
// 	requestHeaders := map[string][]string{
// 		"Cookie":         {"sharing_sid=_xxxxxxxxxx_xxxxxxxxxxxxxxx_xxxx"},
// 		"X-Syno-Sharing": {"k5SnJvlVW"},
// 	}
// 	albumCode := "k5SnJvlVW"
// 	server := mockServer(200, response, nil, requestHeaders, `offset=0&limit=1000&api="SYNO.Foto.Browse.Item"&method="list"&version=1`)
// 	defer server.s.Close()
// 	want := []int{15052, 10401}
// 	ids, err := fetchSynoAlbum(server.s.URL, &cookie, albumCode)
// 	if !reflect.DeepEqual(ids, want) || err != nil {
// 		t.Fatalf(`fetchSynoAlbum() = %v, %v, want match for %#v, nil`, ids, err, want)
// 	}
// }

func TestCachedPhoto(t *testing.T) {
	tmpdir := t.TempDir()
	id := 15052
	existentFile := path.Join(tmpdir, fmt.Sprintf("%d.png", 15052))
	_, err := os.Create(existentFile)
	want := true
	got := isCached(id, tmpdir)
	if got != want {
		t.Fatalf(`isCached() = %v, %v, want match for %#v, nil`, got, err, want)
	}
}

func TestNotCachedPhoto(t *testing.T) {
	tmpdir := t.TempDir()
	id := 99999
	existentFile := path.Join(tmpdir, fmt.Sprintf("%d.png", 15052))
	_, err := os.Create(existentFile)
	want := false
	got := isCached(id, tmpdir)
	if got != want {
		t.Fatalf(`isCached() = %v, %v, want match for %#v, nil`, got, err, want)
	}
}

func TestDownloadPhoto(t *testing.T) {
	tmpdir := t.TempDir()
	name := path.Join(tmpdir, fmt.Sprintf("%d.png", 1337))
	server := mockServer(200, "{}", nil, nil, "")
	defer server.s.Close()
	req, _ := http.NewRequest("GET", server.s.URL, strings.NewReader(""))
	downloadErr := downloadPhoto(*req, name)
	_, err := os.Stat(path.Join(tmpdir, fmt.Sprintf("%d.png", 1337)))
	if err != nil || downloadErr != nil {
		t.Fatalf(`downloadPhoto() = %v, %v`, downloadErr, err)
	}
}
