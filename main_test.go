package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
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
		t.Error("Expected to find 6 audio divs, got ", len(audioDivs))
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

func TestServe(t *testing.T) {
	serve("1234")

	resp, err := http.Get("http://localhost:1234/latest/license/cc-byncnd/tag/indie/mood/calm")
	if err != nil {
		t.Error("Expected no error, got ", err)
		return
	}

	jsonStr, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error("Expected no error, got ", err)
		return
	}

	var data []map[string]interface{}
	err = json.Unmarshal(jsonStr, &data)
	if err != nil {
		t.Error("Expected no error, got ", err)
		return
	}

	if len(data) <= 0 {
		t.Error("Expected to have 1 audio or more, got ", len(data))
	}
}
