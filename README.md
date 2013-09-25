comp is a tool for querying information from files. Its main goal is
to provide a unified interface to the variety of data representations found
in public data sets. To achieve this goal comp introduces a small query
language with type coercion (e.g. "2" + 2 == 4) and a powerful iteration
mechanism based on [list comprehensions][0].

    $ comp '"2" + 2'
    4
    $ comp '[i | i <- [1, 2, 3], i != 2]'
    [ 1, 3 ]
    $ comp '[i * j | i <- [1, 2, 3], j <- [10, 20]]'
    [ 10, 20, 20, 40, 30, 60 ]

Query the standard input:

    $ curl https://api.github.com/repos/torvalds/linux/commits > commits.json
    $ cat commits.json | comp -t json '[ i.commit.author.name | i <- in ]'

Preload files and run queries from the web console `http://localhost:9090/console` or via HTTP:

    $ comp -l :9090 -f commits.json,authors.csv
    $ curl -d '{"expr": "[ c | c <- contacts, c.zip == 8001]"}' http://localhost:9090/full

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

for example, the following expressions:

    "2.14" + 1
    8 / 2
    3 < 3.14 && 3.14 > 3.13
    "hello" ++ " world"

will produce:

    3.14
    4
    true
    "hello world"

Iterations are formulated as list comprehensions:

    [ e | g1, g2, ..., gN ]

pronounced as "a list of all e where g1 and g2 ... and gN", where `e`
represents an expression and `gX` is either an iteration over a list
or a boolean expression:

    [i*2 | i <- [1, 2, 3], i != 2]
    [i*j | i <- [1, 2, 3], j <- [10, 20]]

will produce:

    [1, 6]
    [10, 20, 20, 40, 30, 60]

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
