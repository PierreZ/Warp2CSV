package main

import (
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type GTSS []struct {
	C string      `json:"c"`
	V [][]float64 `json:"v"`
}

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

		if !strings.Contains(f.Name(), ".mc2") {
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

		defer resp.Body.Close()

		var rawgtss GTSS

		err = json.NewDecoder(resp.Body).Decode(&rawgtss)
		if err != nil {
			panic(err)
		}

		var lines [][]string
		lines = append(lines, []string{"timestamp(in sec)", "value"})

		for _, gts := range rawgtss {

			for _, tick := range gts.V {
				ts := strconv.FormatFloat(tick[0]/1000.0/1000.0, 'E', -1, 64)
				value := strconv.FormatFloat(tick[1], 'E', -1, 64)
				lines = append(lines, []string{ts, value})
			}

			out, err := os.Create(dir + "/" + f.Name() + "-" + gts.C + ".csv")
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
