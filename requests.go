package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func checkLinkAvailability(ctx context.Context, link string) (redirectLink string, err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	msg := make(chan string)
	prefixes := []string{"", "https://", "http://"}

	for _, pf := range prefixes {
		go func() {
			c := &http.Client{}
			resp, err := c.Get(pf + link)
			if err != nil || resp.StatusCode != http.StatusOK {
				slog.Debug("not found", "link", pf+link)
				return
			}
			slog.Debug("link available", "link", pf+link, "status", resp.Status)
			select {
			case <-ctx.Done():
				return
			case msg <- pf + link:
				cancel()
			}
		}()
	}
	timeout := time.NewTicker(4 * time.Second)
	select {
	case linkRes := <-msg:
		slog.Debug("Provided link available", "link", linkRes)
		return linkRes, nil
	case <-timeout.C:
		slog.Warn("Provided link reached timeout", "link", link)
		return "", fmt.Errorf("provided link does not exist: %s", link)
	}
}
