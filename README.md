Examples:
``` 
[ e.name, e.elevation | e <- geonames, e.name =~ "Z[uü]e?rich" ]
[ e.name, e.lat, e.lon | e <- geonames, dist(e.lat, e.lon, 47.366667, 8.55) < 0.500 ]
```

See [wikipedia](http://en.wikipedia.org/wiki/List_comprehension) for more information.

Here is how to build a binary:
```
$ go tool yacc -o src/comp/y.go -p "comp_" src/comp/grammar.y
$ go test comp
$ go install comp
```
