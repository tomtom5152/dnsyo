# DNSYO
## AALLLLL THE DNS

DNSYO is a little tool to help keep track of DNS propagation, a Go port
of [YoSmudge/dnsyo](https://github.com/YoSmudge/dnsyo/).

In short, it's `dig`, if `dig` queried over 1000 servers and collated their results.

Here's what it does

    $ dnsyo -q 0 example.com

     - RESULTS
    I asked 1214 servers for A records related to example.com,
    456 responded with records and 758 gave errors
    Here are the results;


    427 servers responded with;
    93.184.216.34

    1 servers responded with;
    217.16.223.29

    1 servers responded with;
    93.184.216.34
    A 8 2 86400 20180212100405 20180122182707 30381 example.com. PH3QiFAbFvkjzGqG1CNRKy+DE1kf6S7WlYgMR4CosE4UtKfzm3q1RFZvBzeZODR4YiW4+OSZum3HRW7GoC404r2bbCyi+AZrxFjQmemvUQWyyEFLOREsMC9WPG85Ctp9Kzyoj1uL/98NVhcxA7Xpr1ZsTfA/Yt6ywvT2mKAn96I=

    26 servers responded with;
    127.0.0.1

    1 servers responded with;
    213.27.209.99


    And here are the errors;

    108 servers responded with;
    CONNECTION REFUSED

    79 servers responded with;
    NOANSWER

    4 servers responded with;
    NXDOMAIN

    9 servers responded with;
    SERVFAIL

    12 servers responded with;
    TIMEOUT

    546 servers responded with;
    REFUSED

Why go? Quite simply, speed. Go handles concurrency much better than Python,
resulting in lookups being significantly quicker.

## Installation

DNSYO requires [go dep](https://github.com/golang/dep) for dependency management.

For basic installation run

    go get github.com/tomtom5152/dnsyo

## Usage

For more information on the flags run `dnsyo help`

### Resolver list

DNSYO can query a master list of servers to determine the currently working servers from your location.
This is done using the `dnsyo update` command. For more usage information, run `dnsyo help update`.

When running an update, DNSYO will check for three known values, and allow up to one failure.

By default, DNSYO will pick 500 servers at random from it's list to query.
You can change this with the `--servers` or `-q` flag.
If you want DNSYO to query all the servers just pass `--servers=0` or `-q=0`.

### Record types

Just like `dig`, you can pass the record type with the `--type` flag, so to get Google's MX records just do

    dnsyo google.com --type MX

## Licence

DNSYO is released under the MIT licence, see `LICENCE.txt` for more info
