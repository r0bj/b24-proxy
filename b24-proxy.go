package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/elazarl/goproxy"
)

func main() {
    proxy := goproxy.NewProxyHttpServer()
    proxy.Verbose = true

	proxy.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Host == "" {
			log.Println(w, "Cannot handle requests without Host header, e.g., HTTP 1.0")
			return
		}
		req.URL.Scheme = "http"
		req.URL.Host = req.Host
		proxy.ServeHTTP(w, req)
	})

	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if req.Header.Get("X-B24-URL") == "" || req.Header.Get("X-B24-PROXY-HOST") == "" || req.Header.Get("X-B24-PROXY-AUTH") == "" {
			return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusInternalServerError, "Missing header")
		}
	
		return req, nil
	})

	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		newURL, err := url.Parse(req.Header.Get("X-B24-URL"))
		if err != nil {
			return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusInternalServerError, "X-B24-URL header parse failed")
		}

		req.URL = newURL
		req.Host = newURL.Host


		proxyURL, err := url.Parse(req.Header.Get("X-B24-PROXY-HOST"))
		if err != nil {
			return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusInternalServerError, "X-B24-PROXY-HOST header parse failed")
		}

		proxyCreds := strings.Split(req.Header.Get("X-B24-PROXY-AUTH"), ":")
		if len(proxyCreds) != 2 {
			return req, goproxy.NewResponse(req, goproxy.ContentTypeText, http.StatusInternalServerError, "X-B24-PROXY-AUTH header parse failed")
		}
		proxyURL.User = url.UserPassword(proxyCreds[0], proxyCreds[1])

		proxy.Tr = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
			// ProxyConnectHeader: http.Header{"Proxy-Authorization": {"Basic " + req.Header.Get("X-B24-PROXY-AUTH")}},
		}

		return req, nil
	})

	log.Fatal(http.ListenAndServe(":8080", proxy))
}
