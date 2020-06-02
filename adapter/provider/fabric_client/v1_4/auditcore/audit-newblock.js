#!/usr/bin/env node
'use strict';

// New block arrival...
var http = require('http');
var request = require('request')

var test_data = require('./test_data.js');


request.post({
    url: "http://localhost:5959/new_block",
    headers: {
        "Content-Type": "application/json"
    },
    body: test_data.newBlockMessage,
    json:true
}, function(error, response, body){
    console.log(error);
    console.log("Received special transaction! (below)");
    console.log(JSON.stringify(response));
    console.log(body);
});
