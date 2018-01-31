package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type GTSS []struct {
	C string            `json:"c"`
	L map[string]string `json:"l"`
	V [][]float64       `json:"v"`
}

type Stack []GTSS

type Config struct {
	Endpoint string `json:"endpoint"`
}

func main() {
	configuration := Config{}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	// Config
	configFile, err := os.Open(dir + "/config.json")
	if err != nil {
		log.Fatal(err)
	}

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {

		if !strings.HasSuffix(f.Name(), ".mc2") {
			continue
		}

		data, err := ioutil.ReadFile(dir + "/" + f.Name())
		if err != nil {
			log.Fatal(err)
		}

		resp, err := http.Post(configuration.Endpoint+"/api/v0/exec", "application/x-www-form-urlencoded;", strings.NewReader(string(data)))
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode != 200 {

			dump, err := httputil.DumpResponse(resp, true)
			if err != nil {
				log.Fatal(err)
			}

			log.Fatalf("%q", dump)
		}

		defer resp.Body.Close()

		var stack Stack

		err = json.NewDecoder(resp.Body).Decode(&stack)
		if err != nil {
			panic(err)
		}

		var lines [][]string
		lines = append(lines, []string{"timestamp(in sec)", "value"})

		for _, gtss := range stack {
			for _, gts := range gtss {

				for _, tick := range gts.V {
					ts := strconv.FormatFloat(tick[0]/1000.0/1000.0, 'E', -1, 64)
					value := strconv.FormatFloat(tick[1], 'E', -1, 64)
					lines = append(lines, []string{ts, value})
				}

				out, err := os.Create(dir + "/" + gts.C + strings.Replace(fmt.Sprintf("%+v", gts.L), "map", "", -1) + ".csv")
				if err != nil {
					panic(err)
				}

				w := csv.NewWriter(out)
				w.WriteAll(lines) // calls Flush internally

				if err := w.Error(); err != nil {
					log.Fatalln("error writing csv:", err)
				}
				out.Close()
			}
		}

	}
}
