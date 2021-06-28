package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	calculationsv1 "github.com/vega-project/ccb-operator/pkg/apis/calculations/v1"

	"github.com/vega-project/ccb-operator-cli/pkg/config"
)

type errorResponse struct {
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

var globalConfig *config.Config

func initializeConfig(cliCtx *cli.Context) error {
	configPath, err := config.GetDefaultConfigPath()
	if err != nil {
		logrus.WithError(err).Fatal("couldn't get configuration path")
	}

	globalConfig, err = config.LoadConfig(configPath)
	if err != nil {
		logrus.WithError(err).Fatal("coudln't load the configuration file")
	}
	return nil
}

func login(c *cli.Context) error {
	url := c.Generic("url")
	token := c.Generic("token")

	if token == "" || url == "" {
		return fmt.Errorf("--token and --url must specified together")
	}

	cfg := config.Config{
		APIURL: fmt.Sprintf("%s", url),
		Token:  fmt.Sprintf("%s", token),
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %v", err)
	}

	user, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
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

func getCalculations(c *cli.Context) error {
	body, responseError, err := request("GET", globalConfig.APIURL+"/calculations", bytes.NewBuffer(nil))
	if err != nil {
		logrus.WithError(err).Fatal("error while perfoming the http request")
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	var data *calculationsv1.CalculationList
	if err := json.Unmarshal(body, &data); err != nil {
		logrus.WithError(err).Fatal("Error unmarshaling json response")
	}
	output(data)
	return nil
}

func getCalculationID(c *cli.Context) error {
	args := os.Args
	calcID := args[len(args)-1]

	body, responseError, err := request("GET", globalConfig.APIURL+"/calculation/"+calcID, bytes.NewBuffer(nil))
	if err != nil {
		logrus.WithError(err).Fatal("error while perfoming the http request")
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	var calc *calculationsv1.Calculation
	if err := json.Unmarshal(body, &calc); err != nil {
		logrus.WithError(err).Fatal("Error unmarshaling json response")
	}

	output(calc)
	return nil
}

func getCalculationResult(c *cli.Context) error {
	teff := c.Float64("teff")
	logG := c.Float64("logG")
	path := c.String("results-download-path")

	resp, err := http.Get(globalConfig.APIURL + "/calculations/results?teff=" + fmt.Sprintf("%0.1f", teff) + "&logg=" + fmt.Sprintf("%0.2f", logG))
	if err != nil {
		logrus.WithError(err).Fatal("Couldn't retrieve the data from headers.")
	}

	fileNameHeader := resp.Header.Get("Content-Disposition")
	if fileNameHeader == "" {
		return fmt.Errorf("couldn't retrieve Content-Disposition header")
	}
	splitFileNameHeader := strings.Split(fileNameHeader, "=")

	defaultPath, err := config.GetPathToCalculation(path, splitFileNameHeader[1])
	if err != nil {
		logrus.WithError(err).Fatal("Couldn't get the path user added to download the calculations.")
	}

	body, responseError, err := request("GET", globalConfig.APIURL+"/calculations/results?teff="+fmt.Sprintf("%0.1f", teff)+"&logg="+fmt.Sprintf("%0.2f", logG), bytes.NewBuffer(nil))
	if err != nil {
		logrus.WithError(err).Fatal("error while perfoming the http request")
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	err = config.CreateAndWriteFile(body, defaultPath)
	if err != nil {
		logrus.WithError(err).Fatal("Something went wrong while creating or writing to a file.")
	}

	logrus.Info("The calculations were downloaded into: ", defaultPath)

	return nil
}

func getCalculationResultByID(c *cli.Context) error {
	args := os.Args
	calcID := args[len(args)-1]

	path := c.String("results-download-path")

	resp, err := http.Get(globalConfig.APIURL + "/calculations/results/" + calcID)
	if err != nil {
		return fmt.Errorf("couldn't retrieve Content-Disposition header")
	}

	fileNameHeader := resp.Header.Get("Content-Disposition")
	if fileNameHeader == "" {
		return fmt.Errorf("couldn't retrieve Content-Disposition header")
	}
	splitFileNameHeader := strings.Split(fileNameHeader, "=")

	defaultPath, err := config.GetPathToCalculation(path, splitFileNameHeader[1])
	if err != nil {
		logrus.WithError(err).Fatal("Couldn't get the path user added to download the calculations.")
	}

	body, responseError, err := request("GET", globalConfig.APIURL+"/calculations/results/"+calcID, bytes.NewBuffer(nil))
	if err != nil {
		logrus.WithError(err).Fatal("error while perfoming the http request")
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	err = config.CreateAndWriteFile(body, defaultPath)
	if err != nil {
		logrus.WithError(err).Fatal("Something went wrong while creating or writing to a file.")
	}

	logrus.Info("The calculations were downloaded into: ", defaultPath)

	return nil
}

func createCalculation(c *cli.Context) error {
	teff := c.Float64("teff")
	logG := c.Float64("logG")

	if teff == 0 || logG == 0 {
		return fmt.Errorf("--teff and --logG must specified together")
	}

	jsonData := fmt.Sprintf("{\"teff\": \"%0.1f\", \"logG\": \"%0.1f\"}", teff, logG)
	body, responseError, err := request("POST", globalConfig.APIURL+"/calculations/create", bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		logrus.WithError(err).Fatal("error while perfoming the http request")
	} else if responseError != nil {
		logrus.WithFields(logrus.Fields{"message": responseError.Message, "status_code": responseError.StatusCode}).Fatal("errors occurred")
	}

	var calc *calculationsv1.Calculation
	if err := json.Unmarshal(body, &calc); err != nil {
		logrus.WithError(err).Fatal("Error unmarshaling json response")
	}

	fmt.Printf("Calculation created: %s", calc.Name)
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
		return nil, nil, fmt.Errorf("Error on response: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("Error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errorResponse *errorResponse
		if err := json.Unmarshal(body, &errorResponse); err != nil {
			return nil, nil, fmt.Errorf("Error unmarshaling json response: %v", err)
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

func main() {
	app := &cli.App{
		Name:  "ccboc",
		Usage: "CLI tool for integrating with the active calculations by communicating with the API server (see https://github.com/vega-project/ccb-operator/tree/master/cmd/apiserver)",
		UsageText: `
	ccboc --teff=10000 --logG=4.0 create (Creates a calculation with teff=10000 and LogG=4.0)
	ccboc get calculation calc-1881i9dh5zvnllip (Gets the calculation with id 'calc-1881i9dh5zvnllip')
	ccboc get calculations [Get all active calculations] 
			`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "url",
				Aliases: []string{"u"},
				Usage:   "URL of the API server.",
			},
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "Token for authenticating to the API server.",
			},
			&cli.Float64Flag{
				Name:  "teff",
				Usage: "specifies the teff value when creating a calculation.",
			},
			&cli.Float64Flag{
				Name:  "logG",
				Usage: "specifies the logG value when creating a calculation.",
			},
			&cli.StringFlag{
				Name:  "results-download-path",
				Usage: "Specified path to the downloaded calculation results.",
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "login",
				Aliases: []string{"l"},
				Usage:   "Login to api server using the provided url and token. Also it generates the configuration file (default path is $HOME/.config/ccbo/config)",
				Action:  login,
			},
			{
				Name:    "get",
				Aliases: []string{"g"},
				Usage:   "Usage: `get calculation CALCID`, or `get calculations` to get all active calculations",
				Before:  initializeConfig,
				Subcommands: []*cli.Command{
					{
						Name:    "calculation",
						Aliases: []string{"calc"},
						Usage:   "get calculation id",
						Action:  getCalculationID,
					},

					{
						Name:    "calculations",
						Aliases: []string{"calcs"},
						Usage:   "",
						Action:  getCalculations,
					},

					{
						Name:    "results",
						Aliases: []string{"res"},
						Usage:   "get results of calculations",
						Action: func(c *cli.Context) error {
							var err error
							if c.Float64("teff") == 0 || c.Float64("logG") == 0 {
								err = getCalculationResultByID(c)
							} else {
								err = getCalculationResult(c)
							}
							return err
						},
					},
				},
			},
			{
				Name:    "create",
				Aliases: []string{"c"},
				Usage:   "Creates a calculation using the values from --teff and --logG flags",
				Before:  initializeConfig,
				Action:  createCalculation,
			},
		},
	}
	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	err := app.Run(os.Args)
	if err != nil {
		logrus.WithError(err).Fatal("errors occurred")
	}
}
