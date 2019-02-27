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


// reads all files from ./snapshots/ and analyzes them
// iterates over the files and extracts when and how often a node was reachable
// creates a struct Reachability for each node id containing the reachability (map), count of online/offline times and calculates the session and intersession lengths
// saves the map containing the structs to analyzedNodes.json
func main() {
//	var m = make(map[string]myNode)
	var nodes = make(map[string]myNode)
	var a = make(map[string]Reachability)
	files, err := ioutil.ReadDir("./snapshots")
	if err != nil {
		log.Print("err")
		return
	}
	
	for _, file := range files {
		fname := file.Name()
		log.Print("fileName: ", fname)
		var thisM = make(map[int]myNode)
		var thisMi = make(map[string]int)

		raw, err := ioutil.ReadFile("./snapshots/" + fname)
		if err != nil {
			log.Print("err1")
			return
		}
		err = json.Unmarshal(raw, &thisM)
		if err != nil {
			log.Print("err2")
			return
		}
		timestamp, err := time.Parse("2006-01-02--15-04-05", fname[2:len(fname)-5])
		if err != nil {
			log.Print("err3")
			return
		}
		
		for k, v := range thisM {
			nodes[v.Id] = v
			thisMi[v.Id] = k
		}
		
		for _, v := range nodes {
			var thisReach = new(Reachability)
			thisReach.Reachable = make(map[time.Time]string)
			if _, ok := a[v.Id]; ok {
				*thisReach = a[v.Id]
			}
			if _, ok := thisMi[v.Id]; ok {
				thisReach.Reachable[timestamp] = "true"
				thisReach.Count[0]++
			} else {
				thisReach.Reachable[timestamp] = "false"
				thisReach.Count[1]++
			}
			
			a[v.Id] = *thisReach
		}
	}

	for k, v := range a {
		thisSession := 0
		sessionCount := 0
		var sessions = []int{}

		thisInterSession := 0
		interSessionCount := 0
		var interSessions = []int{}

		lastEntry := "false"
		
		for _, file := range files {
			timestamp, err := time.Parse("2006-01-02--15-04-05", file.Name()[2:len(file.Name())-5])
			if err != nil {
				log.Print("err3")
				return
			}
			if _, ok := v.Reachable[timestamp]; !ok {
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

		if sessionCount > 0 {
			for _, i := range sessions {
				sessionLength += float32(i)
			}
			log.Printf("sessionCount=%d, sessionLength=%f", sessionCount, sessionLength)
			sessionLength /= float32(sessionCount)
		}
		if interSessionCount > 0 {
			for _, j := range interSessions {
				interSessionLength += float32(j)
			}
			log.Printf("interSessionCount= %d, interSessionLength=%f", interSessionCount, interSessionLength)
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
		log.Print("err3")
		return
	}
	err = ioutil.WriteFile("./nodeInfo/sessionInfo.json", ajson, 0644)
	if err != nil {
		log.Print("err4")
		return
	}

//	njson, _ := json.MarshalIndent(n, "", "\t")
//	_ = ioutil.WriteFile("mapi.json", njson, 0644)

	log.Printf("Ende, len: %d", len(a))
}
