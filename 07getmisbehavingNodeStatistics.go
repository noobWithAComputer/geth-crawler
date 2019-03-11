package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"strconv"
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

type stats struct {
	Sybil_count16 float32             `json:"sybil_count16"`
	Min_sybils16  int                 `json:"min_sybils16"`
	Med_sybils16  int                 `json:"med_sybils16"`
	Avg_sybils16  float32             `json:"avg_sybils16"`
	Max_sybils16  int                 `json:"max_sybils16"`
	Hist_sybils16 map[int]int         `json:"hist_sybils16"`
	Sybil_count24 float32             `json:"sybil_count24"`
	Min_sybils24  int                 `json:"min_sybils24"`
	Med_sybils24  int                 `json:"med_sybils24"`
	Avg_sybils24  float32             `json:"avg_sybils24"`
	Max_sybils24  int                 `json:"max_sybils24"`
	Hist_sybils24 map[int]int         `json:"hist_sybils24"`
	Eclipse_count float32             `json:"eclipseCount"`
	Min_eclipse   int                 `json:"min_eclipse"`
	Med_eclipse   int                 `json:"med_eclipse"`
	Avg_eclipse   float32             `json:"avg_eclipse"`
	Max_eclipse   int                 `json:"max_eclipse"`
	Huge_sybils16 map[string]*sybil1  `json:"huge_sybils16"`
	Huge_sybils24 map[string]*sybil2  `json:"huge_sybils24"`
	Huge_eclipse  map[string]*eclipse `json:"huge_eclipse"`
}


//reads all files from ./nodeInfo/ which contain either "eclipse" or "sybils" and analyzes them
//iterates over the files and extracts how many eclipses/sybils occured
//creates a struct stats that contains statistics of the occurences of eclipses and sybils
//saves the struct to ./nodeInfo/misbehavingNodeStats.json
func main() {
	var Stats = new(stats)
	Stats.Huge_eclipse = make(map[string]*eclipse)
	Stats.Huge_sybils16 = make(map[string]*sybil1)
	Stats.Huge_sybils24 = make(map[string]*sybil2)
	Stats.Hist_sybils16 = make(map[int]int)
	Stats.Hist_sybils24 = make(map[int]int)
	var huge_eclipse = make(map[string]*eclipse)
	var huge_sybils16 = make(map[string]*sybil1)
	var huge_sybils24 = make(map[string]*sybil2)
	var hist_sybils16 = make(map[int]int)
	var hist_sybils24 = make(map[int]int)
	
	eclipse_count := 0
	sybil_count16 := 0
	sybil_count24 := 0
	var eclipseSizes = []int{}
	var sybilSizes16 = []int{}
	var sybilSizes24 = []int{}
	var max_eclipse_p = ""
	var max_eclipse_e eclipse
	var max_sybils_p16 = ""
	var max_sybils_s16 sybil1
	var max_sybils_p24 = ""
	var max_sybils_s24 sybil2
	
	min_eclipse := 100
	avg_eclipse := float32(0)
	med_eclipse := 0
	max_eclipse := 0
	min_sybils16 := 100
	avg_sybils16 := float32(0)
	med_sybils16 := 0
	max_sybils16 := 0
	min_sybils24 := 100
	avg_sybils24 := float32(0)
	med_sybils24 := 0
	max_sybils24 := 0
	
	//get all files from the directory
	files, err := ioutil.ReadDir("./nodeInfo")
	if err != nil {
		log.Fatal(err)
	}
	
	number_files := 0
	
	//iterate over all files
	for _, file := range files {
		fname := file.Name()
		if !strings.Contains(fname, "eclipse") {
			continue
		}
		log.Printf("Reading file: %s", fname)
		
		var eclipses = make(map[string]*eclipse)
		
		raw, err := ioutil.ReadFile("./nodeInfo/" + fname)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(raw, &eclipses)
		if err != nil {
			log.Fatal(err)
		}
		
		//iterate over all eclipses in this file
		for p, e := range eclipses {
			//find min and max values
			if e.Count < min_eclipse {
				min_eclipse = e.Count
			}
			if e.Count > max_eclipse {
				max_eclipse = e.Count
				max_eclipse_p = p
				max_eclipse_e = *e
			}
			eclipseSizes = append(eclipseSizes, e.Count)
			eclipse_count++
		}
		number_files++
	}
	
	//iterate over all files
	for _, file := range files {
		fname := file.Name()
		if !strings.Contains(fname, "sybils") {
			continue
		}
		log.Printf("Reading file: %s", fname)
		
		var sybils = make(map[string]*sybil1)
		
		raw, err := ioutil.ReadFile("./nodeInfo/" + fname)
		if err != nil {
			log.Fatal(err)
		}
		err = json.Unmarshal(raw, &sybils)
		if err != nil {
			log.Fatal(err)
		}
		
		//iterate over all /16 sybils in this file
		for p, s := range sybils {
			//find min and max values
			if s.Count < min_sybils16 {
				min_sybils16 = s.Count
			}
			if s.Count > max_sybils16 {
				max_sybils16 = s.Count
				max_sybils_p16 = p
				max_sybils_s16 = *s
			}
			sybilSizes16 = append(sybilSizes16, s.Count)
			sybil_count16++
			hist_sybils16[s.Count]++
			
			//iterate over all /24 sybils in this /16 subnet
			for q, t := range s.IP {
				//find min and max values
				if t.Count < min_sybils24 {
					min_sybils24 = t.Count
				}
				if t.Count > max_sybils24 {
					max_sybils24 = t.Count
					max_sybils_p24 = p + "." + strconv.Itoa(q)
					max_sybils_s24 = *t
				}
				sybilSizes24 = append(sybilSizes24, t.Count)
				sybil_count24++
				hist_sybils24[t.Count]++
			}
		}
	}
	
	//calculate averages
	sum_eclipses := 0
	
	for _, e := range eclipseSizes {
		sum_eclipses += e
	}
	
	avg_eclipse = float32(float32(sum_eclipses)/float32(len(eclipseSizes)))
	
	sum_sybils16 := 0
	sum_sybils24 := 0
	
	for _, s := range sybilSizes16 {
		sum_sybils16 += s
	}
	
	for _, s := range sybilSizes24 {
		sum_sybils24 += s
	}
	
	avg_sybils16 = float32(float32(sum_sybils16)/float32(len(sybilSizes16)))
	avg_sybils24 = float32(float32(sum_sybils24)/float32(len(sybilSizes24)))
	
	//sort arrays
	sort.Ints(eclipseSizes)
	sort.Ints(sybilSizes16)
	sort.Ints(sybilSizes24)
	
	//get the medians
	med_eclipse = eclipseSizes[len(eclipseSizes)/2]
	med_sybils16 = sybilSizes16[len(sybilSizes16)/2]
	med_sybils24 = sybilSizes24[len(sybilSizes24)/2]
	
	//get the biggest occurences
	huge_eclipse[max_eclipse_p] = &max_eclipse_e
	huge_sybils16[max_sybils_p16] = &max_sybils_s16
	huge_sybils24[max_sybils_p24] = &max_sybils_s24
	
	*Stats = stats {
		Sybil_count16: float32(float32(sybil_count16)/float32(number_files)),
		Min_sybils16: min_sybils16,
		Med_sybils16: med_sybils16,
		Avg_sybils16: avg_sybils16,
		Max_sybils16: max_sybils16,
		Hist_sybils16: hist_sybils16,
		Sybil_count24: float32(float32(sybil_count24)/float32(number_files)),
		Min_sybils24: min_sybils24,
		Med_sybils24: med_sybils24,
		Avg_sybils24: avg_sybils24,
		Max_sybils24: max_sybils24,
		Hist_sybils24: hist_sybils24,
		Eclipse_count: float32(float32(eclipse_count)/float32(number_files)),
		Min_eclipse: min_eclipse,
		Med_eclipse: med_eclipse,
		Avg_eclipse: avg_eclipse,
		Max_eclipse: max_eclipse,
		Huge_sybils16: huge_sybils16,
		Huge_sybils24: huge_sybils24,
		Huge_eclipse: huge_eclipse,
	}	
	
	sjson, err := json.MarshalIndent(Stats, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("./nodeInfo/misbehavingNodeStats.json", sjson, 0644)
	if err != nil {
		log.Fatal(err)
	}
	
	log.Printf("Ende")
}


