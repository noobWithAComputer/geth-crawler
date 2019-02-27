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
	var thisM = make(map[int]myNode)
	var thisMi = make(map[string]int)
	var mc = make(map[string][]string)
	var nc = make(map[int]string)
	var nci = make(map[string]int)
	var mas = make(map[string][]string)
	var nas = make(map[int]string)
	var nasi = make(map[string]int)
	var eci = make(map[int][]int)
	var easi = make(map[int][]int)
	
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
	
	for i, n := range thisM {
		thisMi[n.Id] = i
	}
	
	for _, n := range thisM {
		for conn, _ := range n.Connections {
			if _, ok := thisMi[conn]; ok {
				if !isStringInSlice(thisM[thisMi[conn]].Country, mc[n.Country]) {
					mc[n.Country] = append(mc[n.Country], thisM[thisMi[conn]].Country)
				}
				if !isStringInSlice(thisM[thisMi[conn]].ASO, mas[n.ASO]) {
					mas[n.ASO] = append(mas[n.ASO], thisM[thisMi[conn]].ASO)
				}
			}
		}
	}

	i := 0
	for c, _ := range mc {
		if _, ok := nci[c]; !ok {
			nc[i] = c
			nci[c] = i
			i++
		}
	}

	for c, conns := range mc {
		for _, conn := range conns {
			if !isIntInSlice(nci[conn], eci[nci[c]]) {
				eci[nci[c]] = append(eci[nci[c]], nci[conn])
			}
		}
	}

	j := 0
	for as, _ := range mas {
		if _, ok := nasi[as]; !ok {
			nas[j] = as
			nasi[as] = j
			j++
		}
	}

	for as, conns := range mas {
		for _, conn := range conns {
			if !isIntInSlice(nasi[conn], easi[nasi[as]]) {
				easi[nasi[as]] = append(easi[nasi[as]], nasi[conn])
			}
		}
	}
	
	newFilename1 := "graphs_countries/" + fname[:len(fname)-5] + "_countries"
	newFilename2 := "graphs_asorgs/" + fname[:len(fname)-5] + "_asorgs"
	
	writeToFile(eci, newFilename1)
	writeToFile(easi, newFilename2)
	
	log.Printf("Finished writing to files from original file %s", fname)
	
}


func writeToFile(edges map[int][]int, fname string) {
	log.Printf("Start writing to file %s", fname)
	f, err := os.Create("./" + fname)
	if err != nil {
		log.Printf("Error creating file %s", fname)
		return
	}
	
	ecount := 0
	ncount := len(edges)
	for _, v := range edges {
		ecount += len(v)
	}
	log.Printf("  Edges: %d", ecount)
	
	f.Write([]byte("test\n"))
	f.Write([]byte(strconv.Itoa(ncount) + "\n"))
	f.Write([]byte(strconv.Itoa(ecount) + "\n"))
	f.Write([]byte("\n"))
	
//	log.Printf("  Starting first iteration")
	
	for i := 0; i < len(edges); i++ {
		f.Write([]byte(strconv.Itoa(i) + ":"))
		e := edges[i]
		for j := 0; j < len(e); j++ {
			f.Write([]byte(strconv.Itoa(e[j])))
			if j == (len(e) - 1) {
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

func isIntInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}









