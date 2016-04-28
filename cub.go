package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	addr    = flag.String("addr", "http://127.0.0.1:8500", "Consul KV endpoint")
	path    = flag.String("path", "", "Consul base path to dump")
	restore = flag.Bool("restore", false, "Restore to Consul KV with json file")
	backup  = flag.Bool("backup", false, "Backup Consul KV to local json file")
	ignore  = flag.String("ignore", "", "Consul key to ignore")
	dump    = flag.String("dump", "dump.json", "Dump to which file")
)

type ConsulKv struct {
	Key   string `json:Key`
	Value string `json:Value`
}

var httpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConnsPerHost:   5,
		ResponseHeaderTimeout: 2 * time.Second,
	},
	Timeout: 5 * time.Second,
}
var ignoreKeyRe *regexp.Regexp

func main() {
	flag.Parse()
	if *ignore != "" {
		if re, err := regexp.Compile(*ignore); err == nil {
			ignoreKeyRe = re
		}
	}

	switch {
	case *backup:
		backupToLocal()
	case *restore:
		restoreToConsul()
	}
}

func backupToLocal() {
	url := fmt.Sprintf("%s/v1/kv/%s?recurse=true", *addr, *path)
	resp, err := httpClient.Get(url)
	if err != nil {
		fmt.Println(fmt.Errorf("Error occured while get kv: %+v\n", err))
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(fmt.Errorf("Error occured while read body: %+v\n", err))
		return
	}
	kvs := []ConsulKv{}
	err = json.Unmarshal(body, &kvs)

	output := make([]ConsulKv, len(kvs))
	i := 0
	for _, kv := range kvs {
		if ignoreKeyRe != nil && ignoreKeyRe.MatchString(kv.Key) {
			continue
		}
		if decoded, err := base64.StdEncoding.DecodeString(kv.Value); err == nil {
			output[i] = ConsulKv{kv.Key, string(decoded)}
			i++
		}
	}
	if data, err := json.MarshalIndent(output, "", "    "); err == nil {
		if file, err := os.Create(*dump); err == nil {
			file.Write(data)
			fmt.Printf("Backup to local '%s' success!\n", *dump)
		}
	}
}

func restoreToConsul() {
	raw, err := ioutil.ReadFile(*dump)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	var kvs []ConsulKv
	json.Unmarshal(raw, &kvs)

	base := fmt.Sprintf("%s/v1/kv", *addr)
	if *path != "" {
		base = fmt.Sprintf("%s/%s", base, *path)
	}

	for _, kv := range kvs {
		if ignoreKeyRe != nil && ignoreKeyRe.MatchString(kv.Key) {
			continue
		}
		url := fmt.Sprintf("%s/%s", base, kv.Key)
		if req, err := http.NewRequest("PUT", url, strings.NewReader(kv.Value)); err == nil {
			resp, err := httpClient.Do(req)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			b, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			if code := resp.StatusCode; code != 200 {
				fmt.Println(fmt.Errorf("PUT %s: %d, reson: %s\n", url, code, b))
			} else {
				fmt.Printf("PUT %s: success!\n", url)
			}
		}
	}
}
