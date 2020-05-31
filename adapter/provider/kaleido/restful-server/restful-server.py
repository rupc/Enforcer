#!/usr/bin/python3

from flask import Flask
from flask import request
from flask_restful import Resource, Api
from provider.kaleido import kaleido

app = Flask(__name__)
api = Api(app)

class HelloWorld(Resource):
    def get(self):
        return {'hello': 'world'}

class GetBlock(Resource):
    def get(self, number):
        block = kaleido.get_block(number)
        r = {
                "number": block.number,
                "hash": block.hash.hex(),
                "transactions": block.transactions,
                }

        return r

class SendTransaction(Resource):
    def post(self):
        req = request.get_json()
        print(req, type(req))
        txid = kaleido.send_transaction(req)
        print("TX 날린 결과", txid.hex(), type(txid))
        return txid.hex()
        #  return "shitty"

api.add_resource(HelloWorld, '/')
api.add_resource(SendTransaction, '/api/v1/SendTransaction')
api.add_resource(GetBlock, '/api/v1/GetBlock/<int:number>')

if __name__ == '__main__':
    app.run(debug=True)
