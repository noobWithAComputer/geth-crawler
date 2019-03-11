package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type myNode struct {
	Id          string          `json:"id"`
	Ip          string          `json:"ip"`
	Port        int             `json:"port"`
	Reachable   bool            `json:"reachable"`
	Country     string          `json:"country"`
	Subdivision string          `json:"subdivision"`
	City        string          `json:"city"`
	ASO         string          `json:"aso"`
	Connections map[string]bool `json:"connections"`
}

type eclipse struct {
	Count int               `json:"count"`
	Nodes map[string][]node `json:"nodes"`
}

type sybil1 struct {
	Count int             `json:"count"`
	IP    map[int]*sybil2 `json:"ip"`
}

type sybil2 struct {
	Count int            `json:"count"`
	Nodes map[int][]node `json:"nodes"`
}

type node struct {
	Id      string `json:"id"`
	Ip      string `json:"ip"`
	Port    int    `json:"port"`
	Country string `json:"country"`
	City    string `json:"city"`
}


//reads all files from ./snapshots/ and analyzes them
//iterates over the files and looks for sybils and eclipse nodes
//creates a map over the IDs (eclipse) and the IPs (sybils) and collects the nodes falling into this subspace
//saves the maps containing the structs to ./nodeInfo/s-TIMESTAMP_eclipse.json and ./nodeInfo/s-TIMESTAMP_sybils.json
func main() {
	//get all files from the directory
	files, err := ioutil.ReadDir("./snapshots")
	if err != nil {
		log.Fatal(err)
	}
	
	//iterate over all files
	for _, file := range files {
		fname := file.Name()
		log.Print("fileName: ", fname)
		var thisM = make(map[int]myNode)
		var eclipses = make(map[string]*eclipse)
		var sybils = make(map[string]*sybil1)

		//read the given file into thisM
		raw, err := ioutil.ReadFile("./geo/" + fname)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(raw, &thisM)
		if err != nil {
			log.Fatal(err)
		}
		
		//iterate over thisM (all nodes)
		for _, v := range thisM {
			thisNode := new(node)
			*thisNode = node {
				Id: v.Id,
				Ip: v.Ip,
				Port: v.Port,
				Country: v.Country,
				City: v.City,
			}
			
			//get the beginning of the nodeID
			bId := v.Id[:4]
			//and the rest
			rId := v.Id[4:]
			
			//if the ID prefix of this node is not in the eclipse map, initialize it
			if _, ok := eclipses[bId]; !ok {
				eclipses[bId] = new(eclipse)
			}
			if eclipses[bId].Nodes == nil {
				eclipses[bId].Nodes = make(map[string][]node)
			}
			
			//add the eclipse
			eclipses[bId].Nodes[rId] = append(eclipses[bId].Nodes[rId], *thisNode)
			eclipses[bId].Count++
			
			//get the parts of the IP
			ip := strings.Split(v.Ip, ".")
			thisIP := ip[0] + "." + ip[1]
			midIP, _ := strconv.Atoi(ip[2])
			lastIP, _ := strconv.Atoi(ip[3])
			
			//if the IP prefix is not in the sybils map, initialize them
			if _, ok := sybils[thisIP]; !ok {
				sybils[thisIP] = new(sybil1)
			}
			if sybils[thisIP].IP == nil {
				sybils[thisIP].IP = make(map[int]*sybil2)
			}
			if _, ok := sybils[thisIP].IP[midIP]; !ok {
				sybils[thisIP].IP[midIP] = new(sybil2)
			}
			if sybils[thisIP].IP[midIP].Nodes == nil {
				sybils[thisIP].IP[midIP].Nodes = make(map[int][]node)				
			}
			
			//add the sybil
			sybils[thisIP].IP[midIP].Nodes[lastIP] = append(sybils[thisIP].IP[midIP].Nodes[lastIP], *thisNode)
		}
		
		//iterate over all eclipses
		for k, v := range eclipses {
			//if there is only one entry for this ID prefix, delete it
			if v.Count < 2 {
				delete(eclipses, k)
			}
		}
		
		//iterate over all sybils
		for k1, v1 := range sybils {
			for k2, v2 := range v1.IP {
				// if there is only one entry for this lastIP, delete it
				if len(v2.Nodes) < 2 {
					delete(sybils[k1].IP, k2)
				}
			}
			//if there is only one entry for this midIP, delete it
			if len(v1.IP) < 2 {
				delete(sybils, k1)
			}
		}

		//count the sybills per level
		for _, v1 := range sybils {
			for _, v2 := range v1.IP {
				for _, n := range v2.Nodes {
					v2.Count += len(n)
				}
				v1.Count += v2.Count
			}
		}
		

		eclipseJson, err := json.MarshalIndent(eclipses, "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile("./nodeInfo/" + fname[:len(fname)-5] + "_eclipse.json", eclipseJson, 0644)
		if err != nil {
			log.Fatal(err)
		}
		
		sybilsJson, err := json.MarshalIndent(sybils, "", "\t")
		if err != nil {
			log.Fatal(err)
		}
		
		err = ioutil.WriteFile("./nodeInfo/" + fname[:len(fname)-5] + "_sybils.json", sybilsJson, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Ende, length=%d", len(files))
}
