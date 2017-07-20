package main

import (
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	gocache "github.com/patrickmn/go-cache"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	baseURL = "http://www.auboutdufil.com"
)

var (
	cache = gocache.New(1*time.Hour, 1*time.Minute)
)

type audio struct {
	Title       string    `json:"title"`
	Artist      string    `json:"artist"`
	TrackURL    string    `json:"track_url"`
	Genres      []string  `json:"genres"`
	CoverArtURL string    `json:"cover_art_url"`
	DownloadURL string    `json:"download_url"`
	License     string    `json:"license"`
	Downloads   int       `json:"downloads"`
	Plays       int       `json:"play_count"`
	Rating      float32   `json:"rating"`
	Date        time.Time `json:"published_date"`
}

func parseInfos(parentNode *html.Node, track audio) (audio, error) {

	infosParentDiv := scrape.FindAllNested(parentNode, scrape.ByClass("pure-u-2-3"))
	if len(infosParentDiv) == 0 {
		log.Warn("Incorrect html data, layout may have changed")
		return track, errors.New("Malformed html")
	}

	notPure23Matcher := func(n *html.Node) bool {
		return n.DataAtom == atom.Div && scrape.Attr(n, "class") != "pure-u-2-3"
	}

	divs := scrape.FindAll(infosParentDiv[0], notPure23Matcher)
	if len(divs) != 10 {
		log.WithFields(log.Fields{
			"divsNumber": len(divs),
			"expected":   "10",
		}).Warn("Incorrect html data, layout may have changed")
		return track, errors.New("Malformed html")
	}

	// Parse title infos
	titleTag, ok := scrape.Find(divs[3], scrape.ByTag(atom.B))
	if !ok {
		log.Warn("Incorrect html data while searching for title, layout may have changed")
		return track, errors.New("Malformed html")
	}
	track.Title = scrape.Text(titleTag)

	// Parse artist name and url
	artistTagParent, ok := scrape.Find(divs[4], scrape.ByTag(atom.Strong))
	if !ok {
		log.Warn("Incorrect html data while searching for artist, layout may have changed")
		return track, errors.New("Malformed html")
	}
	artistTag, ok := scrape.Find(artistTagParent, scrape.ByTag(atom.A))
	if !ok {
		log.Warn("Incorrect html data while searching for artist, layout may have changed")
		return track, errors.New("Malformed html")
	}
	track.Artist = scrape.Text(artistTag)
	track.TrackURL = scrape.Attr(artistTag, "href")

	// Parse genres
	genreTags := scrape.FindAll(divs[6], scrape.ByTag(atom.Span))
	for _, genreTag := range genreTags {
		track.Genres = append(track.Genres, scrape.Text(genreTag))
	}

	return track, nil
}

func parseAudioData(node *html.Node) (track audio, err error) {

	track, err = parseInfos(node, track)
	if err != nil {
		return track, err
	}

	// look for cover image
	coverParentDiv := scrape.FindAllNested(node, scrape.ByClass("pure-u-1-3"))
	if len(coverParentDiv) == 0 {
		log.Warn("Incorrect html data while searching for cover url, layout may have changed")
		return track, errors.New("Malformed html")
	}

	notPure13Matcher := func(n *html.Node) bool {
		return n.DataAtom == atom.Div && scrape.Attr(n, "class") != "pure-u-1-3"
	}

	divs := scrape.FindAll(coverParentDiv[0], notPure13Matcher)
	if len(divs) != 6 {
		log.WithFields(log.Fields{
			"divsNumber": len(divs),
			"expected":   "6",
		}).Warn("Incorrect html data while searching for cover url, layout may have changed")
		return track, errors.New("Malformed html")
	}

	coverTag := scrape.FindAllNested(divs[5], scrape.ByTag(atom.Img))
	if len(coverTag) != 1 {
		log.Warn("Incorrect html data while searching for cover url, layout may have changed")
		return track, errors.New("Malformed html")
	}
	track.CoverArtURL = scrape.Attr(coverTag[0], "src")

	// download url
	mp3PlayerDiv, ok := scrape.Find(node.Parent, scrape.ByClass("mp3player"))
	if !ok {
		log.Warn("Incorrect html data while searching for download url, layout may have changed")
		return track, errors.New("Malformed html")
	}
	downloadURLParent := scrape.FindAllNested(mp3PlayerDiv, scrape.ByClass("sm2-playlist-bd"))
	if len(downloadURLParent) != 1 {
		log.Warn("Incorrect html data while searching for download url, layout may have changed")
		return track, errors.New("Malformed html")
	}
	downloadURLTag := scrape.FindAllNested(downloadURLParent[0], scrape.ByTag(atom.A))
	if len(downloadURLTag) != 1 {
		log.Warn("Incorrect html data while searching for download url, layout may have changed")
		return track, errors.New("Malformed html")
	}
	track.DownloadURL = scrape.Attr(downloadURLTag[0], "href")

	// additional infos
	additionalInfosParent, ok := scrape.Find(node.Parent, scrape.ByClass("legenddata"))
	if !ok {
		log.Warn("Incorrect html data, layout may have changed")
		return track, errors.New("Malformed html")
	}

	additionalInfosSpans := scrape.FindAll(additionalInfosParent, scrape.ByTag(atom.Span))
	if len(additionalInfosSpans) != 5 {
		log.Warn("Incorrect html data while searching for additional infos, layout may have changed")
		return track, errors.New("Malformed html")
	}

	licenseTag, ok := scrape.Find(additionalInfosSpans[4], scrape.ByTag(atom.A))
	if !ok {
		log.Warn("Incorrect html data while searching for license infos, layout may have changed")
		return track, errors.New("Malformed html")
	}
	track.License = strings.Split(scrape.Attr(licenseTag, "href"), "license=")[1]

	return track, nil
}

func scrapePage(url string) (tracks []audio) {
	resp, err := http.Get(url)

	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
			"err": err,
		}).Error("Failed to get page")
		return
	}

	body := resp.Body
	defer body.Close()

	root, err := html.Parse(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"url": url,
		}).Error("Unable to parse this web page")
		return
	}

	matcher := func(n *html.Node) bool {
		if n.DataAtom == atom.Div && n.Parent != nil {
			return strings.Contains(scrape.Attr(n, "class"), "audio-wrapper")
		}
		return false
	}

	audioWrappers := scrape.FindAllNested(root, matcher)
	for _, wrapper := range audioWrappers {
		track, err := parseAudioData(wrapper)
		if err != nil {
			continue
		}

		tracks = append(tracks, track)
	}

	return tracks
}

func handleLatest(w http.ResponseWriter, r *http.Request) {
	tracks, found := cache.Get("latest")
	if !found {
		log.Info("Cache expired, scraping data...")
		scrapeTracks := scrapePage(baseURL)
		scrapeTracks = append(scrapeTracks, scrapePage(baseURL+"/index.php?page=2")...)
		scrapeTracks = append(scrapeTracks, scrapePage(baseURL+"/index.php?page=3")...)
		cache.Set("latest", scrapeTracks, 0)
		tracks = scrapeTracks
	}

	body, err := json.Marshal(tracks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func server(port string) {
	server := http.NewServeMux()
	server.HandleFunc("/latest", handleLatest)

	log.WithFields(log.Fields{
		"port": port,
	}).Info("Starting HTTP Server")

	http.ListenAndServe(":"+port, server)

}

func main() {
	var (
		port = flag.String("p", "14000", "Port used for server")
	)
	flag.Parse()

	server(*port)
}
