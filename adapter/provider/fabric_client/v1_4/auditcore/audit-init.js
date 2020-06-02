#!/usr/bin/env node

// This script should be exectued only after AuditCore is initialized,
// It registers both a client and its corrensponding auditor. For sample workload, it register a pair of (client#, auditor#)
// Checkout handler parts in audit_handler.go (maybe) in server side

'use strict';


var http = require('http');
var request = require('request');
var config = require('../config.js');

var hfc = require('fabric-client');

var sdk_host = "http://141.223.121.56:" + hfc.getConfigSetting("port");

var tx_send_address = sdk_host + "/auditcore/send-tx";

var state_send_address = sdk_host + "/auditcore/state";

var report_address = sdk_host + "/auditcore/report";

// var metric_fetch_address = hfc.getConfigSetting("metric")["leveldb_path"];

var fetch_method = hfc.getConfigSetting("metric")["fetch_method"];

var metric_fetch_address;

if (fetch_method ==  "redisdb") {
    metric_fetch_address = hfc.getConfigSetting("metric")["redisdb_path"];
} else if (fetch_method == "leveldb") {
    metric_fetch_address = hfc.getConfigSetting("metric")["leveldb_path"];
} else if (fetch_method == "prom") {
    metric_fetch_address = hfc.getConfigSetting("metric")["prom_address"];
} else {
    metric_fetch_address = "redis"
}

var auditcore_address = hfc.getConfigSetting("auditcore")["address"];
var fetch_address = "141.223.121.56:12000";
// var metric_fetch_address = hfc.getConfigSetting("metric")["prom_address"];

var metricNames = require("../metrics/metrics.js").GetMetricNames();



// console.log(tx_send_address);
// console.log(state_send_address);
// console.log(metricNames);

// Number of client and its corrensponding auditor

function sample_init(names, num_client) {
    var reqs = []
    for (var i = 0; i < num_client; ++i) {
        var r = {
            "cmd":"init",
            "creator_id": names[i],
            "auditor_id": "auditor:" + names[i],
            "target_chain_id": "basschain",
            "record_chain_id": "auditchain",
            "tx_send_address": tx_send_address,
            "state_send_address": state_send_address,
            "fetch_method": fetch_method,
            "metric_fetch_address": metric_fetch_address,
            "metric_names": metricNames,
            "report_address": report_address,
            "fetch_address": fetch_address,
            "fidelity": "1.0",
        }
        reqs.push(r);
    }

    return reqs;
}


// Send registration requests, i.e., pairs of (auditor, client) to AuditCore 
function make_request(names, num_client) {
    // if (num_client === null) {
    //     num_client = 4
    // }

    // build requests
    var init_reqs = sample_init(names, num_client);

    init_reqs.forEach(function(initReq) {
        request.post({
            url: auditcore_address + "/audit",
            headers: {
                "Content-Type": "application/json"
            },
            body: initReq,
            json:true
        }, function(error, response, body){
            if (error != null) {
                console.log(error);
            }
            console.log(body);
        });
    });

    return init_reqs;
}

exports.make_request = make_request;
