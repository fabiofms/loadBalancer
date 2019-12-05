package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

var servers *Servers

func main() {
	LogInfo("Spinning up load balancer...")
	LogInfo("Reading Config.yml...")
	proxy, err := ReadConfig()
	if err != nil {
		LogErr("An error occurred while trying to parse config.yml")
		LogErrAndCrash(err.Error())
	}
	LogInfo("Finish Reading Config.yml...")
	i, err := strconv.Atoi(os.Args[1])
	if err != nil {
		i = 0
	}
	if i == 0 {
		servers = newServers(false)
	} else {
		servers = newServers(true)
	}

	go backendServe()
	clientServe(&proxy)
}

func clientServe(proxy *Proxy) {
	// Proxy server
	proxyMux := http.NewServeMux()
	proxyMux.HandleFunc("/", proxy.handler)
	LogInfo("Port: " + strconv.Itoa(proxy.Port))
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(proxy.Port), proxyMux))
	LogInfo("Listening to requests on port: " + strconv.Itoa(proxy.Port))
}

func backendServe() {
	// Config server
	configMux := http.NewServeMux()
	configMux.HandleFunc("/", servers.configServer)
	log.Fatal(http.ListenAndServe(":5000", configMux))
}
