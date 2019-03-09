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
	InterSessions float32              `json:"intersessionlength"`
	SCount        int                  `json:"sessioncount"`
	ISCount       int                  `json:"intersessioncount"`
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
	}
	
	//iterate over all files
	for _, file := range files {
		fname := file.Name()
		log.Print("Reading file: ", fname)
		//int->node
		var thisM = make(map[int]myNode)
		//nodeID->int
		var thisMi = make(map[string]int)

		//read the given file into thisM
		raw, err := ioutil.ReadFile("./snapshots/" + fname)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(raw, &thisM)
		if err != nil {
			log.Fatal(err)
		}
		
		timestamp, err := time.Parse("2006-01-02--15-04-05", fname[2:len(fname)-5])
		if err != nil {
			log.Fatal(err)
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
			} else {
				//it was not reachable
				thisReach.Reachable[timestamp] = "false"
			}
			//write Reachability to map
			a[v.Id] = *thisReach
		}
	}
	
	//iterate over all Reachabilities
	for k, v := range a {
		
		var thisReach = new(Reachability)
		*thisReach = v
		count := [2]int{0, 0}
		
		//for all time entries
		for _, file := range files {
			timestamp, err := time.Parse("2006-01-02--15-04-05", file.Name()[2:len(file.Name())-5])
			if err != nil {
				log.Fatal(err)
			}
			
			//if the entry does not exist
			if _, ok := thisReach.Reachable[timestamp]; !ok {
				//add it as offline
				thisReach.Reachable[timestamp] = "false"
			}
			if thisReach.Reachable[timestamp] == "true" {
				//increase counter
				count[0]++
			} else {
				//increase counter
				count[1]++
			}
		}
		
		*thisReach = Reachability {
			Sessions: float32(0),
			InterSessions: float32(0),
			SCount: 0,
			ISCount: 0,
			Count: count,
			Reachable: thisReach.Reachable,
		}
		
		a[k] = *thisReach
	}

	//iterate over all Reachabilities
	for k, v := range a {
		thisSession := 0
		sessionCount := 0
		var sessions = []int{}

		thisInterSession := 0
		interSessionCount := 0
		var interSessions = []int{}

		//lastEntry can hav thre different states: "true", "false" and ""
		lastEntry := ""
		
		//iterate over all files
		for _, file := range files {
			timestamp, err := time.Parse("2006-01-02--15-04-05", file.Name()[2:len(file.Name())-5])
			if err != nil {
				log.Fatal(err)
			}
			
			//if the node was reachable at this time
			if v.Reachable[timestamp] == "true" {
				//if the current session hasn't started
				if lastEntry != "true" {
					//count it as a new session
					sessionCount++
				}
				//if the last entry was "false"
				if lastEntry == "false" {
					//end this intersession by appending its length to the array
					interSessions = append(interSessions, thisInterSession)
					thisInterSession = 0
				}
				//and increase the session length
				thisSession++
			} else {
				if lastEntry != "false" {
					//count it as a new intersession
					interSessionCount++
				}
				if lastEntry == "true" {
					//end this session by appending its length to the array
					sessions = append(sessions, thisSession)
					thisSession = 0
				}
				//and increase the intersession length
				thisInterSession++
			}
			lastEntry = v.Reachable[timestamp]
		}
		if lastEntry == "true" {
			sessions = append(sessions, thisSession)
		} else {
			interSessions = append(interSessions, thisInterSession)
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
			SCount: sessionCount,
			ISCount: interSessionCount,
			Count: v.Count,
			Reachable: v.Reachable,
		}
		a[k] = *thisReach
	}

	ajson, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("./nodeInfo/sessionInfo.json", ajson, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Ende, len: %d", len(a))
}
