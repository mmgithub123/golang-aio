package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

var dirEngineChan sync.Map

func main() {

	dHandler := dagHandler{}
	http.Handle("/v1/dagScan", dHandler)

	cHandler := callbackHandler{}
	http.Handle("/v1/callback", cHandler)

	http.ListenAndServe(":80", nil)

}

type callbackHandler struct{}

func (h callbackHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	//send doStatus message to chan
	query := req.URL.Query()
	dirID := query.Get("dirID")
	engine := query.Get("engine")
	doStatus := query.Get("doStatus")

	doStatusChan, _ := dirEngineChan.Load(dirID + engine)
	doStatusChan.(chan string) <- doStatus

	data := []byte("response message")
	res.WriteHeader(200)
	res.Write(data)

}

type dagHandler struct{}

func (h dagHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	dag := `{
		"345": [
			{
				"fuwenben": "192.168.2.156"
			},
			{
				"mutilEngine": "192.168.2.156",
				"webshell": "192.168.2.156"
			},
			{
				"pe": "192.168.2.156"
			}
		]
	}`

	go dagScan(dag)

	failStr := "message"
	data := []byte(failStr)
	res.WriteHeader(200)
	res.Write(data)

}

func dagScan(dag string) {

	var m map[string][]map[string]string
	json.Unmarshal([]byte(dag), &m)

	dagChan := make(chan string)
	for dirID, engineIpList := range m {
		for _, engineIpMap := range engineIpList {
			var chanList []chan string
			for engine, ip := range engineIpMap {
				dirEngineChan.Store(dirID+engine, dagChan)
				resultChan, _ := dirEngineChan.Load(dirID + engine)
				chanList = append(chanList, resultChan.(chan string))

				httpRequest, err := NewHttpRequest(nil)
				if err != nil {
					panic(err.Error())
				}
				//http://192.168.2.156:8080/v1/engineControl?addr=192.168.2.154&dirID=345&engine=f
				requestUrl := "http://" + ip + ":8080/v1/engineControl?addr=192.168.2.154&dirID=" + dirID + "&engine=" + engine
				fmt.Println(requestUrl)
				httpErr := httpRequest.putMessageUsedGetMethod(requestUrl)
				if httpErr != nil {
					fmt.Println(string(httpRequest.DumpResponse(true)))
				}

				fmt.Println("put dirID:" + dirID + " and engine:" + engine + "to " + ip + " successed")
				fmt.Println(string(httpRequest.LastResponseBody))
			}
			for _, readChan := range chanList {
				fmt.Println("i will block here , until have callback")
				dirEngineResult := <-readChan
				if dirEngineResult != "ok" {
					//del chan
					//del the element key dirID+engine of map
					fmt.Println("return value is not ok, can't go ahead ")
					fmt.Println("del the element key dirID+engine of map")
					return
				}

			}
			fmt.Println()
			fmt.Println()
		}

	}

	//all done
	//del chan
	//del the key dirID+engine element of map
	//del the element of table vs_dirs
	fmt.Println("del the element of table vs_dirs")

}
