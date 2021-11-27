package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	netUrl "net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	"github.com/vega-project/ccb-operator-cli/pkg/config"

	calculationsv1 "github.com/vega-project/ccb-operator/pkg/apis/calculations/v1"
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

	var data *calculationsv1.CalculationList
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}
	output(data)
	return nil
}

func getCalculationID() error {
	args := os.Args
	calcID := args[len(args)-1]

	body, responseError, err := request("GET", globalConfig.APIURL+"/calculation/"+calcID, bytes.NewBuffer(nil))
	if err != nil {
		return err
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	var calc *calculationsv1.Calculation
	if err := json.Unmarshal(body, &calc); err != nil {
		return err
	}

	output(calc)
	return nil
}

func getCalculationResult() error {
	u, err := netUrl.Parse(globalConfig.APIURL)
	if err != nil {
		return err
	}

	u.Path = "/calculations/results"

	q := u.Query()
	q.Set("teff", fmt.Sprintf("%0.1f", teff))
	q.Set("logG", fmt.Sprintf("%0.2f", logG))
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), bytes.NewBuffer(nil))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", globalConfig.Token))
	req.Header.Add("Accept", "application/json")

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fileNameHeader := resp.Header.Get("Content-Disposition")
	if fileNameHeader == "" {
		return fmt.Errorf("couldn't retrieve Content-Disposition header")
	}
	splitFileNameHeader := strings.Split(fileNameHeader, "=")

	defaultPath, err := config.GetPathToCalculation(path, splitFileNameHeader[1])
	if err != nil {
		return err
	}

	err = config.CreateAndWriteFile(body, defaultPath)
	if err != nil {
		return err
	}

	logrus.Info("The calculations were downloaded into: ", defaultPath)

	return nil
}

func getCalculationResultByID() error {
	args := os.Args
	calcID := args[len(args)-1]

	if len(os.Args) != 4 {
		return fmt.Errorf("not enough arguments to use `get results`")
	}

	req, err := http.NewRequest(http.MethodGet, globalConfig.APIURL+"/calculations/results/"+calcID, bytes.NewBuffer(nil))
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", globalConfig.Token))
	req.Header.Add("Accept", "application/json")

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fileNameHeader := resp.Header.Get("Content-Disposition")
	if fileNameHeader == "" {
		return fmt.Errorf("couldn't retrieve Content-Disposition header")
	}
	splitFileNameHeader := strings.Split(fileNameHeader, "=")

	defaultPath, err := config.GetPathToCalculation(path, splitFileNameHeader[1])
	if err != nil {
		return err
	}

	err = config.CreateAndWriteFile(body, defaultPath)
	if err != nil {
		return err
	}

	logrus.Info("The calculations were downloaded into: ", defaultPath)

	return nil
}

func createCalculation() error {
	if teff == 0 || logG == 0 {
		return fmt.Errorf("--teff and --logG must specified together")
	}

	jsonData := fmt.Sprintf("{\"teff\": \"%0.1f\", \"logG\": \"%0.1f\"}", teff, logG)
	body, responseError, err := request("POST", globalConfig.APIURL+"/calculations/create", bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return err
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	var calc *calculationsv1.Calculation
	if err := json.Unmarshal(body, &calc); err != nil {
		return err
	}

	fmt.Printf("Calculation created: %s\n", calc.Name)
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

	switch v := iface.(type) {
	case *calculationsv1.Calculation:
		t.AppendRows([]table.Row{{0, v.Name, v.Spec.Teff, v.Spec.LogG, v.Phase, v.Assign}})
	case *calculationsv1.CalculationList:
		for i, calc := range v.Items {
			t.AppendRows([]table.Row{{i + 1, calc.Name, fmt.Sprintf("%0.1f", calc.Spec.Teff), fmt.Sprintf("%0.2f", calc.Spec.LogG), calc.Phase, calc.Assign}})
		}
		t.AppendSeparator()
		t.AppendFooter(table.Row{"Total", len(v.Items), "", ""})
	}
	t.Render()
}
