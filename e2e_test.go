package main

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func Test_e2e(t *testing.T) {
	go func() {
		os.Args = []string{"test", "--dev"}
		run()
	}()
	time.Sleep(2 * time.Second)
	client := &http.Client{}
	resp, err := client.Get("http://localhost" + config.Port)
	if err != nil {
		t.Fatalf("Response had error in index: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Main page didn't return 200")
	}
	buff := make([]byte, 1024)
	n, err := resp.Body.Read(buff)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("Failed to read body: %v", err)
	}
	buff = buff[:n]
	body := string(buff)
	hasUrl := strings.Contains(body, "Enter url:")
	hasHeader := strings.Contains(body, "<h1>URL Shortener</h1>")
	if !hasUrl || !hasHeader {
		t.Fatalf("Page was somewhat empty: url=%v, header=%v", hasUrl, hasHeader)
	}
	// Make request for new url
	bodyContent := url.Values{}
	bodyContent.Add("redirect-url", "google.com")
	body = bodyContent.Encode()
	resp, err = client.Post("http://localhost"+config.Port+"/result",
		"application/x-www-form-urlencoded",
		strings.NewReader(body))
	if err != nil {
		t.Fatalf("Response had error in result: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Status code was not 200")
	}

	var collectText func(n *html.Node, buf *bytes.Buffer)
	collectText = func(n *html.Node, buf *bytes.Buffer) {
		if n.Type == html.TextNode && strings.Contains(n.Data, "/") {
			buf.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			collectText(c, buf)
		}
	}

	nodes, _ := html.Parse(resp.Body)
	buf := &bytes.Buffer{}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			collectText(n, buf)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(nodes)
	resp, err = client.Get("http://" + buf.String())
	if err != nil {
		t.Fatalf("Failed to get redirect url: %s", buf.String())
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to redirect: statuscode: %v", resp.Status)
	}
}
