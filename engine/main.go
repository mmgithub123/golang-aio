package main

import (
	"fmt"
	"net/http"
)

func main() {

	eHandler := engineHandler{}
	http.Handle("/v1/engineControl", eHandler)

	http.ListenAndServe(":8080", nil)

}

type engineHandler struct{}

func (h engineHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	query := req.URL.Query()

	dirID := query.Get("dirID")
	engineName := query.Get("engine")
	addr := query.Get("addr")

	fmt.Println(dirID, engineName, addr)
	go doScan(dirID, engineName, addr)

	data := []byte("response message")
	res.WriteHeader(200)
	res.Write(data)
}

func doScan(dirID string, engine string, addr string) {
	//download file
	//scan file
	//write result
	//call the callback

	//fmt.Println("receive message ,but do not send the callback request in some reason, in this monment,the router will block, until timeout")
	//return

	httpRequest, err := NewHttpRequest(nil)
	if err != nil {
		panic(err.Error())
	}
	httpErr := httpRequest.putMessageUsedGetMethod("http://" + addr + ":80/v1/callback?doStatus=ok&dirID=" + dirID + "&engine=" + engine + "")
	if httpErr != nil {
		fmt.Println(string(httpRequest.DumpResponse(true)))
	}
	fmt.Println("call " + addr + " callback " + dirID + " and" + engine + " message successed")
	fmt.Println(string(httpRequest.LastResponseBody))
}
