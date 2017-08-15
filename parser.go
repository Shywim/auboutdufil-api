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
		return scrape.Attr(n, "class") != "pure-u-2-3" && (n.DataAtom == atom.Div || n.DataAtom == atom.Span)
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
		track.Title = getTitle(divs[2])
		track.Artist, track.TrackURL = getArtist(divs[3])
		track.Genres = getGenres(divs[5])
		track.Date = getDate(divs[7])
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
	dateTag, ok := scrape.Find(div, scrape.ByTag(atom.I))
	if !ok {
		log.Error("Malformed html when looking for date")
		return time.Time{}
	}

	date := scrape.Text(dateTag)
	if date != "" {
		dateTime, err := time.Parse(timeLayout, strings.TrimSpace(strings.Replace(date, "Publié le", "", -1)))
		if err == nil {
			return dateTime
		}
	}

	log.Error("Malformed html when looking for date")
	return time.Time{}
}

func getRating(span *html.Node) float32 {
	ratingTag, ok := scrape.Find(span, scrape.ByClass("teil"))
	if !ok {
		log.Error("Malformed html when looking for rating")
		return 0.0
	}

	rating := scrape.Text(ratingTag)
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
	downloadsCountTag, ok := scrape.Find(span, scrape.ByClass("fontbold"))
	if !ok {
		log.Error("Malformed html when looking for downloads")
		return 0
	}

	downloadsCount := scrape.Text(downloadsCountTag)
	if downloadsCount != "" {
		downloadsCount = strings.TrimSuffix(strings.Replace(downloadsCount, " ", "", -1), "téléchargements")
		downloads, err := strconv.Atoi(downloadsCount)
		if err == nil {
			return downloads
		}
	}

	log.Error("Malformed html when looking for downloads")
	return 0
}

func getPlays(span *html.Node) int {
	playsCountTag, ok := scrape.Find(span, scrape.ByClass("fontbold"))
	if !ok {
		log.Error("Malformed html when looking for plays")
		return 0
	}

	playsCount := scrape.Text(playsCountTag)
	if playsCount != "" {
		playsCount = strings.TrimSuffix(strings.Replace(playsCount, " ", "", -1), "écoutes")
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
	additionalInfosTabs := scrape.FindAll(node, scrape.ByClass("tab2"))
	if len(additionalInfosTabs) != 6 {
		log.Error("Malformed html when looking for additional infos")
		return
	}

	//track.Date = getDate(additionalInfosTabs[0])
	track.Plays = getPlays(additionalInfosTabs[0])
	track.Downloads = getDownloadsCount(additionalInfosTabs[1])
	track.License = getLicense(additionalInfosTabs[2])
	track.Rating = getRating(additionalInfosTabs[3])

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
	return root, err
}

func getAudioDivs(root *html.Node) []*html.Node {
	matcher := func(n *html.Node) bool {
		if n.DataAtom != atom.Div || n.Parent == nil {
			return false
		}

		if scrape.Attr(n, "class") != "box" {
			return false
		}

		pg := scrape.FindAll(n, scrape.ByClass("pure-g"))
		if len(pg) != 1 {
			return false
		}

		c := scrape.FindAll(pg[0], scrape.ByClass("pure-u-md-1-2"))
		if len(c) != 2 {
			return false
		}

		return true
	}

	return scrape.FindAllNested(root, matcher)
}

func scrapePage(url string) (tracks []*audio, err error) {
	log.WithFields(log.Fields{"url": url}).Info("Start scraping...")
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
