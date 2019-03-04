package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"
)

//type ID [32]byte

type myNode struct {
	Id          string          `json:"id"`
	Ip          string          `json:"ip"`
	Port        int             `json:"port"`
	Reachable   bool            `json:"reachable"`
	Connections map[string]bool `json:"connections"`
}

type Reachability struct {
	Sessions      float32              `json:"sessionlength"`
	InterSessions float32              `json:"intersession"`
	Count         [2]int               `json:"count"`
	Reachable     map[time.Time]string `json:"reachable"`
}


//reads all files from ./snapshots/ and analyzes them
//iterates over the files and extracts when and how often a node was reachable
//creates a struct Reachability for each node id containing the reachability (map), count of online/offline times and calculates the session and intersession lengths
//saves the map containing the structs to analyzedNodes.json
func main() {
	//global nodes map, nodeID->node
	var nodes = make(map[string]myNode)
	//nodeID->Reachability
	var a = make(map[string]Reachability)
	
	//get all files from the directory
	files, err := ioutil.ReadDir("./snapshots")
	if err != nil {
		log.Fatal(err)
		return
	}
	
	//iterate over all files
	for _, file := range files {
		fname := file.Name()
		log.Print("fileName: ", fname)
		//int->node
		var thisM = make(map[int]myNode)
		//nodeID->int
		var thisMi = make(map[string]int)

		//read the given file into thisM
		raw, err := ioutil.ReadFile("./snapshots/" + fname)
		if err != nil {
			log.Fatal(err)
			return
		}
		err = json.Unmarshal(raw, &thisM)
		if err != nil {
			log.Fatal(err)
			return
		}
		
		timestamp, err := time.Parse("2006-01-02--15-04-05", fname[2:len(fname)-5])
		if err != nil {
			log.Fatal(err)
			return
		}
		
		//iterate over thisM (all nodes)
		for k, v := range thisM {
			//add this node to global nodes map
			nodes[v.Id] = v
			//contruct nodeID->int map
			thisMi[v.Id] = k
		}
		
		//iterate over all nodes found yet
		for _, v := range nodes {
			var thisReach = new(Reachability)
			thisReach.Reachable = make(map[time.Time]string)
			//if this node was already seen
			if _, ok := a[v.Id]; ok {
				//get its Reachability
				*thisReach = a[v.Id]
			}
			//if this node is in the latest nodes map
			if _, ok := thisMi[v.Id]; ok {
				//it was reachable
				thisReach.Reachable[timestamp] = "true"
				//increase counter
				thisReach.Count[0]++
			} else {
				//it was not reachable
				thisReach.Reachable[timestamp] = "false"
				//increase counter
				thisReach.Count[1]++
			}
			//write Reachability to map
			a[v.Id] = *thisReach
		}
	}

	//iterate over all Reachabilities
	for k, v := range a {
		thisSession := 0
		sessionCount := 0
		var sessions = []int{}

		thisInterSession := 0
		interSessionCount := 0
		var interSessions = []int{}

		lastEntry := "false"
		
		//iterate over all files
		for _, file := range files {
			timestamp, err := time.Parse("2006-01-02--15-04-05", file.Name()[2:len(file.Name())-5])
			if err != nil {
				log.Fatal(err)
				return
			}
			//if the timestamp is not in the map for this node
			if _, ok := v.Reachable[timestamp]; !ok {
				//count it as intersession
				if thisInterSession == 0 {
					interSessionCount++
				}
				thisInterSession++
				continue
			}
			// if the node was reachable at this time
			if v.Reachable[timestamp] == "true" {
				// if the current session hasn't started
				if lastEntry == "false" {
					// count it as a new session
					sessionCount++
					interSessions = append(interSessions, thisInterSession)
					thisInterSession = 0
				}
				// and increase the session length
				thisSession++
			} else {
				if lastEntry == "true" {
					interSessionCount++
					sessions = append(sessions, thisSession)
					thisSession = 0
				}
				thisInterSession++
			}
			lastEntry = v.Reachable[timestamp]
		}
		if lastEntry == "false" {
			interSessions = append(interSessions, thisInterSession)
		} else {
			sessions = append(sessions, thisSession)
		}
		
		var sessionLength float32
		var interSessionLength float32

		sessionLength = 0.0
		interSessionLength = 0.0

		//compute session lengths
		if sessionCount > 0 {
			for _, i := range sessions {
				sessionLength += float32(i)
			}
			sessionLength /= float32(sessionCount)
		}
		//compute intersession lengths
		if interSessionCount > 0 {
			for _, j := range interSessions {
				interSessionLength += float32(j)
			}
			interSessionLength /= float32(interSessionCount)
		}
		
		thisReach := new(Reachability)
		*thisReach = Reachability {
			Sessions: sessionLength,
			InterSessions: interSessionLength,
			Count: v.Count,
			Reachable: v.Reachable,
		}
		a[k] = *thisReach
	}

	ajson, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		log.Fatal(err)
		return
	}
	err = ioutil.WriteFile("./nodeInfo/sessionInfo.json", ajson, 0644)
	if err != nil {
		log.Fatal(err)
		return
	}

//	njson, _ := json.MarshalIndent(n, "", "\t")
//	_ = ioutil.WriteFile("mapi.json", njson, 0644)

	log.Printf("Ende, len: %d", len(a))
}
