package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"golang.org/x/net/html"
)

func TestRequestHash(t *testing.T) {
	reqA := &request{
		URL:  "/latest/licence/cc-by/mood/sad/genre/acoustique",
		page: 1,
		options: &requestOptions{
			genre:   "acoustique",
			license: "cc-by",
			mood:    "sad",
			sorting: "latest",
		},
	}
	reqB := &request{
		URL:  "/latest/licence/cc-by/mood/sad/genre/acoustique",
		page: 1,
		options: &requestOptions{
			genre:   "acoustique",
			license: "cc-by",
			mood:    "sad",
			sorting: "latest",
		},
	}

	hashA := reqA.getHash()
	hashB := reqB.getHash()

	if hashA != hashB {
		t.Error("Expected hashes to be equals")
	}
}

func TestGetPage(t *testing.T) {
	_, err := getPage(baseURL + "sort=latest")

	if err != nil {
		t.Error("Expected no error, got ", err)
	}
}

func TestGetAudioDivs(t *testing.T) {
	root, err := getPage(baseURL + "sort=latest")

	if err != nil {
		t.Error("Dependency not met")
		return
	}

	audioDivs := getAudioDivs(root)

	if len(audioDivs) != 6 {
		t.Error("Expected to find 6 audio divs, got", len(audioDivs))
	}
}

func TestGetInfoDivs(t *testing.T) {
	root, err := getPage(baseURL + "sort=latest")

	if err != nil {
		t.Error("Dependency not met")
		return
	}

	audioDivs := getAudioDivs(root)

	if len(audioDivs) == 0 {
		t.Error("Dependency not met")
	}

	_, err = getInfoDivs(audioDivs[0])
	if err != nil {
		t.Error(err)
	}
}

func checkHasMusic(t *testing.T, resp *http.Response) {
	if resp.StatusCode != http.StatusOK {
		t.Error("Expected status", http.StatusOK, "got", resp.StatusCode)
		return
	}

	jsonStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	var data []map[string]interface{}
	err = json.Unmarshal(jsonStr, &data)
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	if len(data) <= 0 {
		t.Error("Expected to have 1 audio or more, got", len(data))
	}
}

func TestServeLatest(t *testing.T) {
	serve("1234")

	resp, err := http.Get("http://localhost:1234/latest/license/cc-byncnd/genre/indie/mood/calm")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	checkHasMusic(t, resp)
}

func TestServeBest(t *testing.T) {
	serve("1234")

	resp, err := http.Get("http://localhost:1234/best/license/cc-byncnd/genre/indie?page=1&mood=calm")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	checkHasMusic(t, resp)
}

func TestServeDownloads(t *testing.T) {
	serve("1234")

	resp, err := http.Get("http://localhost:1234/downloads/license/cc-byncnd?page=1&mood=calm&genre=indie")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	checkHasMusic(t, resp)
}

func TestServePlays(t *testing.T) {
	serve("1234")

	resp, err := http.Get("http://localhost:1234/plays?page=1&mood=calm&genre=indie&license=cc-byncnd")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	checkHasMusic(t, resp)
}

func TestHomeRedirect(t *testing.T) {
	serve("1234")

	client := &http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get("http://localhost:1234")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	if resp.StatusCode != http.StatusTemporaryRedirect {
		t.Error("Unexpected status code, expected", http.StatusTemporaryRedirect, "got", resp.StatusCode)
	}
}

func TestMain(t *testing.T) {
	main()

	resp, err := http.Get("http://localhost:14000/latest")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		t.Error("Unexpected status code, expected", http.StatusOK, "got", resp.StatusCode)
	}
}

/* Test proper errors */

func TestPathError(t *testing.T) {
	serve("1234")

	resp, err := http.Get("http://localhost:1234/garbage")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Error("Unexpected status code, expected", http.StatusNotFound, "got", resp.StatusCode)
	}
}

func TestWrongParamNumber(t *testing.T) {
	serve("1234")

	resp, err := http.Get("http://localhost:1234/latest/1/2/3")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Unexpected status code, expected", http.StatusBadRequest, "got", resp.StatusCode)
	}
}

func TestUnknownParam(t *testing.T) {
	serve("1234")

	resp, err := http.Get("http://localhost:1234/latest/garbage/cc-by")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Unexpected status code, expected", http.StatusBadRequest, "got", resp.StatusCode)
	}
}

func TestGetTitleFail(t *testing.T) {
	s := getTitle(&html.Node{})

	if s != "" {
		t.Error("Expected empty result, got", s)
	}
}

func TestGetArtistFail(t *testing.T) {
	a, u := getArtist(&html.Node{})

	if a != "" {
		t.Error("Expected empty result, got", a)
	}
	if u != "" {
		t.Error("Expected empty result, got", u)
	}
}

func TestGetGenresFail(t *testing.T) {
	s := getGenres(&html.Node{})

	if len(s) > 0 {
		t.Error("Expected empty result, got", len(s))
	}
}

func TestGetInfoDivsFail(t *testing.T) {
	_, err := getInfoDivs(&html.Node{})

	if err == nil {
		t.Error("Expected to have error, got no error")
	}
}

func TestGetParseInfos(t *testing.T) {
	_, err := parseInfos(&html.Node{}, audio{})

	if err == nil {
		t.Error("Expected to have error, got no error")
	}
}

func TestScrapePageFail(t *testing.T) {
	_, err := scrapePage("http://garbage.garbage")

	if err == nil {
		t.Error("Expected to have error, got no error")
	}
}
