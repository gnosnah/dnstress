# dnstress

A simple dns query stress tool built on github.com/miekg/dns.

Use alexa top 1 million as query domain source. 

top1m: http://s3.amazonaws.com/alexa-static/top-1m.csv.zip

## build
```
go build -o dnstress
```

## usage
```
Simple DNS stress tool
Options:
  -debug
        Show debug info (default false)
  -domainNum int
        How many domain names to use in the test (default 1)
  -port string
        DNS server Port to test (default "53")
  -srv string
        DNS server IP address to stress (default "127.0.0.1")
  -timeout int
        UDP timeout (seconds, default 5s) (default 5)
  -workerNum int
        Number of simultaneous test workers to run (default 1)

```

## output example
```
./dnstress -srv 8.8.8.8 -port 53 -workerNum 4 -domainNum 2000 -timeout 3            
got 648362 domains
stress...
result: Queries total:2000, succeed:1988, failed:12,  success rate:0.9940, elapsed:59.433(s)
```

