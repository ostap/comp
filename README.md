Examples:
```
[ e.name, e.elevation | e <- geonames, e.name =~ "Z[uü]e?rich" ]
[ e.name, e.lat, e.lon | e <- geonames, dist(e.lat, e.lon, 47.366667, 8.55) < 0.500 ]
[ e.place, e.lat, e.lon | e <- zipcodes, e.zip == 8008 ]
```

See [wikipedia](http://en.wikipedia.org/wiki/List_comprehension) for more information.

Here is how to build and run the binary:
``` bash
$ go tool yacc -o src/comp/y.go -p "comp_" src/comp/grammar.y
$ go test comp
$ go install comp
$ comp -bind :9090 -data geonames.txt,zipcodes.txt
$ lynx localhost:9090
```
