package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
)

type Servers struct {
	serverList []Server
	weights    []int
	quantity   int
	nextRR     int
	roundRobin bool
}

func newServers(roundR bool) (servers *Servers) {
	servers = new(Servers)
	servers.quantity = 0
	servers.roundRobin = roundR
	return
}

func (servers *Servers) findServer(url string) (int, bool) {
	for i, s := range servers.serverList {
		if s.Url() == url {
			return i, true
		}
	}
	return -1, false
}

func (servers *Servers) removeServer(url string) {
	i, _ := servers.findServer(url)
	LogInfo("Remove server index: " + strconv.Itoa(i))
	if i >= 0 {
		servers.serverList = append(servers.serverList[:i], servers.serverList[i+1:]...)
		servers.quantity--
	}
	return
}

func (servers *Servers) configServer(w http.ResponseWriter, r *http.Request) {
	server := Server{}
	jsn, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal("Error reading the body", err)
	}
	err = json.Unmarshal(jsn, &server)
	if err != nil {
		log.Fatal("Decoding error: ", err)
	}
	_, found := servers.findServer(server.Url())
	if !found {
		servers.serverList = append(servers.serverList, server)
		servers.quantity++
		servers.weights = append(servers.weights, 0)
		LogInfo("Add server " + server.Name + " new quantity: " + strconv.Itoa(servers.quantity))
	}
	return
}

func (servers *Servers) nextRoundRobin() (*Server, bool) {

	if servers.quantity == 0 {
		return &Server{}, false
	}

	index := (servers.nextRR) % servers.quantity
	servers.nextRR = (index + 1) % servers.quantity

	return &servers.serverList[index], true
}

func (servers *Servers) nextRandom() (*Server, bool) {

	if servers.quantity == 0 {
		return &Server{}, false
	}

	index := rand.Intn(servers.quantity)

	return &servers.serverList[index], true
}

func (servers *Servers) chooseServer() (*Server, bool) {

	if servers.roundRobin {
		return servers.nextRoundRobin()
	} else {
		return servers.nextRandom()
	}
}
