package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
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

type C struct {
	IDs int `json:"ids"`
	IPs int `json:"ips"`
}

type Counts struct {
	Countries map[string]C   `json:"countries"`
	AS        map[string]int `json:"as"`
}

type timeCount struct {
	Count map[time.Time]int `json:"count"`
}


//reads all files from ./nodeInfo/ which contain "geo_count" and analyzes them
//iterates over the files and extracts how many nodes from which AS were online
//creates a map containing the number of online node per snapshot for all ASes (histogram)
//saves the map to ./nodeInfo/asTimeStats.txt
func main() {
	var ases = make(map[string]bool)
	var timeStats = make(map[string]*timeCount)
	
	//get all files from the directory
	files, err := ioutil.ReadDir("./nodeInfo")
	if err != nil {
		log.Fatal(err)
	}
	
	//iterate over all files
	for _, file := range files {
		fname := file.Name()
		if !strings.Contains(fname, "geo_counts") {
			continue
		}
		
		var counts = new(Counts)
		
		counts.Countries = make(map[string]C)
		counts.AS = make(map[string]int)
		
		raw, err := ioutil.ReadFile("./nodeInfo/" + fname)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(raw, &counts)
		if err != nil {
			log.Fatal(err)
		}
		
		for as, _ := range(counts.AS) {
			if _, ok := ases[as]; !ok {
				ases[as] = true
			}
		}
	}
	
	for _, file := range files {
		fname := file.Name()
		if !strings.Contains(fname, "geo_counts") {
			continue
		}
		log.Print("Reading file: ", fname)
		
		timestamp, err := time.Parse("2006-01-02--15-04-05", fname[2:len(fname)-16])
		if err != nil {
			log.Fatal(err)
		}
		
		var counts = new(Counts)
		
		counts.Countries = make(map[string]C)
		counts.AS = make(map[string]int)
		
		raw, err := ioutil.ReadFile("./nodeInfo/" + fname)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(raw, &counts)
		if err != nil {
			log.Fatal(err)
		}
		
		for as, _ := range(ases) {
			if _, ok := timeStats[as]; !ok {
				timeStats[as] = new(timeCount)
				timeStats[as].Count = make(map[time.Time]int)
			}
			
			if _, ok := counts.AS[as]; !ok {
				timeStats[as].Count[timestamp] = 0
			} else {
				timeStats[as].Count[timestamp] = counts.AS[as]
			}
		}
	
	}
	
	writeToFile(timeStats, "./nodeInfo/asTimeStats.txt")
	
	log.Printf("Ende, len: %d", len(timeStats))
}

func writeToFile(asStats map[string]*timeCount, fname string) {
	
	f, err := os.Create(fname)
	if err != nil {
		log.Fatal(err)
	}
	
	files, err := ioutil.ReadDir("./nodeInfo")
	if err != nil {
		log.Fatal(err)
	}
	
	for as, s := range asStats {
		f.Write([]byte(as + "$"))
		
		for _, file := range files {
			fname := file.Name()
			if !strings.Contains(fname, "geo_counts") {
				continue
			}
			
			timestamp, err := time.Parse("2006-01-02--15-04-05", fname[2:len(fname)-16])
			if err != nil {
				log.Fatal(err)
			}
			
			thisCount := strconv.Itoa(s.Count[timestamp])
			f.Write([]byte(thisCount + "$"))
		}
		f.Write([]byte("\n"))
	}
	
	err = f.Close()
	if err != nil {
		log.Fatal(err)
	}
}
