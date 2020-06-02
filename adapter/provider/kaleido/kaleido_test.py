#!/usr/bin/python3
import web3
import base64

from web3 import Web3, HTTPProvider
from web3.middleware import geth_poa_middleware
from time import sleep
import json
import binascii
# TESTED WITH python 3.6

USER = "k0a70jrp15"
PASS = "Uy3k9SFgqaXuPS-M0XT118QwQXKULV6tMXALfvhyJBM"
RPC_ENDPOINT = "https://k0f5xi38x9-k0pkr068zj-rpc.kr0-aws.kaleido.io"

# Encode the username and password from the app creds into USER:PASS base64 encoded string
auth = USER + ":" + PASS
encodedAuth = base64.b64encode(auth.encode('ascii')).decode('ascii')

# Build the header object with the Basic auth and the standard headers
headers = {'headers': {'Authorization': 'Basic %s' % encodedAuth,
                        'Content-Type': 'application/json',
                        'User-Agent': 'kaleido-web3py'}}

# Construct a Web3 object by constructing and passing the HTTP Provider
provider = HTTPProvider(endpoint_uri=RPC_ENDPOINT, request_kwargs=headers)
w3 = Web3(provider)


# Add the Geth POA middleware needed for ExtraData Header size discrepancies between consensus algorithms
# See: http://web3py.readthedocs.io/en/stable/middleware.html#geth-style-proof-of-authority
# ONLY for GETH/POA; If you are using quorum, comment out the line below
w3.middleware_stack.inject(geth_poa_middleware, layer=0)


def print_pretty_block(block):
    print("height", block.number)
    print("hash", block.hash.hex())
    print("# of Tx", len(block.transactions))
    print("transactions", block.transactions)
    for tx in block.transactions:
        # print("TXTX")
        t = w3.eth.getTransaction(tx)
        if t["from"] == t["to"]:
            print("Special Trasaction!")
        # print(t)
    # r = json.loads(block.data.decode())
    # print(r)


def build_transaction(block):
    tx = {
            "hash": block.hash.hex(),
            "number": block.number,
            "id": USER,
            "chainID": 1645172027,
            }
    return tx

def send_transaction(tx):
    #  data = json.dumps(tx, indent=2)
    data = json.dumps(tx, indent=2).encode('utf-8')
    print(data)
    txid = w3.eth.sendTransaction({
        'to': w3.toChecksumAddress('0xaefd959964419618ad866d3bc04a2d1164769447'),
        'from': w3.toChecksumAddress('0xaefd959964419618ad866d3bc04a2d1164769447'),
        # 'from': w3.toChecksumAddress('0x3a278e53a2f29a283eba788ba773ef20a6682a24'),
        'value': 0,
        'data': data
        })
    #  print("Committed Tx", binascii.b2a_hex(txid))
    return txid

# Get the latest block in the chain
def get_block(num):
    block = w3.eth.getBlock(num)
    return block

block = get_block(10000)
print_pretty_block(block)

# signed_txn = w3.eth.account.signTransaction(dict(
#     nonce=w3.eth.getTransactionCount('yourAddress'),
#     gasPrice = w3.eth.gasPrice,
#     gas = 1,
#     to='recipientAddress',
#     value=web3.toWei(12345,'ether')
#   ),
#   'yourprivatekey')
#
# w3.eth.sendRawTransaction(signed_txn.rawTransaction)

# Print the block out to the console
# print(block)

# Alice
# k0ujl3cmh1
# u9PzQ9bdxmaQ5SaKerEJncLKPI5alAaD02b04OrTJtE
alice = {'name': 'k0ujl3cmh1', 'passwd': 'u9PzQ9bdxmaQ5SaKerEJncLKPI5alAaD02b04OrTJtE'}
app = {'name': 'k0a70jrp15', 'passwd': 'Uy3k9SFgqaXuPS-M0XT118QwQXKULV6tMXALfvhyJBM'}

