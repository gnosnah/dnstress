# dnstress

A simple DNS query stress tool built on [miekg/dns](https://github.com/miekg/dns).

## Build

```bash
go build -o dnstress ./cmd/
```

## Usage

Must set either `-q` (max queries) or `-l` (max duration), or both.

```
dnstress version 1.0.2
options:
  -D    set the DNSSEC OK bit (implies EDNS)
  -T string
        specify the default query type (default "A")
  -c int
        specify the number of concurrent queries (default 1)
  -d string
        specify the input data file (default "example.txt")
  -g    show real-time progress
  -h    show help
  -l int
        specify how long to run tests in seconds (0 = unlimited)
  -p int
        set the port on which to query the server (default 53)
  -q int
        specify the maximum number of queries to send (0 = unlimited)
  -r int
        set RTT statistics array size (default 50000)
  -s string
        sets the server to query (IP or hostname) (default "127.0.0.1")
  -t int
        specify the timeout for query completion in seconds (default 5)
  -u int
        set RTT statistics top percentile value(1-99) (default 95)
  -v    show debug info
```

Success means a DNS response with `NOERROR`. Transport errors and other RCODEs count as failed.

## Input file

Each non-comment line is: `domain [qtype]`

```
# comment
google.com A
youtube.com AAAA
example.com
```

If `qtype` is omitted, `-T` is used (default `A`).

## Examples

### Limit by query count

```bash
$ ./dnstress -d example.txt -s 8.8.8.8 -c 50 -q 1000

Statistics
╭──────────────┬─────────────────────╮
│ METRIC       │ VALUE               │
├──────────────┼─────────────────────┤
│ Total        │ 1000                │
│ Succeeded    │ 845                 │
│ Failed       │ 155                 │
│ Success rate │ 84.50%              │
│ QPS          │ 49                  │
│ RTT TP95     │ 54ms                │
│ Started      │ 2024-10-27 20:50:55 │
│ Finished     │ 2024-10-27 20:51:15 │
│ Elapsed      │ 20.0s               │
╰──────────────┴─────────────────────╯
```

### Real-time progress

```bash
$ ./dnstress -d example.txt -s 8.8.8.8 -c 50 -q 1000 -g
dnstress  8.8.8.8:53  workers=50  type=A  queries=1000
[████████░░░░░░░░░░░░░░░░]  42.5%  425/1000  ok 400  fail 25  94.1%  120 qps  3.2s
```

### Debug

```bash
$ ./dnstress -d example.txt -s 8.8.8.8 -c 50 -q 1000 -v
[debug] load query items: 28
[debug] query: reddit.com(A) RTT: 39ms
[debug] query: facebook.com(NS) RTT: 40ms
...
```
