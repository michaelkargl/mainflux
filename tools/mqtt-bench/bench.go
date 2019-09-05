// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package bench

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/cisco/senml"
)

// Keep struct names exported, otherwise Viper unmarshaling won't work
type mqttBrokerConfig struct {
	URL string `toml:"url" mapstructure:"url"`
}

type mqttMessageConfig struct {
	Size    int    `toml:"size" mapstructure:"size"`
	Payload string `toml:"payload" mapstructure:"payload"`
	Format  string `toml:"format" mapstructure:"format"`
	QoS     int    `toml:"qos" mapstructure:"qos"`
	Retain  bool   `toml:"retain" mapstructure:"retain"`
}

type mqttTLSConfig struct {
	MTLS       bool   `toml:"mtls" mapstructure:"mtls"`
	SkipTLSVer bool   `toml:"skiptlsver" mapstructure:"skiptlsver"`
	CA         string `toml:"ca" mapstructure:"ca"`
}

type mqttConfig struct {
	Broker  mqttBrokerConfig  `toml:"broker" mapstructure:"broker"`
	Message mqttMessageConfig `toml:"message" mapstructure:"message"`
	TLS     mqttTLSConfig     `toml:"tls" mapstructure:"tls"`
}

type testConfig struct {
	Count int `toml:"count" mapstructure:"count"`
	Pubs  int `toml:"pubs" mapstructure:"pubs"`
	Subs  int `toml:"subs" mapstructure:"subs"`
}

type logConfig struct {
	Quiet bool `toml:"quiet" mapstructure:"quiet"`
}

type mainfluxFile struct {
	ConnFile string `toml:"connections_file" mapstructure:"connections_file"`
}

type mfThing struct {
	ThingID  string `toml:"thing_id" mapstructure:"thing_id"`
	ThingKey string `toml:"thing_key" mapstructure:"thing_key"`
	MTLSCert string `toml:"mtls_cert" mapstructure:"mtls_cert"`
	MTLSKey  string `toml:"mtls_key" mapstructure:"mtls_key"`
}

type mfChannel struct {
	ChannelID string `toml:"channel_id" mapstructure:"channel_id"`
}

type mainflux struct {
	Things   []mfThing   `toml:"things" mapstructure:"things"`
	Channels []mfChannel `toml:"channels" mapstructure:"channels"`
}

// Config struct holds benchmark configuration
type Config struct {
	MQTT mqttConfig   `toml:"mqtt" mapstructure:"mqtt"`
	Test testConfig   `toml:"test" mapstructure:"test"`
	Log  logConfig    `toml:"log" mapstructure:"log"`
	Mf   mainfluxFile `toml:"mainflux" mapstructure:"mainflux"`
}

// JSONResults are used to export results as a JSON document
type JSONResults struct {
	Runs   []*runResults `json:"runs"`
	Totals *totalResults `json:"totals"`
}

// Benchmark - main benckhmarking function
func Benchmark(cfg Config) {
	var wg sync.WaitGroup
	var err error

	//checkConnection(cfg.MQTT.Broker.URL, 1)
	subsResults := make(subsResults)
	var caByte []byte
	if cfg.MQTT.TLS.MTLS {
		caFile, err := os.Open(cfg.MQTT.TLS.CA)
		defer caFile.Close()
		if err != nil {
			fmt.Println(err)
		}

		caByte, _ = ioutil.ReadAll(caFile)
	}

	mf := mainflux{}
	if _, err := toml.DecodeFile(cfg.Mf.ConnFile, &mf); err != nil {
		log.Fatalf("Cannot load Mainflux connections config %s \nuse tools/provision to create file", cfg.Mf.ConnFile)
	}

	resCh := make(chan *runResults)
	finishPub := make(chan bool)
	doneSub := make(chan bool)
	startStamp := time.Now()

	n := len(mf.Channels)
	var cert tls.Certificate

	msg := buildSenML(cfg.MQTT.Message.Size, cfg.MQTT.Message.Payload)
	getSenML := func() senml.SenML {
		return msg
	}
	// Subscribers
	for i := 0; i < cfg.Test.Subs; i++ {
		mfConn := mf.Channels[i%n]
		mfThing := mf.Things[i%n]

		if cfg.MQTT.TLS.MTLS {
			cert, err = tls.X509KeyPair([]byte(mfThing.MTLSCert), []byte(mfThing.MTLSKey))
			if err != nil {
				log.Fatal(err)
			}
		}

		c := &Client{
			ID:         strconv.Itoa(i),
			BrokerURL:  cfg.MQTT.Broker.URL,
			BrokerUser: mfThing.ThingID,
			BrokerPass: mfThing.ThingKey,
			MsgTopic:   getTopic(mfConn.ChannelID, startStamp),
			MsgSize:    cfg.MQTT.Message.Size,
			MsgCount:   cfg.Test.Count,
			MsgQoS:     byte(cfg.MQTT.Message.QoS),
			Quiet:      cfg.Log.Quiet,
			MTLS:       cfg.MQTT.TLS.MTLS,
			SkipTLSVer: cfg.MQTT.TLS.SkipTLSVer,
			CA:         caByte,
			ClientCert: cert,
			Retain:     cfg.MQTT.Message.Retain,
			Message:    nil,
		}

		wg.Add(1)

		go c.runSubscriber(&wg, &subsResults, cfg.Test.Count*cfg.Test.Pubs, &finishPub, &doneSub)
	}

	wg.Wait()

	// Publishers
	start := time.Now()

	for i := 0; i < cfg.Test.Pubs; i++ {
		mfConn := mf.Channels[i%n]
		mfThing := mf.Things[i%n]

		if cfg.MQTT.TLS.MTLS {
			cert, err = tls.X509KeyPair([]byte(mfThing.MTLSCert), []byte(mfThing.MTLSKey))
			if err != nil {
				log.Fatal(err)
			}
		}

		c := &Client{
			ID:         strconv.Itoa(i),
			BrokerURL:  cfg.MQTT.Broker.URL,
			BrokerUser: mfThing.ThingID,
			BrokerPass: mfThing.ThingKey,
			MsgTopic:   getTopic(mfConn.ChannelID, startStamp),
			MsgSize:    cfg.MQTT.Message.Size,
			MsgCount:   cfg.Test.Count,
			MsgQoS:     byte(cfg.MQTT.Message.QoS),
			Quiet:      cfg.Log.Quiet,
			MTLS:       cfg.MQTT.TLS.MTLS,
			SkipTLSVer: cfg.MQTT.TLS.SkipTLSVer,
			CA:         caByte,
			ClientCert: cert,
			Retain:     cfg.MQTT.Message.Retain,
			Message:    getSenMLPayload,
			GetSenML:   getSenML,
		}

		go c.runPublisher(resCh)
	}

	// Collect the results
	var results []*runResults
	if cfg.Test.Pubs > 0 {
		results = make([]*runResults, cfg.Test.Pubs)
	}

	go func() {
		for i := 0; i < cfg.Test.Pubs; i++ {
			select {
			case result := <-resCh:
				{
					fmt.Printf("done, results prepared\n")

					results[i] = result
					if i == cfg.Test.Pubs-1 {
						fmt.Printf("Publishers finished %d\n", i)
						finishPub <- true
					}
				}
			}
		}
	}()

	finishSub := make(chan bool)
	go func() {
		for i := 0; i < cfg.Test.Subs; i++ {
			select {
			case <-doneSub:
				{
					fmt.Printf("done with subscribers %d\n", i)
					// every time subscriber receives MsgCount messages it will signal done
					if i == cfg.Test.Subs-1 {
						fmt.Printf("done with subscribers %d\n", i)
						finishSub <- true
						break
					}

				}
			}
		}
	}()
	<-finishPub
	<-finishSub

	totalTime := time.Now().Sub(start)
	totals := calculateTotalResults(results, totalTime, subsResults)
	if totals == nil {
		return
	}

	// Print sats
	printResults(results, totals, cfg.MQTT.Message.Format, cfg.Log.Quiet)
}

func buildSenML(sz int, payload string) senml.SenML {

	t := (float64)(time.Now().UnixNano())
	timeStamp := senml.SenMLRecord{
		BaseName: "pub-2019-08-31T12:38:25.139715762+02:00-57",
		Value:    &t,
	}

	tsByte, err := json.Marshal(timeStamp)
	if err != nil || len(payload) == 0 {
		log.Fatalf("Failed to create test message")
	}

	sml := senml.SenMLRecord{}
	err = json.Unmarshal([]byte(payload), &sml)
	if err != nil {
		log.Fatalf("Cannot unmarshal payload")
	}

	msgByte, err := json.Marshal(sml)
	if err != nil {
		log.Fatalf("Failed to create test message")
	}

	// how many records to make message long sz bytes
	n := (sz-len(tsByte))/len(msgByte) + 1
	if sz < len(tsByte) {
		n = 1
	}

	records := make([]senml.SenMLRecord, n)
	records[0] = timeStamp
	for i := 1; i < n; i++ {
		// is this needed
		// i think we need id to be saved with db writer to t
		sml.Time = float64(time.Now().UnixNano())
		records[i] = sml
	}

	s := senml.SenML{
		Records: records,
	}

	return s
}

func getSenMLPayload(cid string, time float64, getSenML func() senml.SenML) ([]byte, error) {
	s := getSenML()
	s.Records[0].Value = &time
	s.Records[0].BaseName = cid
	payload, err := senml.Encode(s, senml.JSON, senml.OutputOptions{})
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func getTopic(ch string, start time.Time) string {
	return fmt.Sprintf("channels/%s/messages/%d/test", ch, start.UnixNano())
}
