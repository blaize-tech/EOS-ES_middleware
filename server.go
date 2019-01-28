package main

import (
	"fmt"
	"io/ioutil"
    "net/http"
	"encoding/json"
	"github.com/olivere/elastic"
)


const ApiPath string = "/v1/history/"
const AccountsIndexPrefix          string = "accounts"
const TransactionsIndexPrefix      string = "transactions"
const TransactionTracesIndexPrefix string = "transaction_traces"
const ActionTracesIndexPrefix      string = "action_traces"


type Config struct {
	Port       uint32 `json:"port"`
	ElasticUrl string `json:"elastic_url"`
}


type Server struct {
	ElasticUrl string
    ElasticClient *elastic.Client
    Indices map[string][]string
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
		s.ElasticUrl = url
		s.getIndices()
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
			w.WriteHeader(http.StatusMethodNotAllowed)
			response := ErrorResult { Code: http.StatusMethodNotAllowed, Message: "Invalid arguments." }
			json.NewEncoder(w).Encode(response)
			return
		}
		h(w, r)
	}
}

func (s *Server) getIndices() {
	prefixes := []string {
		AccountsIndexPrefix,
		TransactionsIndexPrefix,
		TransactionTracesIndexPrefix,
		ActionTracesIndexPrefix }
	s.Indices = getIndices(s.ElasticUrl, prefixes)
}

//handleGetActions returns http handler that takes
//http.ResponseWriter and *http.Request as arguments
//it tries to parse parameters from request body
//and passes them to getActions()
//The result of getActions() is encoded and sent as a response
func (s *Server) handleGetActions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := ErrorResult { Code: http.StatusInternalServerError, Message: err.Error() }
			json.NewEncoder(w).Encode(response)
			return
		}

		var params GetActionsParams
		err = json.Unmarshal(bytes, &params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response := ErrorResult { Code: http.StatusBadRequest, Message: "Invalid arguments." }
			json.NewEncoder(w).Encode(response)
			return
		}
		if params.Pos == nil {
			params.Pos = new(int64)
			*params.Pos = -1
		}
		if params.Offset == nil {
			params.Offset = new(int64)
			*params.Offset = -20
		}

		result, err := getActions(s.ElasticClient, params, s.Indices)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := ErrorResult { Code: http.StatusInternalServerError, Message: err.Error() }
			json.NewEncoder(w).Encode(response)
			return
		}
		b, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := ErrorResult { Code: http.StatusInternalServerError, Message: err.Error() }
			json.NewEncoder(w).Encode(response)
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
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := ErrorResult { Code: http.StatusInternalServerError, Message: err.Error() }
			json.NewEncoder(w).Encode(response)
			return
		}

		var params GetTransactionParams
		err = json.Unmarshal(bytes, &params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response := ErrorResult { Code: http.StatusBadRequest, Message: "Invalid arguments." }
			json.NewEncoder(w).Encode(response)
			return
		}

		result, err := getTransaction(s.ElasticClient, params, s.Indices)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := ErrorResult { Code: http.StatusInternalServerError, Message: err.Error() }
			json.NewEncoder(w).Encode(response)
			return
		}
		b, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := ErrorResult { Code: http.StatusInternalServerError, Message: err.Error() }
			json.NewEncoder(w).Encode(response)
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
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := ErrorResult { Code: http.StatusInternalServerError, Message: err.Error() }
			json.NewEncoder(w).Encode(response)
			return
		}

		var params GetKeyAccountsParams
		err = json.Unmarshal(bytes, &params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response := ErrorResult { Code: http.StatusBadRequest, Message: "Invalid arguments." }
			json.NewEncoder(w).Encode(response)
			return
		}
		
		result, err := getKeyAccounts(s.ElasticClient, params, s.Indices)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := ErrorResult { Code: http.StatusInternalServerError, Message: err.Error() }
			json.NewEncoder(w).Encode(response)
			return
		}
		b, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := ErrorResult { Code: http.StatusInternalServerError, Message: err.Error() }
			json.NewEncoder(w).Encode(response)
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
		defer r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := ErrorResult { Code: http.StatusInternalServerError, Message: err.Error() }
			json.NewEncoder(w).Encode(response)
			return
		}

		var params GetControlledAccountsParams
		err = json.Unmarshal(bytes, &params)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			response := ErrorResult { Code: http.StatusBadRequest, Message: "Invalid arguments." }
			json.NewEncoder(w).Encode(response)
			return
		}

		result, err := getControlledAccounts(s.ElasticClient, params, s.Indices)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := ErrorResult { Code: http.StatusInternalServerError, Message: err.Error() }
			json.NewEncoder(w).Encode(response)
			return
		}
		b, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response := ErrorResult { Code: http.StatusInternalServerError, Message: err.Error() }
			json.NewEncoder(w).Encode(response)
			return
		}
		fmt.Fprintf(w, string(b))
	}
}