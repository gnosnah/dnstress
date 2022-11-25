# dnstress

A simple dns query stress tool built on github.com/miekg/dns.

Use specified data file or alexa top 1 million as query domain source. See **testdata** directory.

top1m: http://s3.amazonaws.com/alexa-static/top-1m.csv.zip

## build
```
go build -o dnstress
```

## usage
```
Simple DNS stress tool
Options:
  -dataDir string
    	Data file directory (default "testdata")
  -debug
    	Show debug info (default false)
  -domainNum int
    	How many domain names to use in the test (default 1)
  -port string
    	DNS server Port to test (default "53")
  -srv string
    	DNS server IP address to stress (default "127.0.0.1")
  -tfile string
    	tfile Specifies the input data file. If not specified, use alexa top 1 million as default
  -timeout int
    	UDP timeout(seconds) (default 5)
  -tp float
    	RTT top percentile (default 0.95)
  -workerNum int
    	Number of simultaneous test workers to run (default 1)
```

## output example

- use alex top 1m as data file
```
./dnstress -srv 8.8.8.8 -port 53 -workerNum 10 -domainNum 2000 -timeout 10
got 746076 domains
stress...
result: Queries total:2000, succeed:1986, failed:14, success rate:99.30%, elapsed:40.96(s), TP95:344(Milliseconds)
```

- use specified data file
```
./dnstress -srv 8.8.8.8 -port 53 -workerNum 10 -domainNum 2000 -timeout 3 -tfile ./testdata/China.top500
got 500 domains
stress...
result: Queries total:500, succeed:496, failed:4, success rate:99.20%, elapsed:9.01(s), TP95:235(Milliseconds)

```

- show debug info
```
./dnstress -srv 8.8.8.8 -port 53 -workerNum 10 -domainNum 200 -timeout 10 -tfile ./testdata/China.top500 -debug true
got 500 domains
stress...
worker[9]: query 4399.com.[A] rtt:10001069666, err:read udp 10.2.8.179:64435->8.8.8.8:53: i/o timeout
result: Queries total:200, succeed:199, failed:1, success rate:99.50%, elapsed:10.21(s), TP95:183(Milliseconds)
```
