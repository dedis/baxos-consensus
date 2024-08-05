package main

import (
	"baxos/common"
	"baxos/replica/src"
	"flag"
	"fmt"
	"os"
)

func main() {
	name := flag.Int64("name", 1, "name of the replica as specified in the local-configuration.yml")
	configFile := flag.String("config", "configuration/local-configuration.yml", "configuration file")
	logFilePath := flag.String("logFilePath", "logs/", "log file path")
	batchSize := flag.Int("batchSize", 50, "batch size")
	batchTime := flag.Int("batchTime", 5000, "maximum time to wait for collecting a batch of requests in micro seconds")
	debugOn := flag.Bool("debugOn", false, "false or true")
	isAsync := flag.Bool("isAsync", false, "false or true to simulate asynchrony")
	debugLevel := flag.Int("debugLevel", 0, "debug level")
	roundTripTime := flag.Int("roundTripTime", 2000, "round trip time in micro seconds")
	keyLen := flag.Int("keyLen", 8, "key length")
	valLen := flag.Int("valLen", 8, "value length")
	benchmarkMode := flag.Int("benchmarkMode", 0, "0: resident store, 1: redis")
	asyncTimeout := flag.Int("asyncTimeout", 500, "artificial asynchronous timeout in milli seconds")
	timeEpochSize := flag.Int("timeEpochSize", 500, "duration of a time epoch for the attacker in milli seconds")

	flag.Parse()

	cfg, err := common.NewInstanceConfig(*configFile, *name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		panic(err)
	}

	rp := src.New(int32(*name), cfg, *logFilePath, *batchSize, *batchTime, *debugOn, *debugLevel, *benchmarkMode, *keyLen, *valLen, *isAsync, *asyncTimeout, *timeEpochSize, int64(*roundTripTime))

	rp.WaitForConnections()
	rp.Run() // this is run in main thread

}
