MQTT benchmarking tool
=========
A simple MQTT (broker) benchmarking tool for Mainflux platform. ( based on github.com/krylovsk/mqtt-benchmark )


The tool supports multiple concurrent clients, publishers and subscribers configurable message size, etc:

```
cd benchmark
go build  -o mqtt-benchmark *.go

> mqtt-benchmark --help
Usage of mqtt-benchmark:
  -broker="tcp://localhost:1883": MQTT broker endpoint as scheme://host:port
  -clients=10: Number of clients to start
  -count=100: Number of messages to send per client
  -format="text": Output format: text|json
  -password="": MQTT password (empty if auth disabled)
  -qos=1: QoS for published messages
  -quiet=false : Suppress logs while running (except errors and the result)
  -size=100: Size of the messages payload (bytes
  -subs=10 number of subscribers
  -pubs=10 number of publishers
  -config=connections.json , file with mainflux channels
  -mtls=false, use mtls
  -ca=ca.crt, use mqtts, pass ca to server validate certificate
```

Two output formats supported: human-readable plain text and JSON.

Before use you need a connections.json file with channels and credentials
```

[
  {
    "ChannelID": "d07a94dd-1e5c-4ead-b1d7-0f178afb471b",
    "ThingID": "652d6cb0-ed3c-4e6f-8512-312f614f3a27",
    "ThingKey": "efc32e8b-2641-4342-b5f6-f7ff77b67097",
    "MtlsCert"  : "-----BEGIN CERTIFICATE-----\nMIIEVTCCAz0CFCcgfqlRnT0lLgivo40jvyko/V4iMA0GCSqGSIb3DQEBCwUAMFcx..V+WGI+d2GImT\nqAA44O0M0Ovc\n-----END CERTIFICATE-----\n",
    "MtlsKey"   : "-----BEGIN PRIVATE KEY-----\nMIIJQwIBAD...S+SHkbHPKGO+O2Y=\n-----END PRIVATE KEY-----\n"

  },
  ....
]
```
Example use and output:

```
go build -o mqtt-benchmark *.go

without mtls
./mqtt-benchmark --broker tcp://localhost:1883 --count 100 --size 100  --qos 0 --format text   --subs 100 --pubs 0 --config connections.json

....

======= CLIENT 27 =======
Ratio:               1 (100/100)
Runtime (s):         16.396
Msg time min (ms):   9.466
Msg time max (ms):   1880.769
Msg time mean (ms):  150.193
Msg time std (ms):   201.884
Bandwidth (msg/sec): 6.099

========= TOTAL (100) =========
Total Ratio:                 1 (10000/10000)
Total Runime (sec):          16.398
Average Runtime (sec):       15.514
Msg time min (ms):           7.766
Msg time max (ms):           2034.076
Msg time mean mean (ms):     140.751
Msg time mean std (ms):      13.695
Average Bandwidth (msg/sec): 6.761
Total Bandwidth (msg/sec):   676.112


```
