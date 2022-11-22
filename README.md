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
        UDP timeout (seconds, default 5s) (default 5)
  -workerNum int
        Number of simultaneous test workers to run (default 1)
```

## output example

- use alex top 1m as data file
```
 ./dnstress -srv 8.8.8.8 -port 53 -workerNum 10 -domainNum 2000 -timeout 10                               
got 648362 domains
stress...
result: Queries total:2000, succeed:1907, failed:93, success rate:95.35%, elapsed:124.62(s), QPS:16.05
```

- use specified data file
```

./dnstress -srv 8.8.8.8 -port 53 -workerNum 10 -domainNum 2000 -timeout 3 -tfile ./testdata/China.top500            
got 500 domains
stress...
result: Queries total:500, succeed:463, failed:37, success rate:92.60%, elapsed:19.26(s), QPS:25.96

```

- show debug info
```
./dnstress -srv 8.8.8.8 -port 53 -workerNum 10 -domainNum 200 -timeout 10 -tfile ./testdata/China.top500 -debug true
got 500 domains
stress...
worker[0]: query hao123.com.[A] err:read udp 192.168.21.152:59532->8.8.8.8:53: i/o timeout
worker[4]: query 2144.com.[A] err:read udp 192.168.21.152:54133->8.8.8.8:53: i/o timeout
...
result: Queries total:200, succeed:193, failed:7, success rate:96.50%, elapsed:14.70(s), QPS:13.61
```
