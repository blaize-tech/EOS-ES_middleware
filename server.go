package main

import (
	"fmt"
	"io/ioutil"
    "net/http"
	"encoding/json"
	"github.com/olivere/elastic"
)


const ApiPath string = "/v1/history/"


type Config struct {
	Port       uint32 `json:"port"`
	ElasticUrl string `json:"elastic_url"`
}


type Server struct {
    ElasticClient *elastic.Client
    //router *someRouter //not sure
}


func (s * Server) listen(port uint32) {
	err := http.ListenAndServe(":" + fmt.Sprint(port), nil)
    if err != nil {
        panic(err)
    }
}


func (s *Server) initElasticClient(url string) {
	client, err := elastic.NewClient(
		elastic.SetURL(url),
		elastic.SetSniff(false))
	if err != nil {
		panic(err)
	} else {
		s.ElasticClient = client
	}
}

func (s *Server) setRoutes() {
	http.HandleFunc(ApiPath + "get_actions", s.onlyGet(s.handleGetActions()))
	http.HandleFunc(ApiPath + "get_transaction", s.onlyGet(s.handleGetTransaction()))
	http.HandleFunc(ApiPath + "get_key_accounts", s.onlyGet(s.handleGetKeyAccounts()))
	http.HandleFunc(ApiPath + "get_controlled_accounts", s.onlyGet(s.handleGetControlledAccounts()))
}


//onlyGet take function (http handler) as an argument
//and returns function that takes http.ResponseWriter and *http.Request
//this function will call given handler only if http method of the request is GET
//otherwise it will respond with 405 error code
func (s *Server) onlyGet(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if (r.Method != http.MethodGet) {
			http.Error(w, "Invalid request method.", 405)
			return
		}
		h(w, r)
	}
}

//handleGetActions returns http handler that takes
//http.ResponseWriter and *http.Request as arguments
//it tries to parse parameters from request body
//and passes them to getActions()
//The result of getActions() is encoded and sent as a response
func (s *Server) handleGetActions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var params GetActionsParams
		err = json.Unmarshal(bytes, &params)
		if err != nil {
			http.Error(w, "Invalid arguments.", 400)
			return
		}
		if params.Pos == nil {
			params.Pos = new(int64)
			*params.Pos = 0
		}
		if params.Offset == nil {
			params.Offset = new(int64)
			*params.Offset = 0
		}

		result, err := getActions(s.ElasticClient, params)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		b, err := json.Marshal(result)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, string(b))
	}
}

//handleGetTransaction returns http handler that takes
//http.ResponseWriter and *http.Request as arguments
//it tries to parse parameters from request body
//and passes them to getTransaction()
//The result of getTransaction() is encoded and sent as a response
func (s *Server) handleGetTransaction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var params GetTransactionParams
		err = json.Unmarshal(bytes, &params)
		if err != nil {
			http.Error(w, "Invalid arguments.", 400)
			return
		}

		result, err := getTransaction(s.ElasticClient, params)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		b, err := json.Marshal(result)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, string(b))
	}
}

//handleGetKeyAccounts returns http handler that takes
//http.ResponseWriter and *http.Request as arguments
//it tries to parse parameters from request body
//and passes them to getKeyAccounts()
//The result of getKeyAccounts() is encoded and sent as a response
func (s *Server) handleGetKeyAccounts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var params GetKeyAccountsParams
		err = json.Unmarshal(bytes, &params)
		if err != nil {
			http.Error(w, "Invalid arguments.", 400)
			return
		}
		
		result, err := getKeyAccounts(s.ElasticClient, params)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		b, err := json.Marshal(result)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, string(b))
	}
}

//handleGetControlledAccounts returns http handler that takes
//http.ResponseWriter and *http.Request as arguments
//it tries to parse parameters from request body
//and passes them to getControlledAccounts()
//The result of getControlledAccounts() is encoded and sent as a response
func (s *Server) handleGetControlledAccounts() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var params GetControlledAccountsParams
		err = json.Unmarshal(bytes, &params)
		if err != nil {
			http.Error(w, "Invalid arguments.", 400)
			return
		}

		result, err := getControlledAccounts(s.ElasticClient, params)
		b, err := json.Marshal(result)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, string(b))
	}
}