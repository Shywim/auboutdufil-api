package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

func utilLaunchServer() {
	go serve("1234")
}

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
	utilLaunchServer()

	resp, err := http.Get("http://localhost:1234/latest/license/CC-BYNCND/genre/indie/mood/calm")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	checkHasMusic(t, resp)
}

func TestServeBest(t *testing.T) {
	utilLaunchServer()

	resp, err := http.Get("http://localhost:1234/best/license/CC-BYNCND/genre/indie?page=1&mood=calm")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	checkHasMusic(t, resp)
}

func TestServeDownloads(t *testing.T) {
	utilLaunchServer()

	resp, err := http.Get("http://localhost:1234/downloads/license/CC-BYNCND?page=1&mood=calm&genre=indie")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	checkHasMusic(t, resp)
}

func TestServePlays(t *testing.T) {
	utilLaunchServer()

	resp, err := http.Get("http://localhost:1234/plays?page=1&mood=calm&genre=indie&license=CC-BYNCND")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	checkHasMusic(t, resp)
}

func TestHomeRedirect(t *testing.T) {
	utilLaunchServer()

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
	go main()

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
	utilLaunchServer()

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
	utilLaunchServer()

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
	utilLaunchServer()

	resp, err := http.Get("http://localhost:1234/latest/garbage/cc-by")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Error("Unexpected status code, expected", http.StatusBadRequest, "got", resp.StatusCode)
	}
}
