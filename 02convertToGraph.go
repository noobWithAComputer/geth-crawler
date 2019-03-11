package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
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

type node struct {
	Id        string
	Ip        string
	Port      int
	Reachable bool
}


//reads all files from ./geo
//constructs multiple maps: nodes (map[string][]int), nodesi (map[int]string), edges (map[string][]string) in two forms each: 1 (all nodes) and 2 (only online nodes)
//puts together adjacency lists for both cases
//saves the lists to "graphs/s-TIMESTAMP" and "graphs_online/s-TIMESTAMP_online"
//can ignore specific countries or AS organizations (see comments in for loop in function readMaps)
func main() {
	//get all files from the directory
	files, err := ioutil.ReadDir("./geo")
	if err != nil {
		log.Fatal(err)
	}
	
	//iterate over all files
	for _, file := range files {
		fname := file.Name()
		
		readMaps(fname)
	}
	
	log.Printf("Ende, len: %d", len(files))
}


func readMaps(fname string) {
	log.Printf("Start reading from file %s", fname)
	//int->node
	var thisM = make(map[int]myNode)
	
	//nodeID->[]int
	//nodeID pointing to the corresponding int representation (all nodes)
	var nodes1 = make(map[string]int)
	//int->nodeID
	//reverse of the above
	var nodesi1 = make(map[int]string)
	
	//nodeID->[]int
	//nodeID pointing to the corresponding int representation (online nodes)
	var nodes2 = make(map[string]int)
	//int->nodeID
	//reverse of the above
	var nodesi2 = make(map[int]string)
	
	//nodeID->[]nodeID
	//connections from nodeID to other nodeIDs (all nodes)
	var edges1 = make(map[string][]string)
	//nodeID->[]nodeID
	//connections from nodeID to other nodeIDs (online nodes)
	var edges2 = make(map[string][]string)
	
	//read the given file into thisM
	raw, err := ioutil.ReadFile("./geo/" + fname)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(raw, &thisM)
	if err != nil {
		log.Fatal(err)
	}
	
	//node counter all nodes
	i := 0
	//node counter online nodes
	j := 0
	//iterate over thisM (all nodes)
	for _, n := range thisM {
	//	if n.Country == "United States" || n.ASO == "Amazon.com, Inc." {
	//		continue
	//	}
	//remove the above comment to filter nodes from specific countries and/or AS organizations out
		//if the node ID is not yet present in nodes1
		if _, ok := nodes1[n.Id]; !ok {
		//add the current it to both maps (int->nodeID and nodeID->int)
			nodes1[n.Id] = i
			nodesi1[i] = n.Id
			i++
		}
		//if the node ID is not yet present in nodes2
		if _, ok := nodes2[n.Id]; !ok {
			//add the current it to both maps (int->nodeID and nodeID->int)
			nodes2[n.Id] = j
			nodesi2[j] = n.Id
			j++
		}
	}
	
	//iterate over thisM (all nodes)
	for _, v := range thisM {
		//iterate over the connections of the current node
		for conn, _ := range v.Connections {
			//if this connected node is not yet in the node map for online nodes
			if _, ok := nodes1[conn]; !ok {
				//add it to both maps (int->node and nodeID->int)
				nodes1[conn] = i
				nodesi1[i] = conn
				i++
			}
		}
	}
	
	//iterate over thisM (all nodes)
	for _, v := range thisM {
		//iterate over the connections of the current node
		for conn, _ := range v.Connections {
			//add this connection to the current node
			edges1[v.Id] = append(edges1[v.Id], conn)
			//if also the connected node is in the nodes map for all nodes
			if _, ok := nodes2[conn]; ok {
				//if it was not yet added as connection for the current node
				if !isStringInSlice(conn, edges2[v.Id]) {
					//add it
					edges2[v.Id] = append(edges2[v.Id], conn)
				}
			}
		}
	}
	
	newFilename1 := "graphs/" + fname[:len(fname)-5] + ""
	newFilename2 := "graphs_online/" + fname[:len(fname)-5] + "_online"
	
	writeToFile(nodes1, nodesi1, edges1, newFilename1)
	writeToFile(nodes2, nodesi2, edges2, newFilename2)
	
	log.Printf("Finished writing to files from original file %s", fname)
	
}


func writeToFile(nodes map[string]int, nodesi map[int]string, edges map[string][]string, fname string) {
	log.Printf("Start writing to file %s", fname)
	f, err := os.Create("./" + fname)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
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











