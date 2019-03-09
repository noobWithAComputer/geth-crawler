package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
//	"sync"
	"time"
)

type degree struct {
	Degree_min     int     `json:"degree_min"`
	Degree_med     int     `json:"degree_med"`
	Degree_avg     float32 `json:"degree_avg"`
	Degree_max     int     `json:"degree_max"`
	In_degree_min  int     `json:"in_degree_min"`
	In_degree_med  int     `json:"in_degree_med"`
	In_degree_avg  float32 `json:"in_degree_avg"`
	In_degree_max  int     `json:"in_degree_max"`
	Out_degree_min int     `json:"out_degree_min"`
	Out_degree_med int     `json:"out_degree_med"`
	Out_degree_avg float32 `json:"out_degree_avg"`
	Out_degree_max int     `json:"out_degree_max"`
}

type stats struct {
	Nodes               int     `json:"nodes"`
	Edges               int     `json:"edges"`
	Resilience_targeted float32 `json:"resilience_targeted"`
	Resilience_random   float32 `json:"resilience_random"`
	Degree              degree  `json:"degree"`
}

type Statistic struct {
	Avg_nodes               float32             `json:"avg_nodes"`
	Avg_edges               float32             `json:"avg_edges"`
	Avg_resilience_targeted float32             `json:"avg_resilience_targeted"`
	Avg_resilience_random   float32             `json:"avg_resilience_random"`
	Avg_degree_min          float32             `json:"avg_degree_min"`
	Avg_degree_med          float32             `json:"avg_degree_med"`
	Avg_degree_avg          float32             `json:"avg_degree_avg"`
	Avg_degree_max          float32             `json:"avg_degree_max"`
	Avg_in_degree_min       float32             `json:"avg_in_degree_min"`
	Avg_in_degree_med       float32             `json:"avg_in_degree_med"`
	Avg_in_degree_avg       float32             `json:"avg_in_degree_avg"`
	Avg_in_degree_max       float32             `json:"avg_in_degree_max"`
	Avg_out_degree_min      float32             `json:"avg_out_degree_min"`
	Avg_out_degree_med      float32             `json:"avg_out_degree_med"`
	Avg_out_degree_avg      float32             `json:"avg_out_degree_avg"`
	Avg_out_degree_max      float32             `json:"avg_out_degree_max"`
	Stats                   map[time.Time]stats `json:"stats"`
}

//reads all files from ./geo
//constructs multiple maps: nodes ([string][]int), nodesi ([int]string), edges ([string][]string) in two forms each: 1 (all nodes) and 2 (only online nodes)
//puts together adjacency lists for both cases
//saves the lists to "graphs/s-TIMESTAMP" and "graphs/s-TIMESTAMP_online"
//can ignore specific countries or AS organizations (see comments in for loop in function readMaps)
func main() {
	var allStats = new(Statistic)
	var statistics = make(map[time.Time]stats)
	
	//get all files from the directory
	files, err := ioutil.ReadDir("./data")
	if err != nil {
		log.Print("Error reading directory.")
		return
	}
	
	//iterate over all files
	for _, file := range files {
		fname := file.Name()
		
		if strings.Contains(fname, "online") || strings.Contains(fname, "asorgs") {
			readStats(fname, statistics)
		}
	}
	
	sum_nodes := 0
	sum_edges := 0
	sum_resilience_targeted := float32(0)
	sum_resilience_random := float32(0)
	sum_degree_min := 0
	sum_degree_med := 0
	sum_degree_avg := float32(0)
	sum_degree_max := 0
	sum_in_degree_min := 0
	sum_in_degree_med := 0
	sum_in_degree_avg := float32(0)
	sum_in_degree_max := 0
	sum_out_degree_min := 0
	sum_out_degree_med := 0
	sum_out_degree_avg := float32(0)
	sum_out_degree_max := 0
	
	for _, s := range statistics {
		sum_nodes += s.Nodes
		sum_edges += s.Edges
		sum_resilience_targeted += s.Resilience_targeted
		sum_resilience_random += s.Resilience_random
		sum_degree_min += s.Degree.Degree_min
		sum_degree_med += s.Degree.Degree_med
		sum_degree_avg += s.Degree.Degree_avg
		sum_degree_max += s.Degree.Degree_max
		sum_in_degree_min += s.Degree.In_degree_min
		sum_in_degree_med += s.Degree.In_degree_med
		sum_in_degree_avg += s.Degree.In_degree_avg
		sum_in_degree_max += s.Degree.In_degree_max
		sum_out_degree_min += s.Degree.Out_degree_min
		sum_out_degree_med += s.Degree.Out_degree_med
		sum_out_degree_avg += s.Degree.Out_degree_avg
		sum_out_degree_max += s.Degree.Out_degree_max
	}
	
	divisor := float32(len(statistics))
	
	*allStats = Statistic {
		Avg_nodes: float32(float32(sum_nodes)/divisor),
		Avg_edges: float32(float32(sum_edges)/divisor),
		Avg_resilience_targeted: float32(sum_resilience_targeted/divisor),
		Avg_resilience_random: float32(sum_resilience_random/divisor),
		Avg_degree_min: float32(float32(sum_degree_min)/divisor),
		Avg_degree_med: float32(float32(sum_degree_med)/divisor),
		Avg_degree_avg: float32(sum_degree_avg/divisor),
		Avg_degree_max: float32(float32(sum_degree_max)/divisor),
		Avg_in_degree_min: float32(float32(sum_in_degree_min)/divisor),
		Avg_in_degree_med: float32(float32(sum_in_degree_med)/divisor),
		Avg_in_degree_avg: float32(sum_in_degree_avg/divisor),
		Avg_in_degree_max: float32(float32(sum_in_degree_max)/divisor),
		Avg_out_degree_min: float32(float32(sum_out_degree_min)/divisor),
		Avg_out_degree_med: float32(float32(sum_out_degree_med)/divisor),
		Avg_out_degree_avg: float32(sum_out_degree_avg/divisor),
		Avg_out_degree_max: float32(float32(sum_out_degree_max)/divisor),
		Stats: statistics,
	}
	
	sjson, err := json.MarshalIndent(allStats, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("./nodeInfo/statistics.json", sjson, 0644)
	if err != nil {
		log.Fatal(err)
	}
	
	log.Printf("Ende, len: %d", len(files))
}


func readStats(fname string, statistics map[time.Time]stats) {
	log.Printf("Start reading from directory %s", fname)
	
	raw, err := ioutil.ReadFile("./data/" + fname + "/DEGREE_DISTRIBUTION/_singles.txt")
	if err != nil {
		log.Fatal(err)
	}
	
	content := string(raw)
	contents := strings.Split(content, "\t")
	
//	for i, t := range contents {
//		log.Printf("  %d: %s", i, t)
//	}
	
	nodes, _ := strconv.ParseFloat(contents[1], 32)
	edges, _ := strconv.ParseFloat(contents[9], 32)
	degree_min, _ := strconv.ParseFloat(contents[17], 32)
	degree_med, _ := strconv.ParseFloat(contents[25], 32)
	degree_avg, _ := strconv.ParseFloat(contents[33], 32)
	degree_max, _ := strconv.ParseFloat(contents[41], 32)
	in_degree_min, _ := strconv.ParseFloat(contents[49], 32)
	in_degree_med, _ := strconv.ParseFloat(contents[57], 32)
	in_degree_avg, _ := strconv.ParseFloat(contents[65], 32)
	in_degree_max, _ := strconv.ParseFloat(contents[73], 32)
	out_degree_min, _ := strconv.ParseFloat(contents[81], 32)
	out_degree_med, _ := strconv.ParseFloat(contents[89], 32)
	out_degree_avg, _ := strconv.ParseFloat(contents[97], 32)
	out_degree_max, _ := strconv.ParseFloat(contents[15], 32)
	
	raw, err = ioutil.ReadFile("./data/" + fname + "/CRITICAL_POINTS-true-LARGEST/_singles.txt")
	if err != nil {
		log.Fatal(err)
	}
	
	content = string(raw)
	contents = strings.Split(content, "\t")
	
	resilience_targeted, _ := strconv.ParseFloat(contents[1], 32)
	
	raw, err = ioutil.ReadFile("./data/" + fname + "/CRITICAL_POINTS-true-RANDOM/_singles.txt")
	if err != nil {
		log.Fatal(err)
	}
	
	content = string(raw)
	contents = strings.Split(content, "\t")
	
	resilience_random, _ := strconv.ParseFloat(contents[1], 32)
	
	thisDegree := new(degree)
	
	*thisDegree = degree {
		Degree_min: int(degree_min),
		Degree_med: int(degree_med),
		Degree_avg: float32(degree_avg),
		Degree_max: int(degree_max),
		In_degree_min: int(in_degree_min),
		In_degree_med: int(in_degree_med),
		In_degree_avg: float32(in_degree_avg),
		In_degree_max: int(in_degree_max),
		Out_degree_min: int(out_degree_min),
		Out_degree_med: int(out_degree_med),
		Out_degree_avg: float32(out_degree_avg),
		Out_degree_max: int(out_degree_max),
	}
	
	thisStat := new(stats)
	
	*thisStat = stats {
		Nodes: int(nodes),
		Edges: int(edges),
		Resilience_targeted: float32(resilience_targeted),
		Resilience_random: float32(resilience_random),
		Degree: *thisDegree,
	}
	
	timestamp, err := time.Parse("2006-01-02--15-04-05", fname[16:36])
	if err != nil {
		log.Fatal(err)
		return
	}
	
	statistics[timestamp] = *thisStat
	
	log.Printf("Finished reading stats from directory %s", fname)
	
}


func writeToFile(nodes map[string]int, nodesi map[int]string, edges map[string][]string, fname string) {
	log.Printf("Start writing to file %s", fname)
	f, err := os.Create("./" + fname)
	if err != nil {
		log.Printf("Error creating file %s", fname)
		return
	}
	
	//get the node and edge count of the graph
	ecount := 0
	ncount := len(nodes)
	for _, v := range edges {
		ecount += len(v)
	}
	
	//write the header of the file
	f.Write([]byte("test\n"))
	f.Write([]byte(strconv.Itoa(ncount) + "\n"))
	f.Write([]byte(strconv.Itoa(ecount) + "\n"))
	f.Write([]byte("\n"))
	
	
	//iterate over all nodes
	for i := 0; i < len(nodes); i++ {
		//write its int ID
		f.Write([]byte(strconv.Itoa(i) + ":"))
		
		//iterate over all edges of the current node
		for j, id := range edges[nodesi[i]] {
			//write the int ID of the connected node
			f.Write([]byte(strconv.Itoa(nodes[id])))
			
			//if this was the last connection
			if j == (len(edges[nodesi[i]]) - 1) {
				break
			}
			//otherwise write ";" for more connections
			f.Write([]byte(";"))
		}
		
		f.Write([]byte("\n"))
	}
	
	err = f.Close()
	if err != nil {
		log.Printf("Error closing file %s", fname)
		return
	}
}


func isStringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}











