# RoutePlanner Backend
This repository contains code that powers the [routeplanner app](https://yurachistic1.github.io/routeplanner-frontend/). The app suggests circular walks or runs based on start location and desired distance. 

##  Motivation
I have been interested in making an app like that for a while since I started running more frequently. I normally use google maps and add multiple destinations to work out a distance of a potential run, but that is rather cumbersome way of doing it. Secondly I thought it would be a good challenge since I would have to figure out how to get suitable map data and then design an algorithm that would produce decent quality circular paths. 

## Implementation
This service is designed to be used as serverless function hosted on google cloud platform with an API gateway to it. 

Map data is retrieved from a public OSM server and this is also the slowest step in the process as it can take a few seconds to download the data. The data is then parsed and routes are calculated. Best 25 results are returned encoded as JSON. 

## Running locally
Requires Go 1.22 or later.

```sh
go test ./...
go run ./cmd/server   # listens on $PORT, default 8080
curl 'localhost:8080/?lat=51.5313&lon=-0.1570&distance=10'
```

`distance` is in km and must be between 1 and 35.

Map data comes from tiles generated ahead of time from an OSM extract, read
from `tiledata/` or the directory in `$TILES_DIR`. Where there are no tiles
the API falls back to the public Overpass API, which is slow, so it's limited
to 10km there.

To generate tiles, download an extract from https://download.geofabrik.de
and run the generator from its own directory:

```sh
cd cmd/tilegen
CGO_ENABLED=0 go run . -in path/to/greater-london-latest.osm.pbf -out ../../tiledata
```

`CGO_ENABLED=0` lets the OSM reading library build without `pkg-config`.
Greater London takes a few seconds and makes about 24MB of tiles.
 
## Result 
The quality of routes does vary a lot. Some areas especially urban areas with a lot of suitable roads and paths can produce great results. Other times the routes might be hard to navigate due to large number of turns or the fact that they keep going back on themselves. This variability was the main reason to return best 25 as it makes it more likely that some of the suggestions are suitable. 

Example of good suggestions:

![good routes in knole park](img/routes.gif)

Back in high school I took part in the Knole run which is a 10km cross country run made up of two loops.

Last thing to note is while I tried to eliminate the possibility of roads and paths that are not suitable for pedestrians being used, they still will likely appear sometimes as OSM data is not always complete or up to date in different parts of the world.  