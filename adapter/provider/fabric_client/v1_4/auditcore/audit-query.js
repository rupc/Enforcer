#!/usr/bin/env node
'use strict';

// Query auditors
//
var http = require('http');
var request = require('request')

var req = {
    "cmd":"query"
}

request.get({
     url: "http://localhost:5959/query",
     headers: {
        "Content-Type": "application/json"
     },
     body: req,
     json:true
}, function(error, response, body){
   console.log(error);
   console.log(body);
});
