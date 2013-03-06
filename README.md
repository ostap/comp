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
$ comp -data geonames.txt,zipcodes.txt
$ open http://localhost:9090/console
```

Distributed mode:
``` bash
host1$ comp -data geonames.0.txt,zipcodes.0.txt -peers http://host2:9090/part
host2$ comp -data geonames.1.txt,zipcodes.1.txt -peers http://host1:9090/part
```
