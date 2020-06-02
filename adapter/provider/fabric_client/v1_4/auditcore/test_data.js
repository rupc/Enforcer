'use strict';

// Test codes...
// Created when a new block arrives
var newBlockMessage = {
    // 
    "auditor_id": "auditor1",
    "content": {
        "number":10,
        // Chain to audit
        "chain_id": "basschain",
        "hash": "0x121abdd",
        "prev_hash": "0xaaa121abdd",

        "tx_set": [
            {
                "member": "peer0",
                "chain_id": "my-channel",
                "number": 11,
                "hash": "0xdeadbeef",
            },
            {
                "member": "peer1",
                "chain_id": "my-channel",
                "number": 11,
                "hash": "0xdeadbeef",
            },
            {
                "member": "peer2",
                "chain_id": "my-channel",
                "number": 11,
                "hash": "0xdeadbeef",
            },
            {
                "member": "peer3",
                "chain_id": "my-channel",
                "number": 11,
                "hash": "0xdeadbeef",
            },
        ], 
    }
}


exports.newBlockMessage = newBlockMessage
