# dnstress

A simple dns query stress tool built on github.com/miekg/dns.

## build
```
go build -o dnstress cmd/main.go
```

## usage
```
dnstress version 1.0.0
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
        specify how a limit for how long to run tests in seconds
  -p int
        set the port on which to query the server (default 53)
  -q int
        specify the maximum number of queries outstanding
  -r int
        set RTT statistics array size (default 50000)
  -s string
        sets the server to query (default "127.0.0.1")
  -t int
        specify the timeout for query completion in seconds (default 5)
  -u int
        set RTT statistics top percentile value(1-99) (default 95)
  -v    show debug info
```

## output example

- set the maximum number of queries outstanding ： 1000

```
$ ./dnstress -d example.txt -s 8.8.8.8 -c 50 -q 1000 
Statistics:
+----------------------+---------------------+
| Queries total        | 1000                |
| Queries succeeded    | 845                 |
| Queries failed       | 155                 |
| Queries success rate | 84.50%              |
| Queries QPS          | 49                  |
| Queries RTT TP95     | 54ms                |
| Queries started at   | 2024-10-27 20:50:55 |
| Queries finished at  | 2024-10-27 20:51:15 |
| Queries elapsed      | 20s                 |
+----------------------+---------------------+

```

- show real-time progress
```
$ ./dnstress -s 8.8.8.8 -c 50 -q 1000 -d example.txt -g
[progress] total:606, succeed:560, failed:46, success rate:92.41%, elapsed:6s, qps:100
...

```

- show debug info
```
$ ./dnstress -s 8.8.8.8 -c 50 -q 1000 -d example.txt -v   

[debug] load query items: 28
[debug] query: reddit.com(A) RTT: 39ms
[debug] query: facebook.com(NS) RTT: 40ms
[debug] query: google.com.hk(A) RTT: 41ms
[debug] query: youtube.com(AAAA) RTT: 42ms
...
```
