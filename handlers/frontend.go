package handlers

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"github.com/juho05/h-bank"
	"github.com/juho05/h-bank/config"
)

type FrontendHandler struct {
	fs    http.FileSystem
	proxy *httputil.ReverseProxy
}

func NewFrontendHandler() *FrontendHandler {
	if !hbank.DevFrontendEnabled {
		return &FrontendHandler{fs: http.FS(hbank.FrontendFS)}
	}
	uri, err := url.Parse(config.Data.DevFrontend)
	if err != nil {
		log.Fatal(err)
	}
	return &FrontendHandler{
		proxy: httputil.NewSingleHostReverseProxy(uri),
	}
}

func (f *FrontendHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if f.proxy != nil {
		f.proxy.ServeHTTP(w, r)
		return
	}
	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
	}
	upath = path.Clean(upath)

	var file http.File
	var err error
	file, err = f.fs.Open(upath)
	if err != nil {
		file, err = f.fs.Open(upath + ".html")
		if err != nil {
			file, err = f.fs.Open("index.html")
			if err != nil {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}
		}
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if info.IsDir() {
		file, err = f.fs.Open(path.Join(strings.TrimPrefix(upath, "/"), "index.html"))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		defer file.Close()
	}

	http.ServeContent(w, r, upath, info.ModTime(), file)
}
