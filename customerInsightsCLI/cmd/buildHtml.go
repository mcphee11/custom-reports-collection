package genesys

import (
	"embed"
	"fmt"
	"os"
	"strings"
	"time"
)

//go:embed _htmlReport/*
var webIndexHTMLFile embed.FS

//buildColumnChart('incomingMediaTypes', 'Inbound Interactions', [{key: 'Voice', count: 23}, {key: 'Message', count: 11}, {key: 'Email', count: 5}, {key: 'Callback', count: 0}])
//buildTableRows('inboundTable', 'inboundTableName', [['Media', 'Interactions'], ['Voice', 23], ['Message', 11], ['Email', 5], ['Callback', 0]])

type ReportData struct {
	createdOn                         string
	organisationName                  string
	organisationID                    string
	region                            string
	startDate                         string
	endDate                           string
	totalAPIReq                       string
	Mos                               MOS
	Rfactor                           RFACTOR
	InboundConnectedInteractionsGraph inboundConnectedInteractionsGraph
	InboundConnectedInteractionsTable inboundConnectedInteractionsTable
}

type MOS struct{}
type RFACTOR struct{}
type inboundConnectedInteractionsGraph struct{}
type inboundConnectedInteractionsTable struct{}

func BuildIndexHTMLFile(
	orgID, orgName, region, startDate, endDate, timeZone,
	mosHTML, rFactorHTML string,
	inboundTotalChart, outboundTotalChart,
	inboundTotalTable, outboundTotalTable,
	recordingsChart, recordingTable,
	perDayVoiceChart, perDayMessageChart, perDayEmailChart, perDayCallbackChart,
	perDayUsersChart,
	flowsTable, routingTable, objectTable []byte) (string, error) {

	var CodeInsert = fmt.Sprintf(`
		document.getElementById('createdDate').innerHTML = 'Created on: <strong>%s</strong>'
		document.getElementById('orgId').innerHTML = '<strong>%s</strong>'
		document.getElementById('orgName').innerHTML = '<strong>%s</strong>'
		document.getElementById('region').innerHTML = '<strong>%s</strong>'
		document.getElementById('startDate').innerHTML = '<strong>%s</strong>'
		document.getElementById('endDate').innerHTML = '<strong>%s</strong>'
		document.getElementById('timeZone').innerHTML = '<strong>%s</strong>'
		document.getElementById('mosLabel').innerHTML = %s
		document.getElementById('rFactorLabel').innerHTML = %s
    buildColumnChart('incomingMediaTypes', 'Inbound Interactions', %s ) 
    buildColumnChart('outgoingMediaTypes', 'Outbound Interactions', %s )
    buildTableRows('inboundTable', 'inboundTableName', %s ) 
    buildTableRows('outboundTable', 'outboundTableName', %s )
    buildColumnChart('recordings', 'Recordings', %s ) 
    buildTableRows('recordingsTable', 'recordingsTableName', %s )
    buildColumnChart('callsPerDay', 'Calls Per Day', %s ) 
    buildColumnChart('messagesPerDay', 'Messages Per Day', %s ) 
    buildColumnChart('emailsPerDay', 'Emails Per Day', %s ) 
    buildColumnChart('callbacksPerDay', 'Callbacks Per Day', %s ) 
		buildStackedColumnChart(%s, %s, %s, %s)
		buildColumnChart('usersPerDay', 'Users Interacting Per Day', %s )
		buildTableRows('flowUsage', 'flowUsageName', %s )
		//buildTableRows('routingUsage', 'routingUsageName', %s)
		buildTableRows('objectUsage', 'objectUsageName', %s)
`, time.Now().Local(), orgID, orgName, region, startDate, endDate, timeZone,
		mosHTML, rFactorHTML,
		inboundTotalChart, outboundTotalChart,
		inboundTotalTable, outboundTotalTable,
		recordingsChart, recordingTable, perDayVoiceChart, perDayMessageChart, perDayEmailChart, perDayCallbackChart,
		perDayVoiceChart, perDayMessageChart, perDayEmailChart, perDayCallbackChart,
		perDayUsersChart, flowsTable, routingTable, objectTable)

	index, err := webIndexHTMLFile.ReadFile("_htmlReport/Index.html")
	if err != nil {
		return "", err
	}
	formattedIndex := strings.ReplaceAll(string(index), "CODE_INSERT", CodeInsert)
	unix := time.Now().Unix()

	err = os.WriteFile(fmt.Sprintf("Index-%d.html", unix), []byte(formattedIndex), 0777)
	if err != nil {
		return "", err
	}
	return "Index.html reporting file has been created", nil
}
