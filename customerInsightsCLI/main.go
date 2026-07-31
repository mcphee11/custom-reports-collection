package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
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
		logText("version", "0.0.1")
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
	} else {

		logText("GenesysAuth", "Logged in to Genesys Cloud")

		// Get Organization Details
		msg0, orgID, orgName, err0 := genesys.GetOrganization(config)
		if err0 != nil {
			logText("GetOrganization", msg0)
		} else {
			logText("GetOrganization", msg0)
		}

		// Get Mos & rFactor
		msgMR, mosHTML, rFactorHTML, errMR := genesys.GetMosTotals(config, startDate, endDate, timeZone)
		if errMR != nil {
			logText("GetMosTotals", msgMR)
		} else {
			logText("GetMosTotals", msgMR)
		}

		// Get Interaction Totals per mediaType and direction
		msg1, inboundTotalsChart, outboundTotalsChart, inboundTotalTable, outboundTotalTable, err1 := genesys.GetConversationTotals(config, startDate, endDate, timeZone)
		if err1 != nil {
			logText("GetConversationTotals", msg1)
		} else {
			logText("GetConversationTotals", msg1)
		}
		// Get Interactions per Day by mediaType
		msg7, perDayVoiceChart, perDayMessageChart, perDayEmailChart, perDayCallbackChart, err2 := genesys.GetInteractionsPerDay(config, startDate, endDate, timeZone)
		if err2 != nil {
			logText("GetInteractionsPerDay", msg7)
		} else {
			logText("GetInteractionsPerDay", msg7)
		}
		// Get Total Recordings per MediaType
		var recordingsChart []genesys.Bar
		recordingsTable := [][]string{{"MediaType", "Interactions"}}
		msg3, results3, err3 := genesys.GetRecordingTotals(config, startDate, endDate, timeZone, "voice")
		if err3 != nil {
			logText("GetRecordingTotals", msg3)
		} else {
			logText("GetRecordingTotals", msg3)
			recordingsTable = append(recordingsTable, []string{"voice", strconv.Itoa(results3)})
			recordingsChart = append(recordingsChart, genesys.Bar{Key: "voice", Count: results3})
		}
		msg4, results4, err4 := genesys.GetRecordingTotals(config, startDate, endDate, timeZone, "message")
		if err4 != nil {
			logText("GetRecordingTotals", msg4)
		} else {
			logText("GetRecordingTotals", msg4)
			recordingsTable = append(recordingsTable, []string{"message", strconv.Itoa(results4)})
			recordingsChart = append(recordingsChart, genesys.Bar{Key: "message", Count: results4})
		}
		msg5, results5, err5 := genesys.GetRecordingTotals(config, startDate, endDate, timeZone, "email")
		if err5 != nil {
			logText("GetRecordingTotals", msg5)
		} else {
			logText("GetRecordingTotals", msg5)
			recordingsTable = append(recordingsTable, []string{"email", strconv.Itoa(results5)})
			recordingsChart = append(recordingsChart, genesys.Bar{Key: "email", Count: results5})
		}
		msg6, results6, err6 := genesys.GetRecordingTotals(config, startDate, endDate, timeZone, "callback")
		if err6 != nil {
			logText("GetRecordingTotals", msg6)
		} else {
			logText("GetRecordingTotals", msg6)
			recordingsTable = append(recordingsTable, []string{"callback", strconv.Itoa(results6)})
			recordingsChart = append(recordingsChart, genesys.Bar{Key: "callback", Count: results6})
		}

		// Get users per Day
		msg7, perDayUsersChart, err2 := genesys.GetUsersPerDay(config, startDate, endDate, timeZone)
		if err2 != nil {
			logText("GetUsersPerDay", msg7)
		} else {
			logText("GetUsersPerDay", msg7)
		}

		// Get flows used
		msg8, flowsTable, err8 := genesys.GetFlows(config, startDate, endDate, timeZone)
		if err8 != nil {
			logText("GetFlows", msg8)
		} else {
			logText("GetFlows", msg8)
		}

		// Get queues used
		msg10, allQueues, objectTableQueues, err10 := genesys.GetObjectQueues(config, startDate, endDate, timeZone)
		if err10 != nil {
			logText("GetQueuesTable", msg10)
		} else {
			logText("GetQueuesTable", msg10)
		}

		// Get Routing Used
		msg9, routingTable, err9 := genesys.GetRouting(config, allQueues, startDate, endDate, timeZone)
		if err9 != nil {
			logText("GetRouting", msg9)
		} else {
			logText("GetRouting", msg9)
		}

		// Create output table with headers
		objectTable := [][]string{{"Object", "Total Built", "Total Used", "In Use"}}
		objectTable = append(objectTable, objectTableQueues)

		jsonRecordingsChart, _ := json.Marshal(recordingsChart)
		jsonRecordingsTable, _ := json.Marshal(recordingsTable)
		jsonPerDayVoiceChart, _ := json.Marshal(perDayVoiceChart)
		jsonPerDayMessageChart, _ := json.Marshal(perDayMessageChart)
		jsonPerDayEmailChart, _ := json.Marshal(perDayEmailChart)
		jsonPerDayCallbackChart, _ := json.Marshal(perDayCallbackChart)
		jsonFlowsTable, _ := json.Marshal(flowsTable)
		jsonRoutingTable, _ := json.Marshal(routingTable)
		jsonObjectTable, _ := json.Marshal(objectTable)

		htmlMsg, err := genesys.BuildIndexHTMLFile(
			orgID, orgName, region, startDate, endDate, timeZone,
			mosHTML, rFactorHTML,
			inboundTotalsChart, outboundTotalsChart,
			inboundTotalTable, outboundTotalTable,
			jsonRecordingsChart, jsonRecordingsTable,
			jsonPerDayVoiceChart, jsonPerDayMessageChart, jsonPerDayEmailChart, jsonPerDayCallbackChart,
			perDayUsersChart,
			jsonFlowsTable, jsonRoutingTable, jsonObjectTable)
		if err != nil {
			logText("BuildIndexHtmlFile", fmt.Sprint(err))
		}
		logText("BuildIndexHtmlFile", htmlMsg)
	}
}

// ----------------- Helpers -------------------

func logText(source, message string) {
	fmt.Printf("%s: %s - %s\n", time.Now(), source, message)
}
