package genesys

import (
	"fmt"
	"strconv"
	"time"

	"github.com/mypurecloud/platform-client-sdk-go/v188/platformclientv2"
)

func getAllflows(config *platformclientv2.Configuration, pageNumber int) ([]platformclientv2.Flow, error) {
	var allFlows []platformclientv2.Flow
	pageSize := 200
	apiInstance := platformclientv2.NewArchitectApiWithConfig(config)

	for {
		data, _, err := apiInstance.GetFlows([]string{}, pageNumber, pageSize, "", "", []string{}, "", "", "", "", "", "", "", "", false, false, false, "", "", []string{})
		if err != nil {
			return nil, err
		}
		if data.Entities == nil || len(*data.Entities) == 0 {
			break
		}
		for _, flow := range *data.Entities {
			if flow.Id != nil {
				allFlows = append(allFlows, flow)
			}
		}
		if len(*data.Entities) < pageSize {
			break
		}
		pageNumber++
	}
	return allFlows, nil
}

func GetFlows(config *platformclientv2.Configuration, startDate, endDate, timeZone string) (message string, flowsTable [][]string, err error) {
	allFlows, err := getAllflows(config, 1)
	if err != nil {
		return fmt.Sprintf("Error retrieving flows: %v\n", err), nil, err
	}

	jobID, err := getUsedFlowsCreateJob(config, startDate, endDate, timeZone)
	if err != nil {
		return fmt.Sprintf("Error creating job: %v\n", err), nil, err
	}

	jobStatus, err := getUsedFlowsGetJobStatus(config, jobID)
	if err != nil {
		return fmt.Sprintf("Error getting job status: %v\n", err), nil, err
	}

	if jobStatus == "FULFILLED" {
		usedFlowsResults, err := getUsedFlowsGetJobResults(config, jobID, "")
		if err != nil {
			return fmt.Sprintf("Error getting job results: %v\n", err), nil, err
		}
		flowsTable, err := formatFlowsResults(usedFlowsResults, allFlows)
		if err != nil {
			return fmt.Sprintf("Error formatting results: %v\n", err), nil, err
		}
		return "Completed GetUsedFlows", flowsTable, nil
	} else {
		return fmt.Sprintf("Unexpected job status: %s\n", jobStatus), nil, fmt.Errorf("unexpected job status: %s", jobStatus)
	}
}

func getUsedFlowsCreateJob(config *platformclientv2.Configuration, startDate, endDate, timeZone string) (string, error) {
	utcStartDate, err := convertLocalToUTCStr(startDate, timeZone)
	if err != nil {
		return "", fmt.Errorf("error converting start date to UTC: %v", err)
	}
	utcEndDate, err := convertLocalToUTCStr(endDate, timeZone)
	if err != nil {
		return "", fmt.Errorf("error converting end date to UTC: %v", err)
	}
	intervalStr := utcStartDate + "/" + utcEndDate
	apiInstance := platformclientv2.NewAnalyticsApiWithConfig(config)
	data, _, err := apiInstance.PostAnalyticsFlowsAggregatesJobs(platformclientv2.Flowasyncaggregationquery{
		Interval: &intervalStr,
		TimeZone: new(timeZone),
		GroupBy:  &[]string{"flowId"},
		Metrics:  &[]string{"tFlow"},
	})
	if err != nil {
		return "", err
	} else {
		return *data.JobId, nil
	}
}

func getUsedFlowsGetJobStatus(config *platformclientv2.Configuration, jobID string) (string, error) {
	apiInstance := platformclientv2.NewAnalyticsApiWithConfig(config)
	for {
		data, _, err := apiInstance.GetAnalyticsFlowsAggregatesJob(jobID)
		if err != nil {
			return "", err
		}
		state := *data.State

		switch state {
		case "FULFILLED":
			return "FULFILLED", nil
		case "FAILED":
			return "FAILED", fmt.Errorf("job failed on server")
		default:
			time.Sleep(2 * time.Second)
		}
	}
}

func getUsedFlowsGetJobResults(config *platformclientv2.Configuration, jobID, cursor string) ([]platformclientv2.Flowaggregatedatacontainer, error) {
	apiInstance := platformclientv2.NewAnalyticsApiWithConfig(config)
	var allResults []platformclientv2.Flowaggregatedatacontainer
	for {
		data, _, err := apiInstance.GetAnalyticsFlowsAggregatesJobResults(jobID, cursor)
		if err != nil {
			return nil, err
		}
		if data.Results != nil {
			allResults = append(allResults, *data.Results...)
		}
		if data.Cursor == nil || *data.Cursor == "" {
			break
		}
		cursor = *data.Cursor
	}
	return allResults, nil
}

func formatFlowsResults(results []platformclientv2.Flowaggregatedatacontainer, allFlows []platformclientv2.Flow) ([][]string, error) {
	table := []UsageTable{}

	table = append(table, UsageTable{Type: "WORKFLOW", CX1: "🔴", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "DIGITALBOT", CX1: "", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "INBOUNDCALL", CX1: "🔴", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "INBOUNDCHAT (EOL)", CX1: "", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "INQUEUECALL", CX1: "🔴", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "SURVEYINVITE", CX1: "", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "VOICESURVEY", CX1: "🔴", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "INBOUNDEMAIL", CX1: "", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "BOT", CX1: "🔴", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "SECURECALL", CX1: "🔴", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "COMMONMODULE", CX1: "🔴", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "INBOUNDSHORTMESSAGE", CX1: "", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "OUTBOUNDCALL", CX1: "🔴", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "WORKITEM", CX1: "🔴", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "INQUEUESHORTMESSAGE", CX1: "", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "INQUEUEEMAIL", CX1: "", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})
	table = append(table, UsageTable{Type: "VOICEMAIL", CX1: "🔴", CX2: "🔴", CX3: "🔴", BuiltCount: 0, UsedCount: 0, InUse: "⛔"})

	// Create a map for quick lookup of table rows by flow type
	tableMap := make(map[string]*UsageTable)
	for i := range table {
		tableMap[table[i].Type] = &table[i]
	}
	for _, flowExists := range allFlows {
		if found, exists := tableMap[*flowExists.VarType]; exists {
			found.BuiltCount++
			for _, flowUsed := range results {
				if *flowExists.Id == (*flowUsed.Group)["flowId"] {
					found.UsedCount++
				}
			}
		}
	}

	for i := range table {
		if table[i].UsedCount > 0 {
			table[i].InUse = "✅"
		}
	}

	// Create output table with headers
	tableData := [][]string{{"Type of flow", "CX1", "CX2", "CX3", "total built", "total used", "In use"}}
	for _, row := range table {
		tableData = append(tableData, []string{row.Type, row.CX1, row.CX2, row.CX3, strconv.Itoa(row.BuiltCount), strconv.Itoa(row.UsedCount), row.InUse})
	}

	return tableData, nil
}
