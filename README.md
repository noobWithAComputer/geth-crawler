# Geth-Crawler

A modified geth client to crawl the Ethereum p2p network and create snapshots of all the nodes online. Additionally there are a few go scripts to prepare the snapshots for use with [GTNA](https://github.com/BenjaminSchiller/GTNA).

## Components

### Geth client

To build the geth client both a Go and a C compiler are required. The client can then be build within the **`/geth-crawler`** directory via

```
make
```

To start the client in console, run 

```
./build/bin/geth --fakepow console
```

To run the client in background, there is a shell script named *start_crawler_background.sh*. With this, the client will run in background and will log to *output.log* and write errors into *error.log*.

With the flag **`--verbosity=i`** (i = 1..5) the level of output information can be set. 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail. 3 is suitable for most cases. To debug the verbosity can be set to 4.

### Go scripts

To run the go scripts we need two resources:

* The [GeoLite2](https://dev.maxmind.com/geoip/geoip2/geolite2/) city and ASN databases. Extract them to **`/GeoLite`**.
* The go reader for these databases [MaxMind DB Reader](https://github.com/oschwald/maxminddb-golang). You can install it via

```
go get github.com/oschwald/maxminddb-golang
```

To run the scripts, use

```
go run [filename]
```

## Functionality

The project is divided into multiple steps:

* Gather snapshots
* Enrich the snapshots with geo information
* Create graphs readyble by GTNA
* Further analysis of the snapshots
* Better representation of the data

### 1. Gather snapshots

The crawler takes about 15 minutes to crawl the network. So after starting the Crawler, the first snapshot will be created 15 minutes after that. All further snapshots are created in 60 minute intervals.
The snapshots are stored in **`geth-crawler/snapshots`**. To use the go scripts, all snapshots should be copied (or moved) to **`/snapshots`**.
The snapshots are json files containing a go map of node found in the network. The filename is **`s-TIMESTAMP.json`**. The timestamp is given as 2006-01-02--15-04-05. The node structure looks like this:

```
node {
  id          string
  ip          string
  port        int
  reachable   bool
  connections map[bool]string
}
```

Further the script creates **`/nodeInfo/s-TIMESTAMP_geo_counts.json`**. This file contains the number of nodes per country and per AS organization.

### 2. Enrich the snapshots

Now to enrich the snapshots with geo information, run 

```
go run 01getGeolocation.go
```

This will add a few information to the snapshot json files. The new json files are stored in **`/geo`** and have this structure:

```
node {
  id          string
  ip          string
  port        int
  reachable   bool
  country     string
  subdivision string
  city        string
  aso         string
  connections map[bool]string
}
```

**`ASO`** is the autonomous system organization.

### 3. Create GTNA graphs

GTNA requires adjacency lists of the nodes in the network. The second and third script will create those lists from our snapshots.
The both of these scripts will create two adjacency lists.

```
go run 02convertToGraphs.go
```

This script will store its jsons in **`/graphs`** and **`/graphs_online`**. The former contain all nodes found and also the nodes which were known by those nodes. The latter only containes the nodes, which were online at the time of crawling.

```
go run 03convertToGraphsASC.go
```

This script will store its jsons in **`/graphs_asorgs`** and **`/graphs_countries`**. These graphs hold combined nodes, which represent a whole AS organization or a whole country. The connections between the ASOs or the countries are the ones of the nodes from the different ASOs or countries.
Example: If there is a node A in country A and there is a node B in country B, in the country graph will be a connection between country A and country B.

All the graphs within the four graph directories can be read by GTNA to analyze their graph properties.

### 4. Further analysis

The next four scripts analyze some further information from the snapshots.

```
go run 04analyzeSessions.go
```

will create four files: **`/nodeInfo/sessionInfo.json`**, **`/nodeInfo/sessionInfoAverage.json`**, **`/nodeInfo/sessionInfoC.json`** and **`/nodeInfo/sessionInfoAS.json`**.
The first contains the following information for every node:

```
stats {
  sessionlength     int
  intersession      int
  sessioncount      int
  intersessioncount int
  count             [2]int
  reachable         map[string]bool
}
```

The sessionlength is the avverage length of all session counted in "snapshots". So if you have four snapshots and the node in question was online in the first, third and fourth snapshot, the sessionlength would be 1.5 as it was online in three snapshots, but as it was offline in the mean time, this will be divided by the two session the node was online in. Same for the intersessions. Sessioncount and intersessioncount are the number of sessions and intersessions respectively. Count contains two ints: the first is the number of snapshots, this node was online at, the second is the number of snapshots, where the node was offline. Reachable is a map that holds if the node was online at s specific snapshot; the string is the timestamp of this snapshot.
The second file contains the absolut number of distinct nodes found, averages of the first four values from the file described above and a histogram of how often a session length and how often an intersession length occured:

```
statsAverages {
  nodes                 int
  avgsessionlength      float32
  medsessionlength      float32
  avgintersessionlength float32
  medintersessionlength float32
  sessions              map[int]int
  intersessions         map[int]int
}
```

The last two files contain the following stats for countries and ASes respectively.

```
stats {
  sessionlength     int
  intersession      int
  sessioncount      int
  intersessioncount int
}
```


```
go run 05findMisbehavingNodes.go
```

will create two files **`/nodeInfo/s-TIMESTAMP_eclipse.json`** and **`/nodeInfo/s-TIMESTAMP_sybils.json`** per snapshot. The eclipse file contains nodes which have a common prefix of two bytes with at least one other node. This includes nodes with the entirely same ID. This file additionally contains the number of node per prefix and the geo information of these nodes. The sybil file contains nodes which share a specific subnet. The nodes are grouped by /16 subnets and further into /24 subnets. This file also contains the number of nodes both per /16 and /24 subnet and the geo information of these nodes.

### 5. Better representation of the data

```
go run 06calculateStatistics.go
```

will create the files **`/nodeInfo/statisticsA.json`** and **`/nodeInfo/statisticsO.json`**. Both hold information about the (total/in/out) degree of nodes and the resilience of the network.

```
Statistic {
  Avg_nodes               float32
  Avg_edges               float32
  Min_resilience_targeted float32
  Avg_resilience_targeted float32
  Max_resilience_targeted float32
  Min_resilience_random   float32
  Avg_resilience_random   float32
  Max_resilience_random   float32
  Avg_degree_min          float32
  Avg_degree_med          float32
  Avg_degree_avg          float32
  Avg_degree_max          float32
  Avg_in_degree_min       float32
  Avg_in_degree_med       float32
  Avg_in_degree_avg       float32
  Avg_in_degree_max       float32
  Avg_out_degree_min      float32
  Avg_out_degree_med      float32
  Avg_out_degree_avg      float32
  Avg_out_degree_max      float32
  Stats                   map[time.Time]stats
}
```

while stats looks like this:

```
stats {
  Nodes               int
  Edges               int
  Resilience_targeted float32
  Resilience_random   float32
  Degree              degree
}
```

and degree like this:

```
degree {
  Degree_min     int
  Degree_med     int
  Degree_avg     float32
  Degree_max     int
  In_degree_min  int
  In_degree_med  int
  In_degree_avg  float32
  In_degree_max  int
  Out_degree_min int
  Out_degree_med int
  Out_degree_avg float32
  Out_degree_max int
}
```


```
go run 07getMisbehavingNodeStatistics.go
```

gathers statistics about the misbehaving nodes collected with 05findMisbehavingnodes.go. It saves the minimum, median, average and max values for sybils on /16 subnet level, sybils on /24 subnet levels and eclipses, saves histograms for both sybil types and saves the biggest occurence found of all types.
They look like this:

```
stats {
  Sybil_count16 float32
  Min_sybils16  int
  Med_sybils16  int
  Avg_sybils16  float32
  Max_sybils16  int
  Hist_sybils16 map[int]int
  Sybil_count24 float32
  Min_sybils24  int
  Med_sybils24  int
  Avg_sybils24  float32
  Max_sybils24  int
  Hist_sybils24 map[int]int
  Eclipse_count float32
  Min_eclipse   int
  Med_eclipse   int
  Avg_eclipse   float32
  Max_eclipse   int
  Huge_sybils16 map[string]*sybil1
  Huge_sybils24 map[string]*sybil2
  Huge_eclipse  map[string]*eclipse
}
```

sybil1, sybil2 and eclipse look like these respectively:

```
eclipse {
  Count int
  Nodes map[string][]node
}

sybil1{
  Count int
  IP    map[int]*sybil2
}

sybil2{
  Count int
  Nodes map[int][]node
}
```

```
go run 08getASHistogramOverTime.go
```

creates a histogram of all AS and their corresponding number of nodes per snapshot. It saves this histogram in **`/nodeInfo/asTimeStats.txt`**.


## Performance

Both the client and the go scripts were used on an average PC with sufficiently good results.
The crawler is timecontrolled so, it will always take the same amount of time to crawl the network. Only the number of found nodes may vary. The main requirements for good performance is a fast CPU (should be at ~3GHz) and a good internet connection, as the crawler sends thousands of pings to other nodes.
On my PC the go scripts performed with the following times:
* 01: 2-3 seconds per snapshot
* 02: 14-15 seconds per snapshot
* 03: 2 seconds per snapshot
* 04: 1 second per snapshot
* 05: 1 second per snapshot
* 06: 30 milliseconds per snapshot
* 07: 30 milliseconds per snapshot
* 08: 40 milliseconds per snapshot



