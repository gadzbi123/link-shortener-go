package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"text/template"
	"time"

	"github.com/spf13/viper"
)

const (
// LinksInMemory        int           = 10
// LinksCleanerSchedule time.Duration = 5 * time.Minute
// port                 string        = ":9090"
)

var (
	linksStore = NewLinksStore()
	config     = Config{}
)

type Config struct {
	Port                 string        `yaml:"port"`
	LinksCleanerSchedule time.Duration `yaml:"linksCleanerSchedule"`
	LinksInMemory        int           `yaml:"linksInMemory"`
}

type PageError struct{ err string }

func (le PageError) Error() string {
	return fmt.Sprintf("link error: %s", le.err)
}

func getErrorPage(w http.ResponseWriter, r *http.Request, pageError PageError) {
	var tmpl = template.Must(template.New("error.html").ParseFiles("templates/error.html"))

	var err error
	if err = tmpl.Execute(w, pageError); err != nil {
		e := fmt.Errorf("error template crashed: %w", err)
		panic(e)
	}
	slog.Debug("Error page presented", "err", pageError)
}

func GetRedirectLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		slog.Warn("wrong method on get endpoint", "method", r.Method)
		getErrorPage(w, r, PageError{"wrong method on request used"})
		return
	}
	idStr := r.PathValue("id")
	if idStr == "" {
		slog.Warn("empty id")
		getErrorPage(w, r, PageError{"id was empty"})
		return
	}

	redirectLink, found := linksStore.FindRedirect(idStr)
	if !found {
		slog.Warn("no redirect link found", "shortLink", idStr)
		getErrorPage(w, r, PageError{"no redirect link found"})
		return
	}

	slog.Info("Redirecting to", "redirectLink", redirectLink, "shortLink", idStr)
	http.Redirect(w, r, redirectLink, http.StatusMovedPermanently)
}

func SaveRedirectLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		slog.Warn("wrong method on post endpoint", "method", r.Method)
		getErrorPage(w, r, PageError{"wrong method used"})
		return
	}

	err := r.ParseForm()
	if err != nil {
		slog.Warn("failed to parse body", "err", err)
		getErrorPage(w, r, PageError{"failed to parse form"})
		return
	}

	link := NewLink()
	link.RedirectUrl = r.PostFormValue("redirect-url")
	if link.RedirectUrl == "" {
		slog.Warn("provided empty link")
		getErrorPage(w, r, PageError{"provided empty link"})
		return
	}

	l, err := checkLinkAvailability(context.Background(), link.RedirectUrl)
	if err != nil {
		getErrorPage(w, r, PageError{err.Error()})
		return
	}
	link.RedirectUrl = l

	// check if link is defined
	shortUrl, found := linksStore.FindShortUrl(link.RedirectUrl)
	if found {
		link.ShortUrl = shortUrl
	} else {
		// if not then make random link
		link.ShortUrl, err = getRandomShortLink()
		if err != nil {
			slog.Warn("got bad link created", "err", err)
			getErrorPage(w, r, PageError{"failed to create link"})
			return
		}
		linksStore.Add(link)
	}
	fullLink := r.URL.Scheme + r.Host + "/" + link.ShortUrl
	tmpl := template.Must(template.New("result.html").ParseFiles("templates/result.html"))
	err = tmpl.Execute(w, fullLink)
	if err != nil {
		err := fmt.Errorf("result template crash: %w", err)
		getErrorPage(w, r, PageError{"failed to parse result"})
		panic(err)
	}
	slog.Info("Saved redirect link", "fullLink", fullLink)
}

func GetIndexPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		slog.Warn("wrong method on index used")
		getErrorPage(w, r, PageError{"wrong method request used"})
	}
	var tmpl = template.Must(template.New("home.html").ParseFiles("./templates/home.html"))
	var err = tmpl.Execute(w, nil)
	if err != nil {
		err := fmt.Errorf("home template crashed: %w", err)
		getErrorPage(w, r, PageError{"server internal error"})
		panic(err)
	}
}

func loadViper(fileName string) {
	viper.AddConfigPath("configs")
	viper.SetConfigName(fileName)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	if err := viper.UnmarshalExact(&config); err != nil {
		panic(err)
	}
	slog.Debug("viper has loaded config")
}

func run() {
	devEnv := flag.Bool("dev", false,
		"specify if dev env should be used instead of prod")
	flag.Parse()
	if devEnv != nil && *devEnv {
		loadViper("dev")
	} else {
		loadViper("prod")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/{$}", GetIndexPage)
	mux.HandleFunc("/{id}", GetRedirectLink)
	mux.HandleFunc("/result", SaveRedirectLink)

	go linksCleaner()

	slog.Info("Listening on", "port", config.Port)
	if err := http.ListenAndServe(config.Port, mux); err != nil {
		slog.Error("Error occurred on listen", "err", err)
		return
	}
	// index page where you put link - DONE
	// error page when something goes wrong - DONE
	// links stay in the memory for X min and then get shrinked to X items - DONE
	// config in viper - DONE
	// write tests for current implementation - DONE

	// Save records to DB on post
	// Don't save when not unique

	// Load 10 recent records from db on boot
	// Load 10 recent records from db on cleanup

	// improve look of front page

	// parse links from a file uploaded

	// links stored in sqlite {shortUrl, redirectUrl, creation time}
	// links are stored in db every time they are created
	// links stay in db for 30 days
	// 10 most recent links are stored in memory at boot
	// mechanism to clean links every minute - DONE

	// save to clipboard button

	// before shrink all records are stored in db
	// db checks if record exists, if not then stores the info
	// links in db never expire

	// EXTEND:
	// - new tab with text saving
	// - zip the values

}

func main() {
	run()
}
