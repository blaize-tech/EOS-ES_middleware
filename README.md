# ESHistoryAPI
History API for Elasticsearch cluster

## Installation
#### Get source code
```sh
$ cd $GOPATH/src
$ git clone https://github.com/InCrypto-io/EOS-ES_middleware.git
$ cd EOS-ES_middleware/
$ git checkout dev
```
#### Get dependencies using dep
```sh
$ dep ensure
```
#### Create config.json
In project directory create file config.json.  
"port" property is for the port on which server will listen.  
"elastic_url" property is for the url of elasticsearch cluster.  
For example:

    {
        "port": 9000,
        "elastic_url": "http://127.0.0.1:9201"
    }
#### Run
Run with  
```sh
$ go run *.go
```

## Usage
This API supports following GET requests:  

#### /v1/history/get_actions
Requires json body with the following properties:  
account_name - name of the eos account. This field is required.  
pos - position in a list of account actions sorted by global_sequence (e.g. in chronological order). This field is not required.  
offset - number of actions to return. This field is not required.  
  
Returns json with the following properties:  
actions - array of actions of given account  
#### /v1/history/get_transaction
Requires json body with the following properties:  
id - id of transaction.  
  
Returns json with the following properties:  
id - id of transaction.  
trx - transaction.  
block_time - timestamp of block which contains requested transaction.  
block_num - number of block which contains requested transaction.  
traces - traces of transaction.  
#### /v1/history/get_key_accounts
Requires json body with the following properties:  
public_key - public key of account
  
Returns json with the following properties:  
account_names - array of accounts that have requested key  
#### /v1/history/get_controlled_accounts
Requires json body with the following properties:  
controlling_account - name of the eos account  
  
Returns json with the following properties:  
controlled_accounts - array of accounts controlled by requested account  