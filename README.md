### comp

comp is a tool for querying information from files. It's main goal is
to provide a unified interface to the variety of data representations found
in public data sets. To achieve this goal comp introduces a small query
language with type coercion (e.g. "2" + 2 == 4) and a powerful iteration
mechanism based on [list comprehensions][1].

    $ comp -bind=:9090
    $ curl -d '{"expr": "\"2\" + 2"}' http://localhost:9090/full
    {"result": 4, "time": "139.581us"}%
    $ curl -d '{"expr": "[i | i <- [1, 2, 3], i != 2]"}' http://localhost:9090/full
    {"result": [ 1, 3 ], "time": "400.913us"}
    $ curl -d '{"expr": "[i * j | i <- [1, 2, 3], j <- [10, 20]]"}' http://localhost:9090/full
    {"result": [ 10, 20, 20, 40, 30, 60 ], "time": "267us"}

to load a tab delimited file and query its contents:

    $ comp -data=contacts.txt -bind=:9090
    $ curl -d '[ c | c <- contacts, c.zip == 8001]' http://localhost:9090/full

to run the queries through an interactive web interface open the http://localhost:9090/console in your browser.

### build and test

    $ cd comp && export GOPATH=$GOPATH:$(pwd)
    $ go tool yacc -o src/comp/y.go -p "comp_" src/comp/grammar.y
    $ go test comp
    $ go install comp

### acknowledgements

comp tool was developed for the purpose of [mingle.io - Query API for Open Data](https://mingle.io) as the backend 
behind the service. It allows mingle.io users to run arbitrary queries across variety of Open Data sets and mix them
together to enrich applications with quality information.

comp language borrows ideas from other programming languages (Haskell,
JavaScript and probably others), but its core (application of comprehensions
to formulate queries) is based on a research paper by Peter Buneman
["comprehension syntax"][1]. I explored this paper thanks to feedback
provided by [Ted Leung][2] after the [Emerging Languages Camp][3] regarding
[Bandicoot][4].

[1]: http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.26.993 "Comprehension Syntax (1994), by Peter Buneman , Leonid Libkin, Dan Suciu, Val Tannen, Limsoon Wong"
[2]: http://www.sauria.com/blog/2012/09/27/strange-loop-2012/ "Ted Leung's Blog"
[3]: https://thestrangeloop.com/preconf "Emerging Languages Camp"
[4]: http://bandilab.org "Bandicoot"
