package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func getInfoDivs(parentNode *html.Node) []*html.Node {
	infosParentDiv := scrape.FindAllNested(parentNode, scrape.ByClass("pure-u-2-3"))
	if len(infosParentDiv) == 0 {
		log.Error("Malformed html when looking for infos divs")
		return nil
	}

	notPure23Matcher := func(n *html.Node) bool {
		return n.DataAtom == atom.Div && scrape.Attr(n, "class") != "pure-u-2-3"
	}

	divs := scrape.FindAll(infosParentDiv[0], notPure23Matcher)

	return divs
}

func getTitle(infoDiv *html.Node) string {
	titleTag, ok := scrape.Find(infoDiv, scrape.ByTag(atom.B))
	if !ok {
		log.Error("Malformed html when looking for title")
		return ""
	}

	return scrape.Text(titleTag)
}

func getArtist(infoDiv *html.Node) (string, string) {
	artistTag, ok := scrape.Find(infoDiv, scrape.ByTag(atom.A))
	if !ok {
		log.Error("Malformed html when looking for artist")
		return "", ""
	}

	return scrape.Text(artistTag), scrape.Attr(artistTag, "href")
}

func getGenres(infoDiv *html.Node) (genres []string) {
	genreTags := scrape.FindAll(infoDiv, scrape.ByTag(atom.Span))
	for _, genreTag := range genreTags {
		genres = append(genres, scrape.Text(genreTag))
	}

	return
}

func parseInfos(parentNode *html.Node, track *audio) {
	divs := getInfoDivs(parentNode)

	if divs != nil {
		track.Title = getTitle(divs[3])
		track.Artist, track.TrackURL = getArtist(divs[4])
		track.Genres = getGenres(divs[6])
	}

	return
}

func getCoverURL(node *html.Node) string {
	coverParentDiv, ok := scrape.Find(node, scrape.ByClass("pure-u-1-3"))
	if !ok {
		log.Error("Malformed html when looking for cover URL")
		return ""
	}

	coverTag, ok := scrape.Find(coverParentDiv, scrape.ByTag(atom.Img))
	if !ok {
		log.Error("Malformed html when looking for cover URL")
		return ""
	}
	return scrape.Attr(coverTag, "src")
}

func getDownloadURL(node *html.Node) string {
	downloadURLParent, ok := scrape.Find(node, scrape.ByClass("sm2-playlist-bd"))
	if !ok {
		log.Error("Malformed html when looking for download URL")
		return ""
	}

	downloadURLTag, ok := scrape.Find(downloadURLParent, scrape.ByTag(atom.A))
	if !ok {
		log.Error("Malformed html when looking for download URL")
		return ""
	}
	return scrape.Attr(downloadURLTag, "href")
}

func getDate(div *html.Node) time.Time {
	date := scrape.Text(div)
	if date != "" {
		dateTime, err := time.Parse(timeLayout, date)
		if err == nil {
			return dateTime
		}
	}

	log.Error("Malformed html when looking for date")
	return time.Time{}
}

func getRating(span *html.Node) float32 {
	rating := scrape.Text(span)
	if rating != "" {
		rating = strings.Replace(strings.TrimSpace(strings.Split(rating, "/")[0]), ",", ".", -1)
		ratingFloat, err := strconv.ParseFloat(rating, 32)
		if err == nil {
			return float32(ratingFloat)
		}
	}

	log.Error("Malformed html when looking for rating")
	return 0.0
}

func getDownloadsCount(span *html.Node) int {
	downloadsCount := scrape.Text(span)
	if downloadsCount != "" {
		downloadsCount = strings.Replace(downloadsCount, " ", "", -1)
		downloads, err := strconv.Atoi(downloadsCount)
		if err == nil {
			return downloads
		}
	}

	log.Error("Malformed html when looking for downloads")
	return 0
}

func getPlays(span *html.Node) int {
	playsCount := scrape.Text(span)
	if playsCount != "" {
		playsCount = strings.Replace(playsCount, " ", "", -1)
		plays, err := strconv.Atoi(playsCount)
		if err == nil {
			return plays
		}
	}

	log.Error("Malformed html when looking for plays")
	return 0
}

func getLicense(span *html.Node) string {
	licenseTag, ok := scrape.Find(span, scrape.ByTag(atom.A))
	if !ok {
		log.Error("Malformed html when looking for license")
		return ""
	}

	return strings.Split(scrape.Attr(licenseTag, "href"), "license=")[1]
}

func parseAdditionalInfos(node *html.Node, track *audio) {
	additionalInfosParent, ok := scrape.Find(node, scrape.ByClass("legenddata"))
	if !ok {
		log.Error("Malformed html when looking for additional infos")
		return
	}

	additionalInfosSpans := scrape.FindAll(additionalInfosParent, scrape.ByTag(atom.Span))
	if len(additionalInfosSpans) != 5 {
		log.Error("Malformed html when looking for additional infos")
		return
	}

	track.Date = getDate(additionalInfosSpans[0])
	track.Rating = getRating(additionalInfosSpans[1])
	track.Downloads = getDownloadsCount(additionalInfosSpans[2])
	track.Plays = getPlays(additionalInfosSpans[3])
	track.License = getLicense(additionalInfosSpans[4])

	return
}

func parseAudioData(node *html.Node) *audio {
	track := &audio{}

	parseInfos(node, track)

	track.CoverArtURL = getCoverURL(node)

	if node.Parent != nil {
		track.DownloadURL = getDownloadURL(node.Parent)
		parseAdditionalInfos(node.Parent, track)
	}

	return track
}

func getPage(url string) (*html.Node, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body := resp.Body
	defer body.Close()

	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	return root, nil
}

func getAudioDivs(root *html.Node) []*html.Node {
	matcher := func(n *html.Node) bool {
		if n.DataAtom == atom.Div && n.Parent != nil {
			return strings.Contains(scrape.Attr(n, "class"), "audio-wrapper")
		}
		return false
	}

	return scrape.FindAllNested(root, matcher)
}

func scrapePage(url string) (tracks []*audio, err error) {
	root, err := getPage(url)
	if err != nil {
		return tracks, err
	}

	audioWrappers := getAudioDivs(root)

	for _, wrapper := range audioWrappers {
		track := parseAudioData(wrapper)
		// only send the track if we have at least the title
		if track.Title != "" {
			tracks = append(tracks, track)
		}
	}

	return tracks, nil
}
