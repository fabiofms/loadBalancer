package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Proxy struct {
	Host   string
	Port   int
	Scheme string
}

func (proxy Proxy) origin() string {
	return (proxy.Scheme + "://" + proxy.Host + ":" + strconv.Itoa(proxy.Port))
}

func (proxy *Proxy) chooseServer() (*Server, bool) {

	return servers.chooseServer()
}

func (proxy *Proxy) removeServer(failServer string) {

	servers.removeServer(failServer)
	return
}

func (proxy Proxy) ReverseProxy(w http.ResponseWriter, r *http.Request, server Server) (int, error) {
	u, err := url.Parse(server.Url() + r.RequestURI)
	if err != nil {
		LogErrAndCrash(err.Error())
	}

	r.URL = u
	r.Header.Set("X-Forwarded-Host", r.Host)
	r.Header.Set("Origin", proxy.origin())
	r.Host = server.Url()
	r.RequestURI = ""

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(r)
	if err != nil {
		LogErr("connection refused")
		return 0, err
	}
	LogInfo("Recieved response: " + strconv.Itoa(resp.StatusCode))

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		LogErr("Proxy: Failed to read response body")
		http.NotFound(w, r)
		return 0, err
	}

	buffer := bytes.NewBuffer(bodyBytes)
	for k, v := range resp.Header {
		w.Header().Set(k, strings.Join(v, ";"))
	}

	w.WriteHeader(resp.StatusCode)

	io.Copy(w, buffer)
	return resp.StatusCode, nil
}

func (proxy *Proxy) attemptServers(w http.ResponseWriter, r *http.Request) {

	server, found := proxy.chooseServer()
	if !found {
		LogErr("Failed to find server for request")
		http.NotFound(w, r)
		return
	}
	LogInfo("Got request: " + r.RequestURI)
	LogInfo("Sending to server: " + server.Name)

	server.Connections += 1
	_, err := proxy.ReverseProxy(w, r, *server)
	server.Connections -= 1

	if err != nil {
		LogWarn("Server did not respond: " + server.Name)
		proxy.removeServer(server.Url())
		proxy.attemptServers(w, r)
		return
	}

	LogInfo("Responded to request successfuly")
}

func (proxy *Proxy) handler(w http.ResponseWriter, r *http.Request) {
	proxy.attemptServers(w, r)
}
