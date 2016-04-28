package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var (
	addr = flag.String("addr", "http://192.168.30.10:8500", "Consul KV endpoint")
	path = flag.String("path", "", "Consul base path to dump")
	dump = flag.String("dump", "dump.json", "Dump to which file")
)

type ConsulKv struct {
	Key   string `json:Key`
	Value string `json:Value`
}

var httpClient = &http.Client{}

func main() {
	flag.Parse()
	url := fmt.Sprintf("%s/v1/kv/%s?recurse=true", *addr, *path)
	resp, err := httpClient.Get(url)
	if err != nil {
		fmt.Errorf("Error occured while get kv: %+v\n", err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Errorf("Error occured while read body: %+v\n", err)
		return
	}
	kvs := []ConsulKv{}
	err = json.Unmarshal(body, &kvs)

	for i, kv := range kvs {
		if decoded, err := base64.StdEncoding.DecodeString(kv.Value); err == nil {
			kvs[i].Value = string(decoded)
		}
	}
	if data, err := json.MarshalIndent(kvs, "", "    "); err == nil {
		if file, err := os.Create(*dump); err == nil {
			file.Write(data)
		}
	}
}
