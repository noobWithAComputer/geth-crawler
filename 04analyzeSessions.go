package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"time"
)

type myNode struct {
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

type Reachability struct {
	Sessions      float32              `json:"sessionlength"`
	InterSessions float32              `json:"intersessionlength"`
	SCount        int                  `json:"sessioncount"`
	ISCount       int                  `json:"intersessioncount"`
	Count         [2]int               `json:"count"`
	Reachable     map[time.Time]string `json:"reachable"`
}

type statsAll struct {
	Nodes            int         `json:"nodes"`
	AvgSessions      float32     `json:"avgsessionlength"`
	MedSessions      int         `json:"medsessionlength"`
	AvgInterSessions float32     `json:"avgintersessionlength"`
	MedInterSessions int         `json:"medintersessionlength"`
	Sessions         map[int]int `json:"sessions"`
	InterSessions    map[int]int `json:"intersessions"`
}

type stats struct {
	Nodes         int     `json:"nodes"`
	Sessions      float32 `json:"sessionlength"`
	InterSessions float32 `json:"intersessionlength"`
	SCount        float32 `json:"sessioncount"`
	ISCount       float32 `json:"intersessioncount"`
}


//reads all files from ./geo/ and analyzes them
//iterates over the files and extracts when and how often a node was reachable
//creates a struct Reachability for each node id containing the reachability (map), count of online/offline times and calculates the session and intersession lengths
//saves the map containing the structs to ./nodeInfo/sessionInfo.json
//calculates averages over all nodes, on country-level and on AS-level.
//saves the averages to ./nodeInfo/sessionInfoAverage.json, ./nodeInfo/sessionInfoC.json and ./nodeInfo/sessionInfoAS.json respectively
//country- and AS-level structs cann also be exported to .txt files, therefore just remove the last comments
func main() {
	//global nodes map, nodeID->node
	var nodes = make(map[string]myNode)
	//nodeID->Reachability
	var a = make(map[string]Reachability)
	
	//get all files from the directory
	files, err := ioutil.ReadDir("./geo")
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
		raw, err := ioutil.ReadFile("./geo/" + fname)
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
	
	log.Print("Looping over reachability map.")
	
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
	
	log.Print("Calculating session lengths.")
	
	allSessions := []int{}
	allInterSessions := []int{}

	//iterate over all Reachabilities
	for k, v := range a {
		thisSession := 0
		sessionCount := 0
		var sessions = []int{}

		thisInterSession := 0
		interSessionCount := 0
		var interSessions = []int{}

		//lastEntry can have three different states: "true", "false" and ""
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
		
		for _, s := range sessions {
			allSessions = append(allSessions, s)
		}
		for _, i := range interSessions {
			allInterSessions = append(allInterSessions, i)
		}
	}
	
	files = []os.FileInfo{}

	ajson, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("./nodeInfo/sessionInfo.json", ajson, 0644)
	if err != nil {
		log.Fatal(err)
	}
	
	log.Print("Calculating averages.")
	
	sumSessions := 0
	sumInterSessions := 0
	medSessions := 0
	medInterSessions := 0
	var histSessions = make(map[int]int)
	var histInterSessions = make(map[int]int)
	
	//sort the arrays
	sort.Ints(allSessions)
	sort.Ints(allInterSessions)
	
	//construct histograms
	for _, s := range allSessions {
		sumSessions += s
		histSessions[s]++
	}
	for _, i := range allInterSessions {
		sumInterSessions += i
		histInterSessions[i]++
	}
	
	//get the medians from the sorted arrays
	medSessions = allSessions[len(allSessions)/2]
	medInterSessions = allInterSessions[len(allInterSessions)/2]
	
	var sessionStats = new(statsAll)
	
	*sessionStats = statsAll {
		Nodes: len(a),
		AvgSessions: float32(float32(sumSessions)/float32(len(allSessions))),
		MedSessions: medSessions,
		AvgInterSessions: float32(float32(sumInterSessions)/float32(len(allInterSessions))),
		MedInterSessions: medInterSessions,
		Sessions: histSessions,
		InterSessions: histInterSessions,
	}
	sjson, err := json.MarshalIndent(sessionStats, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("./nodeInfo/sessionInfoAverage.json", sjson, 0644)
	if err != nil {
		log.Fatal(err)
	}
	
	histSessions = make(map[int]int)
	histInterSessions = make(map[int]int)
	
	log.Print("Calculating averages for countries.")
	
	countryStats := make(map[string]stats)
	
	//iterate over all reachability entries
	for id, reach := range a {
		//initialize variables
		country := nodes[id].Country
		theseStats := new(stats)
		node := 1
		sessions := reach.Sessions
		interSessions := reach.InterSessions
		sCount := float32(reach.SCount)
		iSCount := float32(reach.ISCount)
		
		//overwrite them in case the entry of this country already exists
		if _, ok := countryStats[country]; ok {
			*theseStats = countryStats[country]
			node = theseStats.Nodes + 1
			sessions = theseStats.Sessions + reach.Sessions
			interSessions = theseStats.InterSessions + reach.InterSessions
			sCount = theseStats.SCount + float32(reach.SCount)
			iSCount = theseStats.ISCount + float32(reach.ISCount)
		}
		
		*theseStats = stats {
			Nodes: node,
			Sessions: sessions,
			InterSessions: interSessions,
			SCount: sCount,
			ISCount: iSCount,
		}
		
		countryStats[country] = *theseStats
	}
	
	for c, s := range countryStats {
		theseStats := new(stats)
		*theseStats = s
		
		//calculate averages
		node := theseStats.Nodes
		sessions := float32(theseStats.Sessions / float32(theseStats.Nodes))
		interSessions := float32(theseStats.InterSessions / float32(theseStats.Nodes))
		sCount := float32(theseStats.SCount / float32(theseStats.Nodes))
		iSCount := float32(theseStats.ISCount / float32(theseStats.Nodes))
		
		*theseStats = stats {
			Nodes: node,
			Sessions: sessions,
			InterSessions: interSessions,
			SCount: sCount,
			ISCount: iSCount,
		}
		
		countryStats[c] = *theseStats
	}
	
	cjson, err := json.MarshalIndent(countryStats, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("./nodeInfo/sessionInfoC.json", cjson, 0644)
	if err != nil {
		log.Fatal(err)
	}
	
	countryStats = make(map[string]stats)
	
	log.Print("Calculating averages for ASes.")
	
	asStats := make(map[string]stats)
	
	//same as for the countries
	for id, reach := range a {
		as := nodes[id].ASO
		theseStats := new(stats)		
		node := 1
		sessions := reach.Sessions
		interSessions := reach.InterSessions
		sCount := float32(reach.SCount)
		iSCount := float32(reach.ISCount)
		
		if _, ok := asStats[as]; ok {
			*theseStats = asStats[as]
			node = theseStats.Nodes + 1
			sessions = theseStats.Sessions + reach.Sessions
			interSessions = theseStats.InterSessions + reach.InterSessions
			sCount = theseStats.SCount + float32(reach.SCount)
			iSCount = theseStats.ISCount + float32(reach.ISCount)
		}
		
		*theseStats = stats {
			Nodes: node,
			Sessions: sessions,
			InterSessions: interSessions,
			SCount: sCount,
			ISCount: iSCount,
		}
		
		asStats[as] = *theseStats
	}
	
	nodes = make(map[string]myNode)
	
	a = make(map[string]Reachability)
	
	for as, s := range asStats {
		theseStats := new(stats)
		*theseStats = s
		
		node := theseStats.Nodes
		sessions := float32(theseStats.Sessions / float32(theseStats.Nodes))
		interSessions := float32(theseStats.InterSessions / float32(theseStats.Nodes))
		sCount := float32(theseStats.SCount / float32(theseStats.Nodes))
		iSCount := float32(theseStats.ISCount / float32(theseStats.Nodes))
		
		*theseStats = stats {
			Nodes: node,
			Sessions: sessions,
			InterSessions: interSessions,
			SCount: sCount,
			ISCount: iSCount,
		}
		
		asStats[as] = *theseStats
	}
	
	asjson, err := json.MarshalIndent(asStats, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("./nodeInfo/sessionInfoAS.json", asjson, 0644)
	if err != nil {
		log.Fatal(err)
	}
	
	log.Printf("Ende, len: %d", len(a))
}
