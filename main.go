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

func getOptionsFromPath(path string, opt *requestOptions) error {
	paths := strings.Split(strings.TrimPrefix(path, "/"), "/")
	pathsLen := len(paths)
	if pathsLen != 0 && pathsLen != 2 && pathsLen != 4 && pathsLen != 6 {
		return errors.New("Unsupported operation")
	}

	for i := 0; i < pathsLen; i = i + 2 {
		p := paths[i]

		switch p {
		case "license":
			opt.license = paths[i+1]
			break
		case "mood":
			opt.mood = paths[i+1]
			break
		case "genre":
			opt.genre = paths[i+1]
			break
		default:
			return errors.New("Unsupported operation")
		}
	}

	return nil
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

	err := getOptionsFromPath(path, req.options)
	if err != nil {
		http.Error(w, "Unsupported operation", http.StatusBadRequest)
		return
	}

	handleRequestOptions(queryParams, req.options)

	return
}

func scrapData(r *request) (musics []*audio, err error) {
	// build url according to options
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

	musics, err = scrapePage(u)

	return
}

func handleRequest(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	req := getRequest(w, r, ps)

	tracks, found := cache.Get(req.getHash())
	if !found {
		// cache expired, scrapping data
		scrapeTracks, err := scrapData(req)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		// store new data in cache
		cache.Set(req.getHash(), scrapeTracks, 0)
		tracks = scrapeTracks
	}

	body, _ := json.Marshal(tracks)

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// redirect to github when hitting '/'
func redirectHomepage(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.Redirect(w, r, "https://github.com/Shywim/auboutdufil-api", http.StatusTemporaryRedirect)
}

func serve(port string) {
	server := httprouter.New()
	server.GET("/", redirectHomepage)
	// we define explicitly the 4 routes so we have 'path' in params and let the router
	// handles 404s
	server.GET("/latest/*path", handleRequest)
	server.GET("/best/*path", handleRequest)
	server.GET("/downloads/*path", handleRequest)
	server.GET("/plays/*path", handleRequest)

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

	serve(*port)
}
