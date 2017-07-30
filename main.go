package main

import (
	"encoding/json"
	"errors"
	"flag"
	"net/http"
	"strconv"
	"strings"
	"time"

	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	gocache "github.com/patrickmn/go-cache"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const (
	baseURL    = "http://www.auboutdufil.com/index.php?"
	timeLayout = "02/01/2006"
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
	ID          uint8     `json:"-"`
}

type requestOptions struct {
	genre   string
	license string
	mood    string
	sorting string
}

type request struct {
	URL     string
	options *requestOptions
	query   string
	page    int
	hash    string
}

func (r *request) getHash() string {
	if r.hash == "" {
		r.hash += r.URL
		r.hash += r.options.genre
		r.hash += r.options.mood
		r.hash += r.options.license
		r.hash += r.options.sorting
		r.hash += strconv.Itoa(int(r.page))
	}
	return r.hash
}

func getInfoDivs(parentNode *html.Node) ([]*html.Node, error) {
	infosParentDiv := scrape.FindAllNested(parentNode, scrape.ByClass("pure-u-2-3"))
	if len(infosParentDiv) == 0 {
		return nil, errors.New("Malformed html")
	}

	notPure23Matcher := func(n *html.Node) bool {
		return n.DataAtom == atom.Div && scrape.Attr(n, "class") != "pure-u-2-3"
	}

	divs := scrape.FindAll(infosParentDiv[0], notPure23Matcher)
	if len(divs) != 10 {
		return nil, errors.New("Malformed html")
	}

	return divs, nil
}

func parseInfos(parentNode *html.Node, track audio) (audio, error) {
	divs, err := getInfoDivs(parentNode)
	if err != nil {
		return track, err
	}

	// Parse title infos
	titleTag, ok := scrape.Find(divs[3], scrape.ByTag(atom.B))
	if !ok {
		return track, errors.New("Malformed html")
	}
	track.Title = scrape.Text(titleTag)

	// Parse artist name and url
	artistTagParent, ok := scrape.Find(divs[4], scrape.ByTag(atom.Strong))
	if !ok {
		return track, errors.New("Malformed html")
	}
	artistTag, ok := scrape.Find(artistTagParent, scrape.ByTag(atom.A))
	if !ok {
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
		return track, errors.New("Malformed html")
	}

	notPure13Matcher := func(n *html.Node) bool {
		return n.DataAtom == atom.Div && scrape.Attr(n, "class") != "pure-u-1-3"
	}

	divs := scrape.FindAll(coverParentDiv[0], notPure13Matcher)
	if len(divs) != 6 {
		return track, errors.New("Malformed html")
	}

	coverTag := scrape.FindAllNested(divs[5], scrape.ByTag(atom.Img))
	if len(coverTag) != 1 {
		return track, errors.New("Malformed html")
	}
	track.CoverArtURL = scrape.Attr(coverTag[0], "src")

	// download url
	mp3PlayerDiv, ok := scrape.Find(node.Parent, scrape.ByClass("mp3player"))
	if !ok {
		return track, errors.New("Malformed html")
	}
	downloadURLParent := scrape.FindAllNested(mp3PlayerDiv, scrape.ByClass("sm2-playlist-bd"))
	if len(downloadURLParent) != 1 {
		return track, errors.New("Malformed html")
	}
	downloadURLTag := scrape.FindAllNested(downloadURLParent[0], scrape.ByTag(atom.A))
	if len(downloadURLTag) != 1 {
		return track, errors.New("Malformed html")
	}
	track.DownloadURL = scrape.Attr(downloadURLTag[0], "href")

	// additional infos
	additionalInfosParent, ok := scrape.Find(node.Parent, scrape.ByClass("legenddata"))
	if !ok {
		return track, errors.New("Malformed html")
	}

	additionalInfosSpans := scrape.FindAll(additionalInfosParent, scrape.ByTag(atom.Span))
	if len(additionalInfosSpans) != 5 {
		return track, errors.New("Malformed html")
	}

	date := scrape.Text(additionalInfosSpans[0])
	if date != "" {
		track.Date, err = time.Parse(timeLayout, date)
	}
	if date == "" || err != nil {
		return track, errors.New("Malformed html")
	}

	rating := scrape.Text(additionalInfosSpans[1])
	if rating != "" {
		rating = strings.Replace(strings.TrimSpace(strings.Split(rating, "/")[0]), ",", ".", -1)
		ratingFloat, err := strconv.ParseFloat(rating, 32)
		if err == nil {
			track.Rating = float32(ratingFloat)
		}
	}
	if rating == "" || err != nil {
		return track, errors.New("Malformed html")
	}

	downloadsCount := scrape.Text(additionalInfosSpans[2])
	if downloadsCount != "" {
		downloadsCount = strings.Replace(downloadsCount, " ", "", -1)
		track.Downloads, err = strconv.Atoi(downloadsCount)
	}
	if downloadsCount == "" || err != nil {
		return track, errors.New("Malformed html")
	}

	playsCount := scrape.Text(additionalInfosSpans[3])
	if playsCount != "" {
		playsCount = strings.Replace(playsCount, " ", "", -1)
		track.Plays, err = strconv.Atoi(playsCount)
	}
	if playsCount == "" || err != nil {
		return track, errors.New("Malformed html")
	}

	licenseTag, ok := scrape.Find(additionalInfosSpans[4], scrape.ByTag(atom.A))
	if !ok {
		return track, errors.New("Malformed html")
	}
	track.License = strings.Split(scrape.Attr(licenseTag, "href"), "license=")[1]

	return track, nil
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

func scrapePage(url string) (tracks []audio) {
	root, err := getPage(url)
	if err != nil {
		return
	}

	audioWrappers := getAudioDivs(root)

	for _, wrapper := range audioWrappers {
		track, err := parseAudioData(wrapper)
		if err != nil {
			continue
		}

		tracks = append(tracks, track)
	}

	return tracks
}

func handleRequestOptions(query url.Values, opts *requestOptions) {
	if opts.genre == "" {
		opts.genre = query.Get("genre")
	}

	if opts.license == "" {
		opts.license = query.Get("license")
	}

	if opts.mood == "" {
		opts.mood = query.Get("mood")
	}

	return
}

func getRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (req *request) {
	req = &request{}
	req.options = &requestOptions{}
	req.URL = r.URL.Path
	queryParams := r.URL.Query()

	page := queryParams.Get("page")
	if page != "" {
		req.page, _ = strconv.Atoi(page)
	} else {
		req.page = 1
	}

	if strings.HasPrefix(req.URL, "/latest") {
		req.options.sorting = "posted"
	} else if strings.HasPrefix(req.URL, "/best") {
		req.options.sorting = "note"
	} else if strings.HasPrefix(req.URL, "/downloads") {
		req.options.sorting = "countweb"
	} else if strings.HasPrefix(req.URL, "/plays") {
		req.options.sorting = "countfla"
	}

	path := ps.ByName("path")
	if path == "" || path == "/" {
		return
	}

	paths := strings.Split(strings.TrimPrefix(path, "/"), "/")
	pathsLen := len(paths)
	if pathsLen != 0 && pathsLen != 2 && pathsLen != 4 && pathsLen != 6 {
		http.Error(w, "Unsupported operation", http.StatusBadRequest)
		return
	}

	for i := 0; i < pathsLen; i = i + 2 {
		p := paths[i]

		switch p {
		case "license":
			req.options.license = paths[i+1]
			break
		case "mood":
			req.options.mood = paths[i+1]
			break
		case "genre":
			req.options.genre = paths[i+1]
			break
		default:
			http.Error(w, "Unsupported operation", http.StatusBadRequest)
			break
		}
	}

	handleRequestOptions(queryParams, req.options)

	return
}

func scrapData(r *request) (musics []audio) {
	u := baseURL + "sort=" + r.options.sorting
	if r.options.license != "" {
		u += "&license=" + r.options.license
	}
	if r.options.mood != "" {
		u += "&mood=" + r.options.mood
	}
	if r.options.genre != "" {
		u += "&tag=" + r.options.genre
	}
	u += "&page=" + strconv.Itoa(r.page)

	log.WithField("url", u).Info("Scraping page...")
	musics = scrapePage(u)

	return
}

func handleRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req := getRequest(w, r, ps)

	tracks, found := cache.Get(req.getHash())
	if !found {
		log.Info("Cache expired, scraping data...")
		scrapeTracks := scrapData(req)
		cache.Set(req.getHash(), scrapeTracks, 0)
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

func redirectHomepage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, "https://github.com/Shywim/auboutdufil-api", http.StatusTemporaryRedirect)
}

func serve(port string) {
	server := httprouter.New()
	server.GET("/", redirectHomepage)
	server.GET("/latest/*path", handleRequest)
	server.GET("/best/*path", handleRequest)
	server.GET("/downloads/*path", handleRequest)
	server.GET("/plays/*path", handleRequest)

	log.WithFields(log.Fields{
		"port": port,
	}).Info("Starting HTTP Server")

	go http.ListenAndServe(":"+port, server)

}

func main() {
	var (
		port = flag.String("p", "14000", "Port used for server")
	)
	flag.Parse()

	serve(*port)
}
