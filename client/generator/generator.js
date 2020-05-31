#!/usr/bin/env node
var jStat = require('jStat');

var log4js = require('log4js');
var logger = new log4js.getLogger('generator');
logger.level = 'debug';

var util = require('util');
var wakeUpCnt = 0;
var lastBlockNum = 0
var txRequestPools = [];
var crypto = require('crypto');




// start registers event handler for event emitter (em)
// It generates a set of special transaction on the received block
// and queues the set into submitbuffer
function start(config, clients) {
    var redis_db = config.db;
    var policy = config.policy;
    var em = config.em;
    var io = config.io;
    var invoker = config.invoker;

    setTimeout(function(){
        logger.info("Exit experiment!")
        process.exit();
    }, 360000);

    logger.info("startGenerator starts");

    function start_evaluation() {
        policy.evaluation.scenarios.delay_change.forEach((s) => {
            logger.info("Evaluation enabled", s);
            setTimeout(()=> {
                if (s.item == "fidelity_change") {
                    var target_index = clients.findIndex(function(c) {
                        return c.id == s.id
                    });
                    clients[target_index].fidelity = s.fidelity
                    logger.debug("Changing fidelity!");
                } else if (s.item == "delay_change") {
                    var target_index = clients.findIndex(function(c) {
                        return c.id == s.id
                    });
                    clients[target_index].delay = s.delay
                    logger.debug("Changing delay!");
                }
            }, s.time);
        });
    }
    start_evaluation();
    logger.info("experimental configuration starts");
    // counts array stores pairs of (client, votes)
    var counts =  new Map();
    for (var c of clients) {
        counts.set(c.id, 0);
    }

    // Generator subscribe a new block event
    // On each new block, it generates a new special transaction based on a given policy
    var num_total_audit_tx = 0;

    em.on('new-block', function (block) {
        // logger.info("on new-block event[%d]", block.header.number);

        var num_audit_tx = 0;
        var txes = [];
        // function filterChaincodeID(argument) {
            
        // }
        block.data.data.forEach((tx) => {
            // Filter special transactions from the block
            var tx_type = "app";
            if (tx.payload.data.actions[0].payload.action.proposal_response_payload.extension.chaincode_id.name == "ascc") {
                num_audit_tx++;
                tx_type = "ascc";
            } 

            // collect committed transaction in this block
            txes.push({
                "txid": tx.payload.header.channel_header.tx_id,
                "type": tx_type,
            });

        });


        var newTxes = [];

        // logger.info(util.inspect(block, false, null, true));
// console.log(JSON.stringify(block, null, '\t'));
        // console.log("start of block");
        // console.dir(JSON.stringify(block), { depth: null })
        // console.log("end of block");
        clients.forEach(function(c){
            if (!does_submit_on_this_block(c)) {
                return;
            }

            // build special transaction
            var tx = {
                "member_id": c.id,
                "number": block.header.number,
                "hash": block.header.data_hash,
                "chain_id": "basschain",
            }

            // console.log("value");
            newTxes.push(tx);

            var votes = counts.get(c.id) + 1;
            counts.set(c.id, votes);

        });


        // emits a set of new special transactions 
        // Note that I had seperated the generation phase and submission phase, 
        // This has benefits in flexibility and controllability: simulating a censorship attack
        em.emit('new-special-tx', newTxes);

        // build a sentData for the dashboard client to display the contents
        var sentData = {
            "numOfMembers": clients.length,
            "votes": JSON.stringify(Array.from(counts.entries())),
            "block_num": block.header.number,
            "data_hash": block.header.data_hash,
            "num_tx": block.data.data.length,
            "num_total_tx": block.data.data.length,
            "num_audit_tx": num_audit_tx,
            "num_total_audit_tx": num_total_audit_tx,
            "txes": JSON.stringify(txes),
        }

        redis_db.get("num_total_tx", function (error, num_total_tx) {
            redis_db.get("num_special_tx", function (error, num_special_tx) {
                var n = Number(num_special_tx) + num_audit_tx;
                redis_db.set("num_special_tx", n);
                var sentData = {
                    numOfTotalTx: Number(num_total_tx),
                    numOfTotalSpecialTx: Number(num_special_tx),
                }
                io.emit("tx", JSON.stringify(sentData));
                // console.log('GET result ->' + num_total_tx, num_special_tx);
            })
        })

        // Emits updated pairs of (client, votes), expected to be plotted in d3.js in client side
        io.emit("votes", JSON.stringify(sentData));
    });

    // On new-special-tx event, it accepts a set of special transactions (txes),
    // then submit them through ASCC interface
    em.on('new-special-tx', async function(txes) {
        // For each built special transaction,
        // it submits each special transaction to the target chain
        // chain_id, number, hash, member
        txes.forEach(async function(tx) {
            var functionName = "storeAuditTx";
            var targetChainID = tx.chain_id;
            var number = tx.number;
            var hash = tx.hash;
            var member = tx.member_id;
            var chaincodeName = "ascc";
            var nonce = crypto.randomBytes(16).toString('hex');

            var args = [functionName, targetChainID, number.toString(), hash, member, nonce];
            // logger.debug(args);
            // Send Special Transaction from AuditCore to AuditChain
            // await invoker.invoke(args);
           
            // adding a delay to the invokeASCC means a selective censorship attack
            // note that if dealy in policy is zero, then randomDelay is also zero
            var target_member = clients.find(function(c) {
                return c.id == member;
            });

            // Injecting a random delay
            var randomDelay = Math.ceil(jStat.normal.sample(target_member.delay, target_member.delay / 3));
            setTimeout(function() {
                invoker.invokeASCC(args);
            }, randomDelay);
        });

    });

    // 
    // Emit a new block event !!! just for testing !!!
    // var blknum = 0;
    // setInterval(()=>{
    //     var block = {
    //         header: {
    //             number: blknum++,
    //             data_hash: crypto.randomBytes(16).toString('hex'),
    //         },
    //         data : {
    //             data : ['a', 'b'],
    //         }
    //     }
    //     em.emit('new-block', block);
    // }, 2000);


    // does_submit_on_this_block decides **probablisticaly** whether the client submits 
    // a new transaction or not based on the fidelity
    function does_submit_on_this_block(c) {
        var r = Math.random();
        if (r < c.fidelity) {
            return true;
        }
        else return false
    }

    // setInterval(()=>{
    //     wakeUpCnt++;

    //     blockstore.get("last_written_block")
    //     .then(function(last_num) {
    //         if (lastBlockNum < last_num) {
    //             lastBlockNum = last_num
    //             logger.debug("New block detection !");
    //             // Create a special transaction.
    //             // XXX 클라이언트 갯수 * 충실도 갯수 만큼 스페셜 트랜잭션 생성
    //             // 물론, 
    //             for (var i = 0; i < policy.clients.length; ++i) {
    //                 var tx = {
    //                     "member_id" : "client" + i,
    //                     "block_num" : lastBlockNum,
    //                 }
    //                 txRequestPools.push(tx);
    //             }

    //         }
    //         logger.debug(last_num) ;
    //     })
    //     .catch(function (err) { console.error(err) })

    // }, policy.period);
    
}

e
