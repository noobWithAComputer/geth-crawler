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

The last two scripts analyze some further information from the snapshots.

```
go run 04analyzeSessions.go
```

will create one file: **`/nodeInfo/sessionInfo.json`**. It contains the following information for every node:

```
node {
  sessionlength int
  intersession  int
  count         [2]int
  reachable     map[string]bool
}
``` 

The sessionlength is the avverage length of all seesion counted in "snapshots". So if you have four snapshots and the node in question was online in the first, third and fourth snapshot, the sessionlength would be 1.5 as it was online in three snapshots, but as it was offline in the mean time, this will be divided by the two session the node was online in. Same for the intersessions.
Count ontains two ints: the first is the number of snapshots, this node was online at, the second is the number of snapshots, where the node was offline.
Reachable is a map that holds if the node was online at s specific snapshot; the string is the timestamp of this snapshot.

```
go run 05findMisbehavingNodes.go
```

will create two files **`/nodeInfo/s-TIMESTAMP_eclipse.json`** and **`/nodeInfo/s-TIMESTAMP_sybils.json`**.
The eclipse file contains nodes which have a common prefix of two bytes with at least one other node. This includes nodes with the entirely same ID. This file additionally contains the number of node per prefix and the geo information of these nodes.
The sybil file contains nodes which share a specific subnet. The nodes are grouped by /16 subnets and further into /24 subnets. This file also contains the number of nodes both per /16 and /24 subnet and the geo information of these nodes.

## Performance

Both the client and the go scripts were used on an average PC with sufficiently good results.
The crawler is timecontrolled so, it will always take the same amount of time to crawl the network. Only the number of found nodes may vary. The main requirements for good performance is a fast CPU (should be at ~3GHz) and a good internet connection, as the crawler sends thousands of pings to other nodes.
On my PC the go scripts performed with the following times:
* 01: 2-3 seconds per snapshot
* 02: 14-15 seconds per snapshot
* 03: 2 seconds per snapshot
* 04: 1 second per snapshot
* 05: 1 second per snapshot
