// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package bench

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	mat "gonum.org/v1/gonum/mat"
	stat "gonum.org/v1/gonum/stat"
)

type runResults struct {
	ID             string  `json:"id"`
	successes      int64   `json:"successes"`
	failures       int64   `json:"failures"`
	runTime        float64 `json:"run_time"`
	msgTimeMin     float64 `json:"msg_time_min"`
	msgTimeMax     float64 `json:"msg_time_max"`
	msgTimeMean    float64 `json:"msg_time_mean"`
	msgTimeStd     float64 `json:"msg_time_std"`
	msgDelTimeMin  float64 `json:"msg_del_time_min"`
	msgDelTimeMax  float64 `json:"msg_del_time_max"`
	msgDelTimeMean float64 `json:"msg_del_time_mean"`
	msgDelTimeStd  float64 `json:"msg_del_time_std"`
	msgsPerSec     float64 `json:"msgs_per_sec"`
}

type subTimes map[string][]float64

type totalResults struct {
	Ratio             float64 `json:"ratio"`
	Successes         int64   `json:"successes"`
	Failures          int64   `json:"failures"`
	TotalRunTime      float64 `json:"total_run_time"`
	AvgRunTime        float64 `json:"avg_run_time"`
	MsgTimeMin        float64 `json:"msg_time_min"`
	MsgTimeMax        float64 `json:"msg_time_max"`
	MsgDelTimeMin     float64 `json:"msg_del_time_min"`
	MsgDelTimeMax     float64 `json:"msg_del_time_max"`
	MsgTimeMeanAvg    float64 `json:"msg_time_mean_avg"`
	MsgTimeMeanStd    float64 `json:"msg_time_mean_std"`
	MsgDelTimeMeanAvg float64 `json:"msg_del_time_mean_avg"`
	MsgDelTimeMeanStd float64 `json:"msg_del_time_mean_std"`
	TotalMsgsPerSec   float64 `json:"total_msgs_per_sec"`
	AvgMsgsPerSec     float64 `json:"avg_msgs_per_sec"`
}

func calculateTotalResults(results []*runResults, totalTime time.Duration, subTimes *subTimes) *totalResults {
	if results == nil || len(results) < 1 {
		return nil
	}
	totals := new(totalResults)
	subTimeRunResults := runResults{}
	msgTimeMeans := make([]float64, len(results))
	msgTimeMeansDelivered := make([]float64, len(results))
	msgsPerSecs := make([]float64, len(results))
	runTimes := make([]float64, len(results))
	bws := make([]float64, len(results))

	totals.totalRunTime = totalTime.Seconds()

	totals.msgTimeMin = results[0].msgTimeMin
	for i, res := range results {
		if len(*subTimes) > 0 {
			times := mat.NewDense(1, len((*subTimes)[res.ID]), (*subTimes)[res.ID])

			subTimeRunResults.msgTimeMin = mat.Min(times)
			subTimeRunResults.msgTimeMax = mat.Max(times)
			subTimeRunResults.msgTimeMean = stat.Mean((*subTimes)[res.ID], nil)
			subTimeRunResults.msgTimeStd = stat.StdDev((*subTimes)[res.ID], nil)

		}
		res.msgDelTimeMin = subTimeRunResults.msgTimeMin
		res.msgDelTimeMax = subTimeRunResults.msgTimeMax
		res.msgDelTimeMean = subTimeRunResults.msgTimeMean
		res.msgDelTimeStd = subTimeRunResults.msgTimeStd

		totals.successes += res.successes
		totals.failures += res.failures
		totals.totalMsgsPerSec += res.msgsPerSec

		if res.msgTimeMin < totals.msgTimeMin {
			totals.msgTimeMin = res.msgTimeMin
		}

		if res.msgTimeMax > totals.msgTimeMax {
			totals.msgTimeMax = res.msgTimeMax
		}

		if subTimeRunResults.msgTimeMin < totals.msgDelTimeMin {
			totals.msgDelTimeMin = subTimeRunResults.msgTimeMin
		}

		if subTimeRunResults.msgTimeMax > totals.msgDelTimeMax {
			totals.msgDelTimeMax = subTimeRunResults.msgTimeMax
		}

		msgTimeMeansDelivered[i] = subTimeRunResults.msgTimeMean
		msgTimeMeans[i] = res.msgTimeMean
		msgsPerSecs[i] = res.msgsPerSec
		runTimes[i] = res.runTime
		bws[i] = res.msgsPerSec
	}

	totals.ratio = float64(totals.successes) / float64(totals.successes+totals.failures)
	totals.avgMsgsPerSec = stat.Mean(msgsPerSecs, nil)
	totals.avgRunTime = stat.Mean(runTimes, nil)
	totals.msgDelTimeMeanAvg = stat.Mean(msgTimeMeansDelivered, nil)
	totals.msgDelTimeMeanStd = stat.StdDev(msgTimeMeansDelivered, nil)
	totals.msgTimeMeanAvg = stat.Mean(msgTimeMeans, nil)
	totals.msgTimeMeanStd = stat.StdDev(msgTimeMeans, nil)

	return totals
}

func printResults(results []*runResults, totals *totalResults, format string, quiet bool) {
	switch format {
	case "json":
		jr := JSONResults{
			Runs:   results,
			Totals: totals,
		}
		data, err := json.Marshal(jr)
		if err != nil {
			log.Printf("Failed to prepare results for printing - %s\n", err.Error())
		}
		var out bytes.Buffer
		json.Indent(&out, data, "", "\t")

		fmt.Println(string(out.Bytes()))
	default:
		if !quiet {
			for _, res := range results {
				fmt.Printf("======= CLIENT %s =======\n", res.ID)
				fmt.Printf("Ratio:               %.3f (%d/%d)\n", float64(res.successes)/float64(res.successes+res.failures), res.successes, res.successes+res.failures)
				fmt.Printf("Runtime (s):         %.3f\n", res.runTime)
				fmt.Printf("Msg time min (us):   %.3f\n", res.msgTimeMin)
				fmt.Printf("Msg time max (us):   %.3f\n", res.msgTimeMax)
				fmt.Printf("Msg time mean (us):  %.3f\n", res.msgTimeMean)
				fmt.Printf("Msg time std (us):   %.3f\n", res.msgTimeStd)

				fmt.Printf("Bandwidth (msg/sec): %.3f\n\n", res.msgsPerSec)
			}
		}
		fmt.Printf("========= TOTAL (%d) =========\n", len(results))
		fmt.Printf("Total Ratio:                 %.3f (%d/%d)\n", totals.ratio, totals.successes, totals.successes+totals.failures)
		fmt.Printf("Total Runtime (sec):         %.3f\n", totals.totalRunTime)
		fmt.Printf("Average Runtime (sec):       %.3f\n", totals.avgRunTime)
		fmt.Printf("Msg time min (us):           %.3f\n", totals.msgTimeMin)
		fmt.Printf("Msg time max (us):           %.3f\n", totals.msgTimeMax)
		fmt.Printf("Msg time mean mean (us):     %.3f\n", totals.msgTimeMeanAvg)
		fmt.Printf("Msg time mean std (us):      %.3f\n", totals.msgTimeMeanStd)

		fmt.Printf("Average Bandwidth (msg/sec): %.3f\n", totals.avgMsgsPerSec)
		fmt.Printf("Total Bandwidth (msg/sec):   %.3f\n", totals.totalMsgsPerSec)
	}
	return
}
