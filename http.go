package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Rsps struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
}

func RspsCreate(code int, msg string) []byte {
	r := Rsps{code, msg}
	d, err := json.Marshal(r)
	if err != nil {
		log.Println(err)
		return nil
	}
	log.Println(string(d))
	return d
}

////////////////////////////////////////////////////////////
// get function
////////////////////////////////////////////////////////////
func GetVersion(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	s := fmt.Sprintf("%s %s", AppName, AppVersion)
	d := RspsCreate(200, s)
	w.Header().Set("Content-Type", "application/json")
	return d, nil
}

////////////////////////////////////////////////////////////
// post function
////////////////////////////////////////////////////////////
func PostStreamStateFind(w http.ResponseWriter, r *http.Request, d []byte) ([]byte, error) {
	return nil, nil
}

func PostStreamStateReport(w http.ResponseWriter, r *http.Request, d []byte) ([]byte, error) {
	return nil, nil
}

////////////////////////////////////////////////////////////
// http handler
////////////////////////////////////////////////////////////
func GetHandler(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	var rsps []byte
	var err error
	url := r.URL.String()

	if strings.Contains(url, "/version") {
		rsps, err = GetVersion(w, r)
	} else {
		err = fmt.Errorf("undefined GET request")
		return nil, err
	}
	return rsps, err
}

func PostHandler(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer r.Body.Close()

	var rsps []byte
	url := r.URL.String()

	if strings.Contains(url, "/streamstate/find") {
		rsps, err = PostStreamStateFind(w, r, d)
	} else if strings.Contains(url, "/streamstate/report") {
		rsps, err = PostStreamStateReport(w, r, d)
	} else {
		err = fmt.Errorf("undefined POST request")
		return nil, err
	}
	return rsps, err
}

func HttpServer(w http.ResponseWriter, r *http.Request) {
	log.Println("====== new task start ======")
	log.Println(r.Proto, r.Method, r.URL, r.RemoteAddr, r.Host, r.Header["Upgrade"])

	w.Header().Set("Server", AppName)
	w.Header().Set("Access-Control-Allow-Origin", conf.AllowOrigin)

	var rsps []byte
	var err error
	if r.Method == "GET" {
		rsps, err = GetHandler(w, r)
	} else if r.Method == "POST" {
		rsps, err = PostHandler(w, r)
	} else {
		err = fmt.Errorf("undefined %s request", r.Method)
	}

	if err != nil {
		log.Println(err)
		goto ERR
	}

	w.Header().Set("Content-length", strconv.Itoa(len(rsps)))
	w.Header().Set("Connection", "Keep-Alive")
	w.Write(rsps)
	return
ERR:
	rsps = RspsCreate(500, err.Error())
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-length", strconv.Itoa(len(rsps)))
	w.Header().Set("Connection", "close")
	w.Write(rsps)
}

////////////////////////////////////////////////////////////
// http handler
////////////////////////////////////////////////////////////
func ReportIDs() {
	for {
		log.Println("ReportIDs to StreamNodeFind")
		time.Sleep(time.Second * 10)
	}
}
