package main

import (
	"testing"

	"golang.org/x/net/html"
)

func TestGetPage(t *testing.T) {
	_, err := getPage(baseURL + "sort=posted")

	if err != nil {
		t.Error("Expected no error, got ", err)
	}
}

func TestGetAudioDivs(t *testing.T) {
	root, err := getPage(baseURL + "sort=posted")

	if err != nil {
		t.Error("Dependency not met")
		return
	}

	audioDivs := getAudioDivs(root)

	if len(audioDivs) != 12 {
		t.Error("Expected to find 12 audio divs, got", len(audioDivs))
	}
}

func TestGetInfoDivs(t *testing.T) {
	root, err := getPage(baseURL + "sort=posted")

	if err != nil {
		t.Error("Dependency not met")
		return
	}

	audioDivs := getAudioDivs(root)

	if len(audioDivs) == 0 {
		t.Error("Dependency not met")
	}

	divs := getInfoDivs(audioDivs[0])
	if divs == nil {
		t.Error("Expected to have divs, got nil")
	}
}

func TestParseAudioDataFail(t *testing.T) {
	track := parseAudioData(&html.Node{})

	if track.Title != "" {
		t.Error("Expected empty track, got", track.Title)
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
	divs := getInfoDivs(&html.Node{})

	if divs != nil {
		t.Error("Expected to have nil divs")
	}
}

func TestGetParseInfos(t *testing.T) {
	track := &audio{}
	parseInfos(&html.Node{}, track)

	if track.Title != "" {
		t.Error("Expected to have empty title, got", track.Title)
	}
}

func TestParseAdditionalInfosFail(t *testing.T) {
	track := &audio{}
	parseAdditionalInfos(&html.Node{}, track)

	if track.License != "" {
		t.Error("Expected to have empty infos, got", track.License)
	}
}

func TestScrapePageFail(t *testing.T) {
	_, err := scrapePage("http://garbage.garbage")

	if err == nil {
		t.Error("Expected to have error, got no error")
	}
}
