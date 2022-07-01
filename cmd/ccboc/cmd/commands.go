package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	"github.com/vega-project/ccb-operator-cli/pkg/config"

	bulkv1 "github.com/vega-project/ccb-operator/pkg/apis/calculationbulk/v1"
	calculationsv1 "github.com/vega-project/ccb-operator/pkg/apis/calculations/v1"
	workersv1 "github.com/vega-project/ccb-operator/pkg/apis/workers/v1"
)

type errorResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

var globalConfig *config.Config

func initializeConfig() error {
	configPath, err := config.GetDefaultConfigPath()
	if err != nil {
		return err
	}

	globalConfig, err = config.LoadConfig(configPath)
	if err != nil {
		return err
	}
	return nil
}

func login() error {
	if token == "" || url == "" {
		return fmt.Errorf("--token and --url must specified together")
	}

	cfg := config.Config{
		APIURL: url,
		Token:  token,
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %v", err)
	}

	user, err := user.Current()
	if err != nil {
		return fmt.Errorf("getting current user failed: %v", err)
	}
	home := user.HomeDir

	data, err := json.MarshalIndent(cfg, "", "")
	if err != nil {
		return fmt.Errorf("couldn't marthal the configuration file: %v", err)
	}

	path := filepath.Join(home, config.DefaultConfigPath)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			return fmt.Errorf("couldn't create dir: %v", err)
		}
	}

	configFilePath := filepath.Join(home, config.DefaultConfigPath, config.DefaultConfigFileName)
	if err := ioutil.WriteFile(configFilePath, data, 0777); err != nil {
		return fmt.Errorf("couldn't write to file: %v", err)
	}

	fmt.Printf("Configuration file generated at %s\n", configFilePath)
	return nil
}

func getCalculations() error {
	body, responseError, err := request("GET", globalConfig.APIURL+"/calculations", bytes.NewBuffer(nil))
	if err != nil {
		return err
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	var response map[string]json.RawMessage
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	var calcList *calculationsv1.CalculationList
	if err := json.Unmarshal(response["data"], &calcList); err != nil {
		return err
	}

	output(calcList)

	return nil
}

func getCalculationByID() error {
	args := os.Args
	calcID := args[len(args)-1]

	body, responseError, err := request("GET", globalConfig.APIURL+"/calculation/"+calcID, bytes.NewBuffer(nil))
	if err != nil {
		return err
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	var response map[string]json.RawMessage
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	var calc *calculationsv1.Calculation
	if err := json.Unmarshal(response["data"], &calc); err != nil {
		return err
	}

	output(calc)

	return nil
}

func getCalculationBulkByID() error {
	args := os.Args
	bulkID := args[len(args)-1]

	body, responseError, err := request("GET", globalConfig.APIURL+"/bulk/"+bulkID, bytes.NewBuffer(nil))
	if err != nil {
		return err
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	var response map[string]json.RawMessage
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	var bulk *bulkv1.CalculationBulk
	if err := json.Unmarshal(response["data"], &bulk); err != nil {
		return err
	}

	output(bulk)

	return nil
}

func getCalculationBulks() error {
	body, responseError, err := request("GET", globalConfig.APIURL+"/bulks", bytes.NewBuffer(nil))
	if err != nil {
		return err
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	var response map[string]json.RawMessage
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	var calcBulkList *bulkv1.CalculationBulkList
	if err := json.Unmarshal(response["data"], &calcBulkList); err != nil {
		return err
	}

	output(calcBulkList)

	return nil
}

func deleteCalculationBulk() error {
	args := os.Args
	calculationBulkName := args[len(args)-1]

	body, responseError, err := request("DELETE", fmt.Sprintf(globalConfig.APIURL+"/bulks/delete/%s", calculationBulkName), bytes.NewBuffer(nil))
	if err != nil {
		return err
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occured")
	}

	var response map[string]json.RawMessage
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	if string(response["status_code"]) == "200" {
		logrus.Infof("Deleted calculation bulk %s successfully", calculationBulkName)
	} else {
		logrus.Infof("Couldn't delete the calculation bulk named %s", calculationBulkName)
	}

	return nil
}

func createWorkerPool() error {
	body, responseError, err := request("POST", fmt.Sprintf(globalConfig.APIURL+"/workerpool/create?name=%s", workerPoolName), bytes.NewBuffer(nil))
	if err != nil {
		return err
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occured")
	}

	var response map[string]json.RawMessage
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	var workerPool *workersv1.WorkerPool
	if err := json.Unmarshal(response["data"], &workerPool); err != nil {
		return err
	}

	logrus.Infof("Created workerpool %s successfully", workerPool.Name)

	return nil
}

func getWorkerPools() error {
	body, responseError, err := request("GET", globalConfig.APIURL+"/workerpools", bytes.NewBuffer(nil))
	if err != nil {
		return err
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	var response map[string]json.RawMessage
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	var workerPoolsList *workersv1.WorkerPoolList
	if err := json.Unmarshal(response["data"], &workerPoolsList); err != nil {
		return err
	}

	output(workerPoolsList)

	return nil
}

func getWorkerPoolByName() error {
	args := os.Args
	workerPoolName := args[len(args)-1]

	body, responseError, err := request("GET", globalConfig.APIURL+"/workerpool/"+workerPoolName, bytes.NewBuffer(nil))
	if err != nil {
		return err
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	var response map[string]json.RawMessage
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	var workerPool *workersv1.WorkerPool
	if err := json.Unmarshal(response["data"], &workerPool); err != nil {
		return err
	}

	output(workerPool)

	return nil
}

func deleteWorkerPool() error {
	args := os.Args
	workerPoolName := args[len(args)-1]

	body, responseError, err := request("DELETE", fmt.Sprintf(globalConfig.APIURL+"/workerpools/delete/%s", workerPoolName), bytes.NewBuffer(nil))
	if err != nil {
		return err
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occured")
	}

	var response map[string]json.RawMessage
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	if string(response["status_code"]) == "200" {
		logrus.Infof("Deleted workerpool %s successfully", workerPoolName)
	} else {
		logrus.Infof("Couldn't delete the workerpool named %s", workerPoolName)
	}

	return nil
}

func createCalculationBulk() error {
	fileBytes, err := ioutil.ReadFile(bulkFile)
	if err != nil {
		logrus.WithError(err).Fatal("Couldn't open the input .json file to create a calculation bulk.")
	}

	body, responseError, err := request("POST", globalConfig.APIURL+"/bulk/create", bytes.NewBuffer(fileBytes))
	if err != nil {
		return err
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	var response map[string]json.RawMessage
	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	var bulk *bulkv1.CalculationBulk
	if err := json.Unmarshal(fileBytes, &bulk); err != nil {
		logrus.WithError(err).Fatal("Couldn't unmarshal the contents of the input .json file.")
	}

	fmt.Println("Calculation bulk created")

	return nil
}

func request(method, endpoint string, buffer *bytes.Buffer) ([]byte, *errorResponse, error) {
	req, err := http.NewRequest(method, endpoint, buffer)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", globalConfig.Token))
	req.Header.Add("Accept", "application/json")

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("error on response: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse *errorResponse
		if err := json.Unmarshal(body, &errorResponse); err != nil {
			return nil, nil, fmt.Errorf("error unmarshaling json response: %v", err)
		}
		return nil, errorResponse, nil
	}

	return body, nil, nil
}

func output(iface interface{}) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"#", "Name", "Teff", "LogG", "Phase", "Worker"})

	bulksWriter := table.NewWriter()
	bulksWriter.SetOutputMirror(os.Stdout)
	bulksWriter.AppendHeader(table.Row{"#", "Name", "Teff", "LogG", "Phase"})

	workersWriter := table.NewWriter()
	workersWriter.SetOutputMirror(os.Stdout)
	workersWriter.AppendHeader(table.Row{"#", "Name", "Workers"})

	workerPoolWriter := table.NewWriter()
	workerPoolWriter.SetOutputMirror(os.Stdout)
	workerPoolWriter.AppendHeader(table.Row{"Name", "Workers"})

	switch v := iface.(type) {
	case *calculationsv1.Calculation:
		t.AppendRows([]table.Row{{0, v.Name, v.Spec.Teff, v.Spec.LogG, v.Phase, v.Assign}})
		t.Render()
	case *calculationsv1.CalculationList:
		for i, calc := range v.Items {
			t.AppendRows([]table.Row{{i + 1, calc.Name, fmt.Sprintf("%0.1f", calc.Spec.Teff), fmt.Sprintf("%0.2f", calc.Spec.LogG), calc.Phase, calc.Assign}})
		}
		t.AppendSeparator()
		t.AppendFooter(table.Row{"Total", len(v.Items), "", ""})
		t.Render()
	case *bulkv1.CalculationBulk:
		for _, c := range v.Calculations {
			bulksWriter.AppendRows([]table.Row{{0, v.Name, fmt.Sprintf("%0.1f", c.Params.Teff), fmt.Sprintf("%0.2f", c.Params.LogG), c.Phase}})
		}
		bulksWriter.AppendSeparator()
		bulksWriter.AppendFooter(table.Row{"Total", len(v.Calculations), "", ""})
		bulksWriter.Render()
	case *bulkv1.CalculationBulkList:
		for i, bulk := range v.Items {
			bulksWriter.AppendRows([]table.Row{{i + 1, bulk.Name, "", "", ""}})
			for _, c := range bulk.Calculations {
				bulksWriter.AppendRows([]table.Row{{"", "", fmt.Sprintf("%0.1f", c.Params.Teff), fmt.Sprintf("%0.2f", c.Params.LogG), c.Phase}})
			}
		}
		bulksWriter.AppendSeparator()
		bulksWriter.AppendFooter(table.Row{"Total", len(v.Items), "", ""})
		bulksWriter.Render()
	case *workersv1.WorkerPoolList:
		for i, worker := range v.Items {
			workersWriter.AppendRows([]table.Row{{i + 1, worker.Name, ""}})
			for _, w := range worker.Spec.Workers {
				workersWriter.AppendRows([]table.Row{{"", "", w.Name}})
			}
		}
		workersWriter.AppendSeparator()
		workersWriter.AppendFooter(table.Row{"Total", len(v.Items), "", ""})
		workersWriter.Render()
	case *workersv1.WorkerPool:
		workerPoolWriter.AppendSeparator()
		workerPoolWriter.AppendFooter(table.Row{v.Name, ""})
		for _, w := range v.Spec.Workers {
			workerPoolWriter.AppendRows([]table.Row{{"", w.Name}})
		}
		workerPoolWriter.Render()
	}
}
