### comp

comp is a tool for querying information from files. Its main goal is
to provide a unified interface to the variety of data representations found
in public data sets. To achieve this goal comp introduces a small query
language with type coercion (e.g. "2" + 2 == 4) and a powerful iteration
mechanism based on [list comprehensions][0].

    $ comp -bind=:9090
    $ curl -d '{"expr": "\"2\" + 2"}' http://localhost:9090/full
    {"result": 4, "time": "139.581us"}%
    $ curl -d '{"expr": "[i | i <- [1, 2, 3], i != 2]"}' http://localhost:9090/full
    {"result": [ 1, 3 ], "time": "400.913us"}
    $ curl -d '{"expr": "[i * j | i <- [1, 2, 3], j <- [10, 20]]"}' http://localhost:9090/full
    {"result": [ 10, 20, 20, 40, 30, 60 ], "time": "267us"}

to load a tab delimited file and query its contents:

    $ comp -data=contacts.txt -bind=:9090
    $ curl -d '{"expr": "[ c | c <- contacts, c.zip == 8001]"}' http://localhost:9090/full

you can also run queries through a web console on `http://localhost:9090/console`.

### syntax overview

comp defines the following types:
  * scalar - `53`, `3.14`, `"hello"`, `true`
  * list - `[10, 20, 30]`, `["a", "b"]`
  * object - `{id: 123, name: "hello"}`

and the following operators:
  * `!` - not
  * `* /` - multiply, divide
  * `+ - ++` - addition, subtraction, string concatenation
  * `< <= > >=` - less than [or equal], greater than [or equal]
  * `== != =~` - [not] equal, regular expression match
  * `&& ||` - logical and, logical or

A list comprehension is a constructs of the following form:

    [ e | g1, g2, ..., gN ]

pronounced as "a list of all e where g1 and g2 ... and gN", where `e`
represents an expression and `gX` is either an iteration over a list
or a boolean expression:

    [i*2 | i <- [1, 2, 3], i != 2]
    [1, 6]

### build and test

    $ cd comp && export GOPATH=$GOPATH:$(pwd)
    $ go tool yacc -o src/comp/y.go -p "comp_" src/comp/grammar.y
    $ go test comp
    $ go install comp

### acknowledgements

comp language borrows ideas from other programming languages (Haskell,
JavaScript and probably others), but its core - the application of
comprehensions to formulate queries - is based on a research paper by Peter
Buneman "[Comprehension Syntax][1]". I explored this paper thanks to feedback
provided by [Ted Leung][2] after the [Emerging Languages Camp][3] regarding
[Bandicoot][4]. comp tool was originally developed as a backend to serve
public data sets at [mingle.io][5].

[0]: http://en.wikipedia.org/wiki/List_comprehension "List Comprehension"
[1]: http://citeseerx.ist.psu.edu/viewdoc/summary?doi=10.1.1.26.993 "Comprehension Syntax (1994), by Peter Buneman , Leonid Libkin, Dan Suciu, Val Tannen, Limsoon Wong"
[2]: http://www.sauria.com/blog/2012/09/27/strange-loop-2012/ "Ted Leung's Blog"
[3]: https://thestrangeloop.com/preconf "Emerging Languages Camp"
[4]: http://bandilab.org "Bandicoot"
[5]: https://mingle.io "mingle.io GmbH"
