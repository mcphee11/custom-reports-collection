package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
	"os"
	"strconv"
	"sync"
	"time"
	_ "time/tzdata"

	genesys "tview/cmd"
)

func main() {
	fmt.Println(`
	8888888      888888                 8   8   8                   
	8  8  8 eeee 8    8 e   e eeee eeee 8   8   8 eeeee eeeee  eeee 
	8  8  8 8  8 8eeee8 8   8 8    8    8   8   8 8   8 8   8  8    
	8  8  8 8    8      8eee8 8eee 8eee 8   8   8 8eee8 8eee8e 8eee 
	8  8  8 8    8      8   8 8    8    8   8   8 8   8 8    8 8   
	8  8  8 8ee8 8      8   8 8eee 8eee 8eee8eee8 8   8 8    8 8eee 
																																 `)

	clientID := flag.String("clientId", "", "ClientID of the OAuth")
	secret := flag.String("secret", "", "Secret of the OAuth")
	startDate := flag.String("startDate", "", "Start date in YYYY-MM-DDTHH:mm:ss")
	endDate := flag.String("endDate", "", "Start date in YYYY-MM-DDTHH:mm:ss")
	region := flag.String("region", "", "Region of the Org eg: mypurecloud.com.au")
	timeZone := flag.String("timeZone", "", "TimeZone in ANIA format eg: Australia/Melbourne")
	version := flag.Bool("v", false, "Version of McPheeWare CLI")

	flag.Parse()

	if *version {
		logText("version", "0.1.1")
		os.Exit(0)
	}

	err := genesys.CheckForEnv(*region, *clientID, *secret, *startDate, *endDate, *timeZone)
	if err != "" {
		logText("CheckForEnv", err)
	}

	run(*region, *clientID, *secret, *startDate, *endDate, *timeZone)
}

func run(region, clientID, secret, startDate, endDate, timeZone string) {
	config, err := genesys.GenesysAuth(region, clientID, secret)
	if err != nil {
		logText("GenesysAuth", err.Error())
		return
	}
	logText("GenesysAuth", "Logged in to Genesys Cloud")

	var wg sync.WaitGroup

	// Declare result variables with exact types for GO routines
	var (
		orgID, orgName       string
		mosHTML, rFactorHTML string
		inboundTotalsChart   []byte
		outboundTotalsChart  []byte
		inboundTotalTable    []byte
		outboundTotalTable   []byte
		perDayVoiceChart     []genesys.Bar
		perDayMessageChart   []genesys.Bar
		perDayEmailChart     []genesys.Bar
		perDayCallbackChart  []genesys.Bar
		perDayUsersChart     []byte
		flowsTable           [][]string
		allQueues            []platformclientv2.Queue
		objectTableQueues    []string
		routingTable         [][]string
		mediaTypes           = []string{"voice", "message", "email", "callback"}
		recordingCounts      = make([]int, len(mediaTypes))
	)

	wg.Go(func() {
		msg, id, name, err := genesys.GetOrganization(config)
		logText("GetOrganization", msg)
		if err == nil {
			orgID, orgName = id, name
		}
	})

	wg.Go(func() {
		msg, mos, rFactor, _ := genesys.GetMosTotals(config, startDate, endDate, timeZone)
		logText("GetMosTotals", msg)
		mosHTML, rFactorHTML = mos, rFactor
	})

	wg.Go(func() {
		msg, inChart, outChart, inTable, outTable, _ := genesys.GetConversationTotals(config,
			startDate,
			endDate, timeZone)
		logText("GetConversationTotals", msg)
		inboundTotalsChart, outboundTotalsChart = inChart, outChart
		inboundTotalTable, outboundTotalTable = inTable, outTable
	})

	wg.Go(func() {
		msg, voice, msgChart, email, callback, _ := genesys.GetInteractionsPerDay(config,
			startDate,
			endDate, timeZone)
		logText("GetInteractionsPerDay", msg)
		perDayVoiceChart, perDayMessageChart = voice, msgChart
		perDayEmailChart, perDayCallbackChart = email, callback
	})

	// Get Recording Totals (4 media types)
	for i, mediaType := range mediaTypes {
		wg.Go(func() {
			msg, count, err := genesys.GetRecordingTotals(config, startDate, endDate, timeZone,
				mediaType)
			logText("GetRecordingTotals ("+mediaType+")", msg)
			if err == nil {
				recordingCounts[i] = count
			}
		})
	}

	wg.Go(func() {
		msg, users, _ := genesys.GetUsersPerDay(config, startDate, endDate, timeZone)
		logText("GetUsersPerDay", msg)
		perDayUsersChart = users
	})

	wg.Go(func() {
		msg, flows, _ := genesys.GetFlows(config, startDate, endDate, timeZone)
		logText("GetFlows", msg)
		flowsTable = flows
	})

	// Get Queues & Dependent Routing (Chained)
	wg.Go(func() {
		msgQueues, queues, objQueues, err := genesys.GetObjectQueues(config, startDate, endDate,
			timeZone)
		logText("GetObjectQueues", msgQueues)
		if err == nil {
			allQueues, objectTableQueues = queues, objQueues

			// GetRouting depends on allQueues
			msgRouting, routing, _ := genesys.GetRouting(config, allQueues, startDate, endDate,
				timeZone)
			logText("GetRouting", msgRouting)
			routingTable = routing
		}
	})

	wg.Wait()

	// Post-processing: Assemble recording results
	var recordingsChart []genesys.Bar
	recordingsTable := [][]string{{"MediaType", "Interactions"}}
	for i, mType := range mediaTypes {
		recordingsTable = append(recordingsTable, []string{mType,
			strconv.Itoa(recordingCounts[i])})
		recordingsChart = append(recordingsChart, genesys.Bar{Key: mType, Count: recordingCounts[i]})
	}

	objectTable := [][]string{{"Object", "Total Built", "Total Used", "In Use"}}
	if len(objectTableQueues) > 0 {
		objectTable = append(objectTable, objectTableQueues)
	}

	jsonRecordingsChart, _ := json.Marshal(recordingsChart)
	jsonRecordingsTable, _ := json.Marshal(recordingsTable)
	jsonPerDayVoiceChart, _ := json.Marshal(perDayVoiceChart)
	jsonPerDayMessageChart, _ := json.Marshal(perDayMessageChart)
	jsonPerDayEmailChart, _ := json.Marshal(perDayEmailChart)
	jsonPerDayCallbackChart, _ := json.Marshal(perDayCallbackChart)
	jsonFlowsTable, _ := json.Marshal(flowsTable)
	jsonRoutingTable, _ := json.Marshal(routingTable)
	jsonObjectTable, _ := json.Marshal(objectTable)

	// Build final HTML file
	htmlMsg, errBuild := genesys.BuildIndexHTMLFile(
		orgID, orgName, region, startDate, endDate, timeZone,
		mosHTML, rFactorHTML,
		inboundTotalsChart, outboundTotalsChart,
		inboundTotalTable, outboundTotalTable,
		jsonRecordingsChart, jsonRecordingsTable,
		jsonPerDayVoiceChart, jsonPerDayMessageChart, jsonPerDayEmailChart,
		jsonPerDayCallbackChart,
		perDayUsersChart,
		jsonFlowsTable, jsonRoutingTable, jsonObjectTable)

	if errBuild != nil {
		logText("BuildIndexHtmlFile", fmt.Sprint(errBuild))
	} else {
		logText("BuildIndexHtmlFile", htmlMsg)
	}
}

// ----------------- Helpers -------------------

func logText(source, message string) {
	fmt.Printf("%s: %s - %s\n", time.Now(), source, message)
}
