#!/bin/bash

PATH_TO_PLUGIN=$(pwd) 
ascc_src_path="$PATH_TO_PLUGIN/ascc.so"
# ascc_dst_path="/home/jyr/go/src/github.com/hyperledger/caliper/packages/caliper-expr/network/fabric-v1.4.1/raft/ascc/"

# auditchain_remote_name="bears.postech.ac.kr"
# auditchain_remote_ip="141.223.121.66"
# ascc_install_path="/home/auditchain/fabric-samples/balance-transfer/artifacts/ascc/ascc.so"



# echo "Copy ascc.so in $ascc_src_path to $ascc_dst_path"
# cp $ascc_src_path $ascc_dst_path


# deploy_remote used in two-chain approach
function deploy_remote() {
    auditchain_remote_name="bears.postech.ac.kr"
    ascc_dst_remote_auditchain_path="auditchain@$auditchain_remote_ip:$ascc_install_path"
    echo "Copy ascc.so in $ascc_src_path to $ascc_dst_remote_auditchain_path"
    scp $ascc_src_path $ascc_dst_remote_auditchain_path
}

# deploy_local used in one-chain approach
function deploy_local() {
    ascc_install_path="/home/jyr/go/src/github.com/hyperledger/caliper/packages/caliper-expr/network/fabric-v1.4.1/raft/ascc/ascc.so"
    #function_body
    cp $ascc_src_path $ascc_install_path
}

deploy_local
