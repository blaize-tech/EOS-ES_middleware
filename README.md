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
Set port on which server will listen.
Set url of elasticsearch cluster.
For example:

    {
        "port": 9000,
        "elastic_url": "http://127.0.0.1:9201"
    }