
'use strict';
var log4js = require('log4js');
var logger = new log4js.getLogger('auditcore');
var request = require('request');
var hfc = require('fabric-client');
var cr
logger.level = 'debug';
var http = require('http');

function doPreliminaryJobBeforeSend(newBlockMessage) {
    var fork_clients = hfc.getConfigSetting("experiments")["fork"]["fork_clients"];
    var fork_block_number = hfc.getConfigSetting("experiments")["fork"]["fork_block_number"];

    if (hfc.getConfigSetting("experiments")["fork"]["conduct"] == true) {
        if (fork_clients.includes(newBlockMessage.client_id)) {
            if (newBlockMessage.content.number == fork_block_number) {
                logger.debug("newBlockMessage.content.Number == fork_block_number, fork occurs")
                // Inject faulty hash
                newBlockMessage.content.hash = "0x000ff1cedeadbeef";
            } else if (newBlockMessage.content.Number < fork_block_number) {
                logger.debug("newBlockMessage.content.Number < fork_block_number, no fork")
            } else {
                logger.debug("newBlockMessage.content.Number > fork_block_number, no fork")
            }
        }
    }

    return newBlockMessage
}

// generateSpecialTransactionRequest is actually same as sendBlockToAuditCore for notiyfing a new block message to Auditcore
//
function generateSpecialTransactionRequest(argument) {
    
}

// sendAnalysisRequest sends a start-analysis request from a dashboard user to AuditCore
function sendAnalysisRequest(params) {
    var auditcoreAddress = hfc.getConfigSetting("auditcore")["address"] + "/audit";

    var bodyParams = {
        "cmd": "start_analysis",
        "auditor_id": "auditor1",
        "analysis_type": params.analysis_type,
        "start_blk_num": params.start_blk_num,
        "end_blk_num": params.end_blk_num,
        "window_size": params.window_size,
    }

    request.post({
        url: auditcoreAddress,
        headers: {
            "Content-Type": "application/json"
        },
        body: bodyParams,
        json:true
    }, function(error, response, body){
        if (!error && response.statusCode == 200) {
            logger.info(body);
        } else {
            logger.error("Cannot connect to AuditCore: Checkout the availability of AuditCore")
            logger.error(error);
        }
    });

}

// sendBlockToAuditCore send a AuditBlcok from AuditChain to AuditCore
function sendBlockToAuditCore(newBlockMessage) {
    // Preliminary jobs before send
    newBlockMessage = doPreliminaryJobBeforeSend(newBlockMessage);

    // var auditcoreAddress = "http://localhost:5959/new_block"
    var auditcoreAddress = hfc.getConfigSetting("auditcore")["address"] + hfc.getConfigSetting("auditcore")["endpoints"]["new_block"];
    
    request.post({
        url: auditcoreAddress,
        headers: {
            "Content-Type": "application/json"
        },
        body: newBlockMessage,
        json:true
    }, function(error, response, body){
        logger.error("Cannot connect to AuditCore: Checkout the availability of AuditCore")
        logger.error(error);
        // logger.error(JSON.stringify(response));
        // logger.error(body);
        // logger.error("AudirCore not available, shutting down client");
        // process.exit();
        // console.log(error);
        // console.log(JSON.stringify(response));
        // console.log(body);
    });
}

exports.generateSpecialTransactionRequest = generateSpecialTransactionRequest
exports.sendBlockToAuditCore = sendBlockToAuditCore
exports.sendAnalysisRequest = sendAnalysisRequest
