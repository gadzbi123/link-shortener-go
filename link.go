package main

import (
	"log/slog"
	"sync"
	"time"
)

type Link struct {
	ShortUrl    string    `json:"shortUrl,omitempty"`
	RedirectUrl string    `json:"redirectUrl"`
	Created     time.Time `json:"created,omitempty"`
}

func NewLink() *Link {
	return &Link{Created: time.Now()}
}

type LinksStore struct {
	links []*Link
	mtx   sync.Mutex
}

func NewLinksStore() *LinksStore {
	return &LinksStore{links: make([]*Link, 0, config.LinksInMemory)}
}

func (ls *LinksStore) FindShortUrl(redirectUrl string) (shortUrl string, found bool) {
	ls.mtx.Lock()
	defer ls.mtx.Unlock()
	for _, l := range ls.links {
		if l.RedirectUrl == redirectUrl {
			shortUrl = l.ShortUrl
			found = true
			return
		}
	}
	return
}

func (ls *LinksStore) FindRedirect(id string) (redirect string, found bool) {
	linksStore.mtx.Lock()
	defer linksStore.mtx.Unlock()
	for _, l := range linksStore.links {
		if l.ShortUrl == id {
			redirect = l.RedirectUrl
			found = true
			return
		}
	}
	return
}

func (ls *LinksStore) Add(link *Link) {
	ls.mtx.Lock()
	defer ls.mtx.Unlock()
	ls.links = append(ls.links, link)
}

func linksCleaner() {
	cleaner := time.NewTicker(config.LinksCleanerSchedule)
	for {
		<-cleaner.C
		linksStore.mtx.Lock()
		linksLen := len(linksStore.links)
		slog.Info("Running cleaner", "linksLen", linksLen)
		if linksLen <= config.LinksInMemory {
			linksStore.mtx.Unlock()
			continue
		}
		tempLinks := make([]*Link, config.LinksInMemory)
		copy(tempLinks, linksStore.links[linksLen-config.LinksInMemory:])
		linksStore.links = tempLinks
		linksStore.mtx.Unlock()
		slog.Debug("Cleaned linkes", "count", linksLen-config.LinksInMemory)
	}
}
