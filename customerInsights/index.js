'use strict' //Enables strict mode is JavaScript
let url = new URL(document.location.href)
let gc_region = url.searchParams.get('gc_region')
let gc_clientId = url.searchParams.get('gc_clientId')
let gc_redirectUrl = url.searchParams.get('gc_redirectUrl')
let globalApiRequests = 0

//Getting and setting the GC details from dynamic URL and session storage
gc_region ? sessionStorage.setItem('gc_region', gc_region) : (gc_region = sessionStorage.getItem('gc_region'))
gc_clientId ? sessionStorage.setItem('gc_clientId', gc_clientId) : (gc_clientId = sessionStorage.getItem('gc_clientId'))
gc_redirectUrl ? sessionStorage.setItem('gc_redirectUrl', gc_redirectUrl) : (gc_redirectUrl = sessionStorage.getItem('gc_redirectUrl'))

let platformClient = require('platformClient')
const client = platformClient.ApiClient.instance
const capi = new platformClient.ConversationsApi()
const uapi = new platformClient.UsersApi()
const oapi = new platformClient.OrganizationApi()
const aapi = new platformClient.ArchitectApi()
const fapi = new platformClient.FlowsApi()
const rapi = new platformClient.RoutingApi()

// Configure Client App
const ClientApp = window.purecloud.apps.ClientApp
const myClientApp = new ClientApp({
  pcEnvironment: gc_region,
})
start()

async function start() {
  try {
    client.setEnvironment(gc_region)
    client.setPersistSettings(true, '_mm_')
    console.log('%cLogging in to Genesys Cloud', 'color: green')
    await client.loginPKCEGrant(gc_clientId, gc_redirectUrl, {})
    getUTCOffset()
    thisMonth()
  } catch (err) {
    console.log('Error: ', err)
  }
}

function last30days() {
  let today = new Date()
  let aMonthAgo = new Date()
  aMonthAgo.setMonth(aMonthAgo.getMonth() - 1)
  // prettier-ignore
  document.getElementById('datepicker').value = `${aMonthAgo.getFullYear()}-${String(aMonthAgo.getMonth() + 1).padStart(2, '0')}-${String(aMonthAgo.getDate()).padStart(2, '0')}/${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`
}

function thisMonth() {
  let today = new Date()
  // prettier-ignore
  document.getElementById('datepicker').value = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-01/${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`
}

function getUTCOffset() {
  const now = new Date()
  const offsetMinutes = now.getTimezoneOffset()
  const offsetHours = -offsetMinutes / 60 // Invert for standard UTC offset representation

  let offsetString = ''
  if (offsetHours === 0) {
    offsetString = '+00:00'
  } else {
    const sign = offsetHours > 0 ? '+' : '-'
    const absHours = Math.abs(Math.floor(offsetHours))
    const minutes = Math.abs(Math.floor((offsetHours - absHours) * 60))
    const hoursString = absHours.toString().padStart(2, '0')
    const minutesString = minutes.toString().padStart(2, '0')
    offsetString = `${sign}${hoursString}:${minutesString}`
  }
  document.getElementById('timeZone').value = offsetString
}

// save pdf
function buildPDF() {
  window.print()
}

async function clearData() {
  console.log('%cClearing data', 'color: green')
  let mm_center = document.getElementsByClassName('mm-center')
  let mm_chart = document.getElementsByClassName('mm-chart')
  let mm_table = document.getElementsByClassName('mm-table')
  for (const center of mm_center) {
    center.innerHTML = ''
  }
  for (const chart of mm_chart) {
    chart.innerHTML = ''
  }
  for (const table of mm_table) {
    table.innerHTML = ''
  }
  document.getElementById('createdDate').innerHTML = `Created on: `
  document.getElementById('orgName').innerHTML = ''
  document.getElementById('orgId').innerHTML = ''
  document.getElementById('region').innerHTML = ''
  document.getElementById('startDate').innerHTML = ''
  document.getElementById('endDate').innerHTML = ''
  document.getElementById('totalAPIs').innerHTML = ''

  document.getElementById('spinner').style.display = 'none'
  document.getElementById('download').style.display = 'none'
  globalApiRequests = 0
}

function notification(type, message) {
  if (window.location !== window.parent.location) {
    // if in an iframe
    myClientApp.alerting.showToastPopup(type, message)
    return
  }
  window.alert(message)
  return
}

// dynamic table creation
async function buildTableRows(location, name, rows) {
  for (const row of rows) {
    let tableBody = document.getElementById(name)
    if (!tableBody) {
      // create the top table row if its not already there
      console.log('Creating table')
      let top = document.getElementById(location)
      let guxTable = document.createElement('gux-table')
      let table = document.createElement('table')
      let header = document.createElement('thead')
      let headerRow = document.createElement('tr')
      let tbody = document.createElement('tbody')

      headerRow.setAttribute('data-row-id', 'head')
      tbody.setAttribute('id', name)
      table.setAttribute('slot', 'data')
      guxTable.setAttribute('resizable-columns', '')

      header.appendChild(headerRow)
      guxTable.appendChild(table)
      table.appendChild(header)
      table.appendChild(tbody)

      // create column names on first row
      for (const item of row) {
        let th = document.createElement('th')
        th.setAttribute('data-column-name', item)
        th.style.textWrap = 'auto'
        th.innerHTML = item
        th.title = item
        headerRow.appendChild(th)
      }
      top.appendChild(guxTable)
      continue
    }
    if (tableBody) {
      let counter = 0
      // add data to the row
      let tr = document.createElement('tr')
      for (const item of row) {
        let column = document.createElement('td')
        if (counter != 0) {
          column.style.textAlign = 'center'
        }
        if (item === '🔴') {
          column.style.fontSize = '.6em'
        }
        column.innerHTML = item
        tr.appendChild(column)
        counter++
      }
      tableBody.appendChild(tr)
    }
  }
}

async function setReportData() {
  let today = new Date()
  document.getElementById('createdDate').innerHTML = `Created on: <strong>${today}</strong>`
  const org = await oapi.getOrganizationsMe()
  globalApiRequests++
  document.getElementById('orgName').innerHTML = `<strong>${org.name}</strong>`
  document.getElementById('orgId').innerHTML = `<strong>${org.id}</strong>`
  document.getElementById('region').innerHTML = `<strong>${gc_region}</strong>`

  document.getElementById('startDate').innerHTML = `<strong>${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}</strong>`
  document.getElementById('endDate').innerHTML = `<strong>${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}</strong>`
}

async function getData() {
  console.log('%cGetting data', 'color: green')
  clearData()
  document.getElementById('spinner').style.display = 'block'
  try {
    await setReportData()
    await mosAndRFactor()

    let inbound = await mediaTypeTotals('inbound')
    console.log('%cInbound Media Types: ', 'color: green', inbound)
    // prettier-ignore
    buildColumnChart('incomingMediaTypes', 'Inbound Interactions', [{ key: 'Voice', count: inbound.voice?.length || 0 }, { key: 'Message', count: inbound.message?.length || 0 }, { key: 'Email', count: inbound.email?.length || 0 }, { key: 'Callback', count: inbound.callback?.length || 0 }])
    // prettier-ignore
    buildTableRows('inboundTable', 'inboundTableName', [['Media', 'Interactions'], ['Voice', inbound.voice?.length || 0], ['Message', inbound.message?.length || 0], ['Email', inbound.email?.length || 0], ['Callback', inbound.callback?.length || 0]])

    let outbound = await mediaTypeTotals('outbound')
    console.log('%cOutbound Media Types: ', 'color: green', outbound)
    // prettier-ignore
    buildColumnChart('outgoingMediaTypes', 'Outbound Interactions', [{ key: 'Voice', count: outbound.voice?.length || 0 }, { key: 'Message', count: outbound.message?.length || 0 }, { key: 'Email', count: outbound.email?.length || 0 }, { key: 'Callback', count: outbound.callback?.length || 0 }])
    // prettier-ignore
    buildTableRows('outboundTable', 'outboundTableName', [['Media', 'Interactions'], ['Voice', outbound.voice?.length || 0], ['Message', outbound.message?.length || 0], ['Email', outbound.email?.length || 0], ['Callback', outbound.callback?.length || 0]])

    let recordings = await mediaRecordingTotals()
    console.log('%cRecordings: ', 'color: green', recordings)
    // prettier-ignore
    buildColumnChart('recordings', 'Recordings', [{ key: 'Voice', count: recordings.voice?.length || 0 }, { key: 'Message', count: recordings.message?.length || 0 }, { key: 'Email', count: recordings.email?.length || 0 }, { key: 'Callback', count: recordings.callback?.length || 0 }])
    // prettier-ignore
    buildTableRows('recordingsTable', 'recordingsTableName', [['Media', 'Interactions'], ['Voice', recordings.voice?.length || 0], ['Message', recordings.message?.length || 0], ['Email', recordings.email?.length || 0], ['Callback', recordings.callback?.length || 0]])

    let callsPerDay = []
    let inboundCallsPerDay = await interactionsPerDay(inbound.voice, 'inboundVoice')
    let outboundCallsPerDay = await interactionsPerDay(outbound.voice, 'outboundVoice')
    let ci = 0
    for (const day of inboundCallsPerDay) {
      callsPerDay.push([day[0], day[1] + outboundCallsPerDay[ci][1]])
      ci++
    }
    console.log('%cCalls Per Day: ', 'color: green', callsPerDay)
    // prettier-ignore
    buildColumnChart('callsPerDay', 'Calls Per Day', callsPerDay.map((row) => ({ key: row[0], count: row[1] })))

    let emailsPerDay = []
    let inboundEmailsPerDay = await interactionsPerDay(inbound.email, 'inboundEmail')
    let outboundEmailsPerDay = await interactionsPerDay(outbound.email, 'outboundEmail')
    let ei = 0
    for (const day of inboundEmailsPerDay) {
      emailsPerDay.push([day[0], day[1] + outboundEmailsPerDay[ei][1]])
      ei++
    }
    console.log('%cEmails Per Day: ', 'color: green', emailsPerDay)
    // prettier-ignore
    buildColumnChart('emailsPerDay', 'Emails Per Day', emailsPerDay.map((row) => ({ key: row[0], count: row[1] })))

    let messagesPerDay = []
    let inboundMessagesPerDay = await interactionsPerDay(inbound.message, 'inboundMessaging')
    let outboundMessagesPerDay = await interactionsPerDay(outbound.message, 'outboundMessaging')
    let mi = 0
    for (const day of inboundMessagesPerDay) {
      messagesPerDay.push([day[0], day[1] + outboundMessagesPerDay[mi][1]])
      mi++
    }
    console.log('%cMessages Per Day: ', 'color: green', messagesPerDay)
    // prettier-ignore
    buildColumnChart('messagesPerDay', 'Messages Per Day', messagesPerDay.map((row) => ({ key: row[0], count: row[1] })))

    let callbacksPerDay = []
    let inboundCallbacksPerDay = await interactionsPerDay(inbound.callback, 'inboundCallback')
    let outboundCallbacksPerDay = await interactionsPerDay(outbound.callback, 'outboundCallback')
    let cbi = 0
    for (const day of inboundCallbacksPerDay) {
      callbacksPerDay.push([day[0], day[1] + outboundCallbacksPerDay[cbi][1]])
      cbi++
    }
    console.log('%cCallbacks Per Day: ', 'color: green', callbacksPerDay)
    // prettier-ignore
    buildColumnChart('callbacksPerDay', 'Callbacks Per Day', callbacksPerDay.map((row) => ({ key: row[0], count: row[1] })))

    // prettier-ignore
    buildStackedColumnChart(callsPerDay.map((row) => ({ key: row[0], count: row[1] })), emailsPerDay.map((row) => ({ key: row[0], count: row[1] })), messagesPerDay.map((row) => ({ key: row[0], count: row[1] })), callbacksPerDay.map((row) => ({ key: row[0], count: row[1] })))

    let usersPerDay = await getUsersPerDay()
    console.log('%cUsers Per Day: ', 'color: green', usersPerDay)
    // prettier-ignore
    buildColumnChart('usersPerDay', 'Users Per Day', usersPerDay.map((row) => ({ key: row[0], count: row[1] })))

    // Build flow usage
    const flows = await architectFlows()
    buildTableRows('flowUsage', 'flowUsageName', flows)

    // Build routing usage
    const queues = await getQueues(1)
    const routingTypes = await getRoutingTypes(queues)
    buildTableRows('routingUsage', 'routingUsageName', routingTypes)

    // Build capability usage
    await getQueuesUsage(queues)
    // Get Skills
    let skills = await getSkills(1)
    let skillsUsedInbdound = getSkillsBuilt(inbound)
    let skillsUsedOutbound = getSkillsBuilt(outbound)
    let uniqueueSkillIdsUsed = new Set(skillsUsedInbdound, skillsUsedOutbound)
    console.log(uniqueueSkillIdsUsed)
    let skillsInUse = '⛔'
    if (uniqueueSkillIdsUsed.size > 0) {
      skillsInUse = '✅'
    }
    let skillsTotals = [['Skills', skills.length, uniqueueSkillIdsUsed.size, skillsInUse]]
    buildTableRows('objectUsage', 'objectUsageName', skillsTotals)

    console.log(`Total API Requests: ${globalApiRequests}`)
    document.getElementById('totalAPIs').innerHTML = `<strong>${globalApiRequests}</strong>`
    document.getElementById('spinner').style.display = 'none'
    document.getElementById('download').style.display = 'block'
  } catch (err) {
    console.log('Error: ', err)
    notification('Error', `Error: ${err}`)
  }
}

async function getConversations(pageNumber, query) {
  query.paging.pageNumber = pageNumber
  let conversations = await capi.postAnalyticsConversationsDetailsQuery(query)
  globalApiRequests++
  if (pageNumber < Math.ceil(conversations.totalHits / 100)) {
    const nextConversations = await getConversations(pageNumber + 1, query)
    return conversations.conversations.concat(nextConversations)
  }
  return conversations.conversations
}

async function mosAndRFactor() {
  let conversations = await getConversations(1, {
    // prettier-ignore
    interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
    order: 'desc',
    orderBy: 'conversationStart',
    paging: {
      pageSize: 100,
      pageNumber: 1,
    },
    conversationFilters: [
      {
        predicates: [
          {
            type: 'dimension',
            operator: 'exists',
            dimension: 'mediaStatsMinConversationMos',
          },
        ],
        type: 'and',
      },
    ],
  })
  console.log('%cTotal MOS & RFactor Conversations: ', 'color: green', conversations)

  // rendering data
  let totalMosInteractions = 0
  let totalRFactorInteractions = 0
  let maxMos = 0
  let totalMos = 0
  let rFactorMax = 0
  let totalRFactor = 0
  let minMos = 4.99
  let rFactorMin = 99.99
  let mosColor = 'green'
  let rFactorColor = 'green'

  for (const conv of conversations) {
    // only if they have MOS data
    if (conv?.mediaStatsMinConversationMos) {
      ++totalMosInteractions
      let mos = Math.round(conv.mediaStatsMinConversationMos * 100) / 100
      if (mos < minMos) {
        minMos = mos
      }
      if (mos > maxMos) {
        maxMos = mos
      }
      totalMos += mos
    }
    // only if they have R-Factor data
    if (conv?.mediaStatsMinConversationRFactor) {
      ++totalRFactorInteractions
      let rFactor = Math.round(conv.mediaStatsMinConversationRFactor * 100) / 100
      if (rFactor < rFactorMin) {
        rFactorMin = rFactor
      }
      if (rFactor > rFactorMax) {
        rFactorMax = rFactor
      }
      totalRFactor += rFactor
    }
  }
  let averageMos = Math.round((totalMos / conversations.length) * 100) / 100
  let averageRFactor = Math.round((totalRFactor / conversations.length) * 100) / 100
  if (minMos < 3.5) {
    mosColor = 'red'
  }
  if (rFactorMin < 50) {
    rFactorColor = 'red'
  }
  // prettier-ignore
  document.getElementById('mosLabel').innerHTML = `<div style="display: inline-flex; width: 100%; justify-content: space-evenly;"><h5>MOS Total Interactions: ${totalMosInteractions}</h5><h5>Min: <strong style="color: ${mosColor}"> ▼ ${minMos}</strong></h5><h5>Max: <strong  style="color: green"> ▲ ${maxMos}</strong></h5><h5>Average: 📈 ${averageMos}</h5></div>`
  // prettier-ignore
  document.getElementById('rFactorLabel').innerHTML = `<div style="display: inline-flex; width: 100%; justify-content: space-evenly;"><h5>MOS Total Interactions: ${totalRFactorInteractions}</h5><h5>Min: <strong style="color: ${rFactorColor}"> ▼ ${rFactorMin}</strong></h5><h5>Max: <strong  style="color: green"> ▲ ${rFactorMax}</strong></h5><h5>Average: 📈 ${averageRFactor}</h5></div>`
}

async function buildColumnChart(destination, title, data) {
  new Chart(document.getElementById(destination), {
    type: 'bar',
    data: {
      labels: data.map((row) => row.key),
      datasets: [
        {
          label: title || 'Unknown',
          data: data.map((row) => row.count),
        },
      ],
    },
  })
}

async function buildStackedColumnChart(voice, email, message, callback) {
  let data = {
    labels: voice.map((row) => row.key), //['Date-1', 'Date-2', 'Date-3', 'Date-4', 'Date-5'],
    datasets: [
      {
        label: 'Call',
        data: voice.map((row) => row.count), //[10, 20, 15, 25, 30],
        backgroundColor: 'rgba(255, 99, 132, 0.6)', // Red with some transparency
        borderColor: 'rgba(255, 99, 132, 1)',
      },
      {
        label: 'Email',
        data: email.map((row) => row.count),
        backgroundColor: 'rgba(54, 162, 235, 0.6)', // Blue with some transparency
        borderColor: 'rgba(54, 162, 235, 1)',
      },
      {
        label: 'Message',
        data: message.map((row) => row.count),
        backgroundColor: 'rgba(255, 206, 86, 0.6)', // Yellow with some transparency
        borderColor: 'rgba(255, 206, 86, 1)',
      },
      {
        label: 'Callback',
        data: callback.map((row) => row.count),
        backgroundColor: 'rgba(75, 192, 192, 0.6)', // Green with some transparency
        borderColor: 'rgb(75, 192, 1)',
      },
    ],
  }
  new Chart(document.getElementById('interactionsPerDay'), {
    type: 'bar',
    data: data,
    options: {
      responsive: true,
      maintainAspectRatio: true,
      scales: {
        x: {
          stacked: true,
        },
        y: {
          stacked: true,
          beginAtZero: true,
        },
      },
      plugins: {
        title: {
          display: true,
          text: 'Interactions',
          font: {
            size: 18,
          },
        },
        tooltip: {
          mode: 'index',
          intersect: false,
        },
      },
    },
  })
}

async function mediaTypeTotals(direction) {
  // Voice
  let conversationsVoice = await getConversations(1, {
    // prettier-ignore
    interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
    order: 'desc',
    orderBy: 'conversationStart',
    paging: {
      pageSize: 100,
      pageNumber: 1,
    },
    conversationFilters: [
      {
        type: 'and',
        predicates: [
          {
            metric: 'nConnected',
            operator: 'exists',
          },
        ],
      },
    ],
    segmentFilters: [
      {
        predicates: [
          {
            dimension: 'mediaType',
            value: 'voice',
            operator: 'matches',
          },
          {
            dimension: 'direction',
            operator: 'matches',
            value: direction,
          },
        ],
        type: 'and',
      },
    ],
  })

  // Messaging
  let conversationsMessaging = await getConversations(1, {
    // prettier-ignore
    interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
    order: 'desc',
    orderBy: 'conversationStart',
    paging: {
      pageSize: 100,
      pageNumber: 1,
    },
    conversationFilters: [
      {
        type: 'and',
        predicates: [
          {
            metric: 'nConnected',
            operator: 'exists',
          },
        ],
      },
    ],
    segmentFilters: [
      {
        predicates: [
          {
            dimension: 'mediaType',
            value: 'message',
            operator: 'matches',
          },
          {
            dimension: 'direction',
            operator: 'matches',
            value: direction,
          },
        ],
        type: 'and',
      },
    ],
  })

  // Email
  let conversationsEmail = await getConversations(1, {
    // prettier-ignore
    interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
    order: 'desc',
    orderBy: 'conversationStart',
    paging: {
      pageSize: 100,
      pageNumber: 1,
    },
    conversationFilters: [
      {
        type: 'and',
        predicates: [
          {
            metric: 'nConnected',
            operator: 'exists',
          },
        ],
      },
    ],
    segmentFilters: [
      {
        predicates: [
          {
            dimension: 'mediaType',
            value: 'email',
            operator: 'matches',
          },
          {
            dimension: 'direction',
            operator: 'matches',
            value: direction,
          },
        ],
        type: 'and',
      },
    ],
  })

  // Callback
  let conversationsCallBack = await getConversations(1, {
    // prettier-ignore
    interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
    order: 'asc',
    orderBy: 'conversationStart',
    paging: {
      pageSize: 100,
      pageNumber: 1,
    },
    conversationFilters: [
      {
        type: 'and',
        predicates: [
          {
            metric: 'nConnected',
            operator: 'exists',
          },
        ],
      },
    ],
    segmentFilters: [
      {
        predicates: [
          {
            dimension: 'mediaType',
            value: 'callback',
            operator: 'matches',
          },
          {
            dimension: 'direction',
            operator: 'matches',
            value: direction,
          },
        ],
        type: 'and',
      },
    ],
  })

  return { voice: conversationsVoice, message: conversationsMessaging, email: conversationsEmail, callback: conversationsCallBack }
}

async function mediaRecordingTotals() {
  // Voice
  let conversationsVoice = await getConversations(1, {
    // prettier-ignore
    interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
    order: 'desc',
    orderBy: 'conversationStart',
    paging: {
      pageSize: 100,
      pageNumber: 1,
    },
    segmentFilters: [
      {
        predicates: [
          {
            dimension: 'mediaType',
            value: 'voice',
            operator: 'matches',
          },
          {
            dimension: 'recording',
            operator: 'exists',
          },
        ],
        type: 'and',
      },
    ],
  })

  // Messaging
  let conversationsMessaging = await getConversations(1, {
    // prettier-ignore
    interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
    order: 'desc',
    orderBy: 'conversationStart',
    paging: {
      pageSize: 100,
      pageNumber: 1,
    },
    segmentFilters: [
      {
        predicates: [
          {
            dimension: 'mediaType',
            value: 'message',
            operator: 'matches',
          },
          {
            dimension: 'recording',
            operator: 'exists',
          },
        ],
        type: 'and',
      },
    ],
  })

  // Email
  let conversationsEmail = await getConversations(1, {
    // prettier-ignore
    interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
    order: 'desc',
    orderBy: 'conversationStart',
    paging: {
      pageSize: 100,
      pageNumber: 1,
    },
    segmentFilters: [
      {
        predicates: [
          {
            dimension: 'mediaType',
            value: 'email',
            operator: 'matches',
          },
          {
            dimension: 'recording',
            operator: 'exists',
          },
        ],
        type: 'and',
      },
    ],
  })

  // Callback
  let conversationsCallBack = await getConversations(1, {
    // prettier-ignore
    interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
    order: 'desc',
    orderBy: 'conversationStart',
    paging: {
      pageSize: 100,
      pageNumber: 1,
    },
    segmentFilters: [
      {
        predicates: [
          {
            dimension: 'mediaType',
            value: 'callback',
            operator: 'matches',
          },
          {
            dimension: 'recording',
            operator: 'exists',
          },
        ],
        type: 'and',
      },
    ],
  })

  return { voice: conversationsVoice, message: conversationsMessaging, email: conversationsEmail, callback: conversationsCallBack }
}

async function interactionsPerDay(conversations, type) {
  let interactionsPerDay = []
  let startDate = document.getElementById('datepicker').value.split('/')[0]
  let endDate = document.getElementById('datepicker').value.split('/')[1]
  let current = new Date(startDate)
  let last = new Date(endDate)
  while (current <= last) {
    let day = current.toISOString().split('T')[0]
    interactionsPerDay.push([day, 0])
    current.setDate(current.getDate() + 1)
  }
  if (!conversations) {
    console.warn('No conversations supplied for: ', type)
    return interactionsPerDay
  }

  for (const conversation of conversations) {
    let date = new Date(conversation.conversationStart)
    // Convert to local time and format as YYYY-MM-DD
    let localDate = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
    let day = localDate.toISOString().split('T')[0]

    if (!interactionsPerDay.some((row) => row[0] === day)) {
      // do nothing TODO: need to consider async interactions here..
      //interactionsPerDay.push([day, 1])
      console.warn('different start date interaction: ', conversation.id)
    } else {
      let index = interactionsPerDay.findIndex((row) => row[0] === day)
      interactionsPerDay[index][1] += 1
    }
  }
  return interactionsPerDay.sort((a, b) => new Date(a[0]) - new Date(b[0]))
}

async function getUsers(pageNumber) {
  let users = await uapi.getUsers({ pageNumber: pageNumber, pageSize: 200, state: 'active' })
  globalApiRequests++
  if (pageNumber < Math.ceil(users.total / 200)) {
    const nextUsers = await getUsers(pageNumber + 1)
    return users.entities.concat(nextUsers)
  }
  return users.entities
}

async function getUsersPerDay() {
  let usersSearch = []
  let userIds = await getUsers(1)
  for (const user of userIds) {
    usersSearch.push({ dimension: 'userId', value: user.id })
  }
  if (usersSearch.length > 100) {
    // need to page as only 100 agents at a time can be checked
    let userChunks = []
    for (let i = 0; i < usersSearch.length; i += 100) {
      userChunks.push(usersSearch.slice(i, i + 100))
    }
    let users = []
    for (const chunk of userChunks) {
      let result = await uapi.postAnalyticsUsersAggregatesQuery({
        // prettier-ignore
        interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
        granularity: 'P1D',
        groupBy: ['userId'],
        metrics: ['tSystemPresence'],
        filter: {
          type: 'and',
          predicates: chunk,
        },
      })
      globalApiRequests++
      users = users.concat(result.results)
    }
    let returnUsers = await formatUsers(users)
    return returnUsers
  } else {
    let users = await uapi.postAnalyticsUsersAggregatesQuery({
      // prettier-ignore
      interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
      granularity: 'P1D',
      groupBy: ['userId'],
      metrics: ['tSystemPresence'],
      filter: {
        type: 'and',
        predicates: usersSearch,
      },
    })
    globalApiRequests++
    let returnUsers = await formatUsers(users.results)
    return returnUsers
  }
}

async function formatUsers(users) {
  let returnUsers = []
  let startDate = document.getElementById('datepicker').value.split('/')[0]
  let endDate = document.getElementById('datepicker').value.split('/')[1]
  let current = new Date(startDate)
  let last = new Date(endDate)
  while (current <= last) {
    let day = current.toISOString().split('T')[0]
    returnUsers.push([day, 0])
    current.setDate(current.getDate() + 1)
  }

  // count up users that have been in stat other then OFFLINE and ++
  for (const user of users) {
    for (const day of user.data) {
      let dateSpilt = new Date(day.interval.split('/')[0])
      // Convert to local time and format as YYYY-MM-DD
      let localDate = new Date(dateSpilt.getTime() - dateSpilt.getTimezoneOffset() * 60000)
      let localDay = localDate.toISOString().split('T')[0]
      let date = returnUsers.find((x) => x[0] == localDay)
      for (const metric of day.metrics) {
        if (metric.qualifier != 'OFFLINE') {
          date[1] = date[1] + 1
          continue
        }
      }
    }
  }
  return returnUsers
}

async function getFlows(pageNumber) {
  let flows = await aapi.getFlows({ pageSize: 200, pageNumber: pageNumber })
  globalApiRequests++
  if (pageNumber != flows.pageCount) {
    const nextFlows = await getFlows(pageNumber + 1)
    return flows.entities.concat(nextFlows)
  }
  return flows.entities
}

async function usedFlows() {
  let flows = await fapi.postAnalyticsFlowsAggregatesQuery({
    // prettier-ignore
    interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
    groupBy: ['flowId'],
    metrics: ['tFlow'],
  })
  globalApiRequests++
  return flows.results
}

async function architectFlows() {
  let typeTotals = [
    { type: 'WORKFLOW', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'DIGITALBOT', CX1: '', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'INBOUNDCALL', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'INBOUNDCHAT (EOL)', CX1: '', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'INQUEUECALL', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'SURVEYINVITE', CX1: '', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'VOICESURVEY', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'INBOUNDEMAIL', CX1: '', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'BOT', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'SECURECALL', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'COMMONMODULE', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'INBOUNDSHORTMESSAGE', CX1: '', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'OUTBOUNDCALL', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'WORKITEM', CX1: 'AddOn', CX2: 'AddOn', CX3: 'AddOn', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'INQUEUESHORTMESSAGE', CX1: '', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'INQUEUEEMAIL', CX1: '', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    { type: 'VOICEMAIL', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' },
    //{ type: 'SPEECH', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' }, // Not used
    //{ type: 'VOICE', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: 0, usedCount: 0, InUse: 'No' }, // Not used
  ]

  let allFlows = await getFlows(1)
  for (const flow of allFlows) {
    let found = typeTotals.find((x) => x.type === flow.type)
    if (found) {
      found.builtCount++
    }
  }

  let allUsedFlows = await usedFlows()
  for (const flow of allUsedFlows) {
    let found = allFlows.find((x) => x.id === flow.group.flowId)
    if (found) {
      let update = typeTotals.find((x) => x.type === found.type)
      if (update) {
        update.usedCount++
      }
    }
  }

  let tableData = [['Type of flow', 'CX1', 'CX2', 'CX3', 'total built', 'total used', 'In use']]
  for (const row of typeTotals) {
    let inUse = '⛔'
    if (row.usedCount > 0) {
      inUse = '✅'
    }
    tableData.push([row.type, row.CX1, row.CX2, row.CX3, row.builtCount.toString(), row.usedCount.toString(), inUse])
  }
  return tableData
}

async function getQueues(pageNumber) {
  let queues = await rapi.getRoutingQueues({ pageNumber: pageNumber, pageSize: 200 })
  globalApiRequests++
  if (pageNumber != queues.pageCount) {
    const nextQueues = await getQueues(pageNumber + 1)
    return queues.entities.concat(nextQueues)
  }
  return queues.entities
}

async function getQueuesUsage(totalQueues) {
  let interactions = await capi.postAnalyticsConversationsAggregatesQuery({
    interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value
      }`,
    groupBy: ['queueId'],
    metrics: ['nConnected'],
  })
  globalApiRequests++
  let totalInteractions = interactions.results

  let queuesReport = [['Object', 'total built', 'total used', 'In use']]
  let inUse = 'No'
  if (totalInteractions.length > 0) {
    inUse = '✅'
  }
  queuesReport.push(['Queues', totalQueues.length, totalInteractions.length, inUse])
  buildTableRows('objectUsage', 'objectUsageName', queuesReport)
}

async function getRoutingTypes(totalQueues) {
  console.log(totalQueues)
  let bullseye = 0,
    predective = 0,
    conditionalGroup = 0,
    prefferedAgent = 0,
    standard = 0
  for (const queue of totalQueues) {
    if (queue.bullseye) {
      bullseye++
      continue
    }
    if (queue.predictiveRouting) {
      predective++
      continue
    }
    if (queue.conditionalGroupRouting) {
      conditionalGroup++
      continue
    }
    if (queue.routingRules) {
      // This is due to UI issue where the array was never set...
      if (queue.routingRules[0]) {
        prefferedAgent++
        continue
      }
      standard++
    } else {
      standard++
    }
  }

  const bullseyeMethod = await getRoutingUsage('Bullseye')
  const predictiveMethod = await getRoutingUsage('Predictive')
  const conditionalMethod = await getRoutingUsage('Conditional')
  const preferredMethod = await getRoutingUsage('Preferred')
  const standardMethod = await getRoutingUsage('Standard')

  // Other non common methods...
  const directMethod = await getRoutingUsage('Direct')
  const lastMethod = await getRoutingUsage('Last')
  const manualMethod = await getRoutingUsage('Manual')
  const vipMethod = await getRoutingUsage('Vip')
  console.log(`Other methods: ${directMethod + lastMethod + manualMethod + vipMethod}`)

  let tableData = []
  let data = [
    { type: 'Bullseye', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: bullseye, usedCount: bullseyeMethod, InUse: '⛔' },
    { type: 'Predective', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: predective, usedCount: predictiveMethod, InUse: '⛔' },
    { type: 'Conditional Group', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: conditionalGroup, usedCount: conditionalMethod, InUse: '⛔' },
    { type: 'Preffered Agent', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: prefferedAgent, usedCount: preferredMethod, InUse: '⛔' },
    { type: 'Standard', CX1: '🔴', CX2: '🔴', CX3: '🔴', builtCount: standard, usedCount: standardMethod, InUse: '⛔' },
  ]

  tableData.push(['Routing Method', 'CX1', 'CX2', 'CX3', 'total built', 'total used', 'In use'])
  for (const row of data) {
    let inUse = '⛔'
    if (row.usedCount > 0) {
      inUse = '✅'
    }
    tableData.push([row.type, row.CX1, row.CX2, row.CX3, row.builtCount.toString(), row.usedCount.toString(), inUse])
  }
  return tableData
}

async function getSkills(pageNumber) {
  let skills = await rapi.getRoutingSkills({ pageNumber: pageNumber, pageSize: 200 })
  globalApiRequests++
  if (pageNumber != skills.pageCount) {
    const nextSkills = await getSkills(pageNumber + 1)
    return skills.entities.concat(nextSkills)
  }
  return skills.entities
}

function getSkillsBuilt(interactions) {
  let totalArray = []
  let voice, message, email, callback

  if (interactions.voice) {
    voice = findIfSkill(interactions.voice)
  }
  if (interactions.message) {
    message = findIfSkill(interactions.message)
  }
  if (interactions.email) {
    email = findIfSkill(interactions.email)
  }
  if (interactions.callback) {
    callback = findIfSkill(interactions.callback)
  }

  // safety against undefined responses
  voice ? null : voice = []
  message ? null : message = []
  email ? null : email = []
  callback ? null : callback = []

  totalArray.push(...voice, ...message, ...email, ...callback)
  console.log(totalArray)
  return totalArray
}

function findIfSkill(interactions) {
  let ids = []
  for (const conv of interactions) {
    for (const participant of conv.participants) {
      for (const session of participant.sessions) {
        if (session.activeSkillIds) {
          ids.push(...session.activeSkillIds)
        }
      }
    }
  }
  return ids
}

async function getRoutingUsageAgg() {
  let response = await capi.postAnalyticsConversationsAggregatesQuery({
    interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
    metrics: [
      "nConnected"
    ],
    groupBy: [
      "queueId",
      "usedRouting"
    ]
  })
  return response
}

async function getRoutingUsage(routingMethod) {
  let method = await capi.postAnalyticsConversationsDetailsQuery({
    interval: `${document.getElementById('datepicker').value.split('/')[0]}T00:00:00${document.getElementById('timeZone').value}/${document.getElementById('datepicker').value.split('/')[1]}T23:59:59${document.getElementById('timeZone').value}`,
    order: "desc",
    orderBy: "conversationStart",
    paging: {
      pageSize: 1,
      pageNumber: 1
    },
    segmentFilters: [
      {
        type: "and",
        predicates: [
          {
            dimension: "requestedRouting",
            value: routingMethod,
          }
        ]
      }
    ],
    conversationFilters: [
      {
        type: "and",
        predicates: [
          {
            metric: "nConnected",
            operator: "exists"
          }
        ]
      }
    ],
  })
  console.log(method.totalHits)
  return method.totalHits
}
