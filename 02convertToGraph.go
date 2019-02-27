package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
//	"sync"
)

//type ID [32]byte

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

type node struct {
	Id        string
	Ip        string
	Port      int
	Reachable bool
}


// reads all files from ./geo
// constructs multiple maps: nodes ([string][]int), nodesi ([int]string), edges ([string][]string) in two forms each: 1 (all nodes) and 2 (only online nodes)
// puts together adjacency lists for both cases
// saves the lists to "graphs/s-TIMESTAMP" and "graphs/s-TIMESTAMP_online"
// can ignore specific countries or AS organizations (see comments in for loop in function readMaps)
func main() {
	files, err := ioutil.ReadDir("./geo")
	if err != nil {
		log.Print("Error reading directory.")
		return
	}
	
	for _, file := range files {
		fname := file.Name()
		
		readMaps(fname)
	}
	
	log.Printf("Ende, len: %d", len(files))
}


func readMaps(fname string) {
	log.Printf("Start reading from file %s", fname)
//	var input = make(map[string]myNode)
	var thisM = make(map[int]myNode)
	var thisMi = make(map[string][]int)
	var nodes1 = make(map[string][]int)
	var nodesi1 = make(map[int]string)
	var nodes2 = make(map[string][]int)
	var nodesi2 = make(map[int]string)
	var edges1 = make(map[string][]string)
	var edges2 = make(map[string][]string)
	var duplicates = make(map[string][]myNode)
	
	raw, err := ioutil.ReadFile("./geo/" + fname)
	if err != nil {
		log.Printf("Error opening file %s", fname)
		return
	}
	err = json.Unmarshal(raw, &thisM)
	if err != nil {
		log.Printf("Error unmarshalling file %s", fname)
		return
	}
	
	for k, n := range thisM {
		thisMi[n.Id] = append(thisMi[n.Id], k)
	}
	
	i := 0
	for _, n := range thisM {
	//	if n.Country == "United States" || n.ASO == "Amazon.com, Inc." {
	//		continue
	//	}
	// remove the above comment to filter nodes from specific countries and/or AS organizations out
		nodes1[n.Id] = append(nodes1[n.Id], i)
		nodesi1[i] = n.Id
		nodes2[n.Id] = append(nodes2[n.Id], i)
		nodesi2[i] = n.Id
		
		i++
	}
	
	for _, v := range thisM {
		for conn, _ := range v.Connections {
			if _, ok := nodes1[conn]; !ok {
				nodes1[conn] = append(nodes1[conn], i)
				nodesi1[i] = conn
				i++
			}
			
			edges1[v.Id] = append(edges1[v.Id], conn)
			if _, ok := nodes2[conn]; ok {
				if !isStringInSlice(conn, edges2[v.Id]) {
					edges2[v.Id] = append(edges2[v.Id], conn)
		//			edges2[conn] = append(edges2[conn], v.Id)
				}
			}
		}
	}
	
	newFilename1 := "graphs/" + fname[:len(fname)-5] + ""
	newFilename2 := "graphs_online/" + fname[:len(fname)-5] + "_online"
	
	writeToFile(nodes1, nodesi1, edges1, newFilename1)
	writeToFile(nodes2, nodesi2, edges2, newFilename2)
	
	for k, v := range thisMi {
		if len(v) < 2 {
			continue
		}
//		log.Printf("  k=%s, len(v)=%d", k, len(v))
		var thisArr = []myNode{}
		for _, j := range v {
//			log.Printf("    j=%d", j)
			thisArr = append(thisArr, thisM[j])
		}
		duplicates[k] = thisArr
	}
	
	// format json
	duplicatesJson, err := json.MarshalIndent(duplicates, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	
	// save json file
	err = ioutil.WriteFile("./nodeInfo/" + fname[:len(fname)-5] + "_duplicates.json", duplicatesJson, 0644)
	if err != nil {
		log.Fatal(err)
	}
	
	log.Printf("Finished writing to files from original file %s", fname)
	
}


func writeToFile(nodes map[string][]int, nodesi map[int]string, edges map[string][]string, fname string) {
	log.Printf("Start writing to file %s", fname)
	f, err := os.Create("./" + fname)
	if err != nil {
		log.Printf("Error creating file %s", fname)
		return
	}
	
	ecount := 0
	ncount := len(nodes)
	for _, v := range edges {
		ecount += len(v)
	}
	log.Printf("  Edges: %d", ecount)
	
	f.Write([]byte("test\n"))
	f.Write([]byte(strconv.Itoa(ncount) + "\n"))
	f.Write([]byte(strconv.Itoa(ecount) + "\n"))
	f.Write([]byte("\n"))
	
//	log.Printf("  Starting first iteration")
	
	for i := 0; i < len(nodes); i++ {
		f.Write([]byte(strconv.Itoa(i) + ":"))
		
//		log.Printf("    Starting second iteration, len=%d, i=%d", len(nodes), i)
		for j, id := range edges[nodesi[i]] {
			f.Write([]byte(strconv.Itoa(nodes[id][0])))
			
			if j == (len(edges[nodesi[i]]) - 1) {
//				log.Printf("      Exiting loop, j=%d", j)
				break
			}
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











