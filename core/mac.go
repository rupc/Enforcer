package core

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/rupc/Enforcer/atypes"
	"github.com/rupc/Enforcer/util"
)

type MetricAnalysisResult struct {
}

type MetricAnalysis struct {
	// CommitDistance float64
	Config MetricAnalysisConfig
	logger *flogging.FabricLogger
}

type ScriptStruct struct {
	DataFiles   []string `json:"data_files"`
	ReportFile  string   `json:"report_file"`
	PlotFile    string   `json:"plot_file"`
	LeveldbPath string   `json:"leveldb_path"`
	RedisDBPath string   `json:"redis_db_path"`

	StartBlkNum string `json:"start_blk_num"`
	EndBlkNum   string `json:"end_blk_num"`
	WindowSize  string `json:"window_size"`
}

type MetricAnalysisConfig struct {
	AnalysisRequestChannel chan *atypes.AnalysisRequest
	FetchMethod            string   // from prom or leveldb
	FetchAddress           string   // source location
	ReportAddress          string   // SDK server api address
	Period                 uint32   // Unit: Seconds
	TargetMetrics          []string // metrics names, e.g., "commit_distance"

	ScriptConfigFile string
	LeveldbScript    string
	LeveldbPath      string
	RedisDBPath      string
}

func InitializeMetricAnalysis(config MetricAnalysisConfig) *MetricAnalysis {
	ma := &MetricAnalysis{
		Config: config,
		logger: flogging.MustGetLogger("ma"),
	}
	return ma
}

func (mac *MetricAnalysis) Run() {
	for {
		r := <-mac.Config.AnalysisRequestChannel
		mac.logger.Debugf("Start analysis request with", r)

		// analysisType := r.AnalysisType
		// startBlkNum := r.StartBlkNum
		// endBlkNum := r.EndBlkNum
		// windowSize := r.WindowSize

		// do real analysis really..
		// 1. update scripts/config.json

		script := ScriptStruct{
			// DataFiles:   dataFiles,
			ReportFile: util.CreateTmpFile("", "/tmp", "report.*.json"),
			// PlotFile:    CreateTmpFile("", "/tmp", "plot.*.svg"),
			// LeveldbPath: mac.Config.LeveldbPath,
			RedisDBPath: mac.Config.RedisDBPath,
			StartBlkNum: r.StartBlkNum,
			EndBlkNum:   r.EndBlkNum,
			WindowSize:  r.WindowSize,
		}

		// Run python script simply
		configContent, _ := json.MarshalIndent(script, "", " ")
		_ = ioutil.WriteFile(mac.Config.ScriptConfigFile, configContent, 0644)
		cmd := exec.Command("./scripts/" + "analysis.py")
		// cmd.Dir = "scripts"

		if err := cmd.Run(); err != nil {
			log.Fatalln("error", err)
			return
		}
		// 2. execute python script

		// 3. wait for result

		// 4. return the result to
	}
}

func (mac *MetricAnalysis) Run_deprecated() {
	for {
		time.Sleep(time.Duration(mac.Config.Period) * time.Second)

		if mac.Config.FetchMethod == "prom" {
			// "http://141.223.121.69:9090" + "/api/v1/query"
			apiURL := mac.Config.FetchAddress + "/api/v1/query"
			queryStrings := []string{"commit_distance[4d]", "up[1m]"}
			dataFiles := make([]string, len(queryStrings))
			configFile := "config.json"

			for _, query := range queryStrings {
				metricData := GetMetricData(apiURL, query)
				dataFileName := util.CreateTmpFile(metricData, "/tmp", "metric.*.json")
				dataFiles = append(dataFiles, dataFileName)
			}

			// 1. Fetch metric data from prometheus server
			// fmt.Println(metricData)

			script := ScriptStruct{
				DataFiles:   dataFiles,
				ReportFile:  util.CreateTmpFile("", "/tmp", "report.*.json"),
				PlotFile:    util.CreateTmpFile("", "/tmp", "plot.*.svg"),
				LeveldbPath: mac.Config.LeveldbPath,
			}

			for _, v := range script.DataFiles {
				defer os.Remove(v)
			}
			defer os.Remove(script.ReportFile)
			defer os.Remove(script.PlotFile)

			configContent, _ := json.MarshalIndent(script, "", " ")
			_ = ioutil.WriteFile(configFile, configContent, 0644)

			// 3. Run python script(refine, analyze, report, plot)
			cmd := exec.Command("./scripts/metric.py")
			// cmd.Dir = "scripts"

			if err := cmd.Run(); err != nil {
				log.Fatalln("error", err)
				return
			}

			// Get analysis result from python script
			// 4. Print report
			type ResultType struct {
				Score float32 `json:"score"`
			}
			r := ResultType{}
			jsonFile, err := os.Open(script.ReportFile)
			if err != nil {
				log.Fatalln("error", err)
			}
			defer jsonFile.Close()

			b, _ := ioutil.ReadAll(jsonFile)
			json.Unmarshal(b, &r)
			log.Printf("%+v", r)

			// score := r.Score

			// Send score to sdk server
			rbytes, _ := json.Marshal(r)
			resp, err := http.Post(mac.Config.ReportAddress, "application/json", bytes.NewBuffer(rbytes))
			if err != nil {
				mac.logger.Infof("Sending stateOutput failed", err)
			}
			mac.logger.Infof(resp.Status)

			// Refine metrics for easily handling in python script
			// Create a file in /tmp/XYZ.json or /tmp/XYZ.txt
			// Write refined metrics to a created file
			// Execute python script with a file name
			// Wait execution completion of python script
			// Report the result to SDK server

		} else if mac.Config.FetchMethod == "leveldb" {

			script := ScriptStruct{
				// DataFiles:   dataFiles,
				// ReportFile:  CreateTmpFile("", "/tmp", "report.*.json"),
				// PlotFile:    CreateTmpFile("", "/tmp", "plot.*.svg"),
				LeveldbPath: mac.Config.LeveldbPath,
			}

			configContent, _ := json.MarshalIndent(script, "", " ")
			_ = ioutil.WriteFile(mac.Config.ScriptConfigFile, configContent, 0644)
			// Run python script simply
			cmd := exec.Command("./scripts/" + mac.Config.LeveldbScript)
			// cmd.Dir = "scripts"

			if err := cmd.Run(); err != nil {
				log.Fatalln("error", err)
				return
			}
		}
	}
}

func GetMetricData(apiURL, query string) string {
	data := url.Values{
		"query": {query},
	}

	response, err := http.PostForm(apiURL, data)

	if err != nil {
		log.Fatalln(err)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		//handle read response error
	}

	return string(body)
}

func (mac *MetricAnalysis) DoMetricAnalysis(block *atypes.DistilledBlock) *MetricAnalysisResult {
	r := &MetricAnalysisResult{}
	// doFairnessAnalysis by malicious BSP
	// doDetectMalicious by malicious Peer
	// doPerformanceAnomaly by whole system
	return r
}
