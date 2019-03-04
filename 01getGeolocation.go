package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
//	"os"
//	"strconv"
//	"sync"

	"github.com/oschwald/geoip2-golang"
)

//type ID [32]byte

type myNode struct {
	Id          string          `json:"id"`
	Ip          string          `json:"ip"`
	Port        int             `json:"port"`
	Reachable   bool            `json:"reachable"`
	Connections map[string]bool `json:"connections"`
}

type Record struct {
	Id          string          `json:"id"`
	Ip          string          `json:"ip"`
	Port        int             `json:"port"`
	Reachable   bool            `json:"reachable"`
	Country     string          `json:"country"`
	Sub         string          `json:"subdivision"`
	City        string          `json:"city"`
	ASO         string          `json:"aso"`
	Connections map[string]bool `json:"connections"`
}

type C struct {
	IDs int `json:"ids"`
	IPs int `json:"ips"`
}

type Counts struct {
	Countries map[string]C `json:"countries"`
	AS        map[string]int `json:"as"`
}



//reads all files from ./snapshots and the GeoLite databases
//iterates over all snapshots and collects geolocational data for each node
//saves a map of the nodes with geo data to ./graphs/s-TIMESTAMP_geo.json
//saves statistics about the distribution of countries and AS to ./graphs/s-TIMESTAMP_geo_counts.json
func main() {
	//get all files from the directory
	files, err := ioutil.ReadDir("./snapshots")
	if err != nil {
		log.Fatal(err)
		return
	}
	
	//load the DBs
	cityDB, err := geoip2.Open("./GeoLite/GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal(err)
		return
	}
	
	asnDB, err := geoip2.Open("./GeoLite/GeoLite2-ASN.mmdb")
	if err != nil {
		log.Fatal(err)
		return
	}
	
	//iterate over all files
	for _, file := range files {
		fname := file.Name()
		
		readMaps(fname, asnDB, cityDB)
	}
	
	cityDB.Close()
	asnDB.Close()
	log.Printf("Ende, len: %d", len(files))
}


func readMaps(fname string, asnDB *geoip2.Reader, cityDB *geoip2.Reader) {
	log.Printf("Start reading from file %s", fname)
	var thisM = make(map[int]myNode)
	var rmap = make(map[int]Record)
	var counts = new(Counts)

	counts.Countries = make(map[string]C)
	counts.AS = make(map[string]int)
	
	//read the given file into thisM
	raw, err := ioutil.ReadFile("./snapshots/" + fname)
	if err != nil {
		log.Printf("Error opening file %s", fname)
		return
	}
	err = json.Unmarshal(raw, &thisM)
	if err != nil {
		log.Printf("Error unmarshalling file %s", fname)
		return
	}
	
	//iterate over thisM (all nodes)
	for k, v := range thisM {
		ip := net.ParseIP(v.Ip)
		
		//get information from DBs
		asn, err := asnDB.ASN(ip)
		if err != nil {
			log.Fatal(err)
		}
		
		city, err := cityDB.City(ip)
		if err != nil {
			log.Fatal(err)
		}
		
		sub := ""
		if len(city.Subdivisions) > 0 {
			sub = city.Subdivisions[0].Names["en"]
		}
		country := city.Country.Names["en"]
		as := asn.AutonomousSystemOrganization
		
		record := new(Record)
		
		*record = Record {
			Id: v.Id,
			Ip: v.Ip,
			Port: v.Port,
			Reachable: v.Reachable,
			Country: country,
			Sub: sub,
			City: city.City.Names["en"],
			ASO: as,
			Connections: v.Connections,
		}
		
		//increase the count for this AS
		counts.AS[as]++
		
		rmap[k] = *record
	}
	
	//iterate over all nodes (with information -> rmap)
	//populate maps(countryname->[]ID/[]IP) with IDs/IPs, if they are not yet within these arrays
	var cIDs = make(map[string][]string)
	var cIPs = make(map[string][]string)
	for _, v := range rmap {
		if !isStringInSlice(v.Id, cIDs[v.Country]) {
			cIDs[v.Country] = append(cIDs[v.Country], v.Id)
		}
		if !isStringInSlice(v.Ip, cIPs[v.Country]) {
			cIPs[v.Country] = append(cIPs[v.Country], v.Ip)
		}
	}
	
	//iterate over the array to get the count of different IDs/IPs per country
	var thisC = new(C)
	for c, a := range cIDs {
		thisC.IDs = len(a)
		thisC.IPs = len(cIPs[c])
//		log.Printf("  Country: %s, IDs: %d, IPs: %d", c, len(a), len(cIPs[c]))
		counts.Countries[c] = *thisC
	}
	
	// format json
	mapJson, err := json.MarshalIndent(rmap, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	
	// save json file
	err = ioutil.WriteFile("./geo/" + fname, mapJson, 0644)
	if err != nil {
		log.Fatal(err)
	}
	
	countJson, err := json.MarshalIndent(counts, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	
	err = ioutil.WriteFile("./nodeInfo/" + fname[:len(fname)-5] + "_geo_counts.json", countJson, 0644)
	if err != nil {
		log.Fatal(err)
	}
	
	log.Printf("Finished writing to files from original file %s", fname)
	
}


func isStringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}







