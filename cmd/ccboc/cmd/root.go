package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	url      string
	token    string
	teff     float64
	logG     float64
	path     string
	bulkFile string

	rootCmd = &cobra.Command{
		Use:   "ccboc",
		Short: "CLI tool for integrating with the active calculations by communicating with the API server (see https://github.com/vega-project/ccb-operator/tree/master/cmd/apiserver)",
		Long: "Examples of usage:" + "\n" +
			"  ccboc --teff=10000 --logG=4.0 create (Creates a calculation with teff=10000 and LogG=4.0)\n" +
			"  ccboc get calculation calc-1881i9dh5zvnllip (Gets the calculation with id 'calc-1881i9dh5zvnllip')\n" +
			"  ccboc get calculations [Get all active calculations]\n",
	}

	loginCmd = &cobra.Command{
		Use:              "login",
		Short:            "Login to api server using the provided url and token. Also it generates the configuration file (default path is $HOME/.config/ccbo/config)",
		TraverseChildren: true,
		Run: func(cmd *cobra.Command, args []string) {
			err := login()
			if err != nil {
				logrus.WithError(err).Fatal("login command failed")
			}
		},
	}

	getCmd = &cobra.Command{
		Use:              "get",
		Short:            "Usage: `get calculation <calc-id>`, or `get calculations` to get all active calculations",
		TraverseChildren: true,
	}

	bulkCmd = &cobra.Command{
		Use:   "bulk",
		Short: "Get calculation bulk by an ID.",
		Run: func(cmd *cobra.Command, args []string) {
			initializeConfig()
			err := getCalculationBulkByID()
			if err != nil {
				logrus.WithError(err).Fatal("get bulk <bulk-id> command failed")
			}
		},
	}

	bulksCmd = &cobra.Command{
		Use:   "bulks",
		Short: "Get all calculation bulks.",
		Run: func(cmd *cobra.Command, args []string) {
			initializeConfig()
			err := getCalculationBulks()
			if err != nil {
				logrus.WithError(err).Fatal("get bulks command failed")
			}
		},
	}

	calculationCmd = &cobra.Command{
		Use:   "calculation",
		Short: "Get calculation by an ID.",
		Run: func(cmd *cobra.Command, args []string) {
			initializeConfig()
			err := getCalculationByID()
			if err != nil {
				logrus.WithError(err).Fatal("get calculation <calc-id> command failed")
			}
		},
	}

	calculationsCmd = &cobra.Command{
		Use:   "calculations",
		Short: "Get all the calculations.",
		Run: func(cmd *cobra.Command, args []string) {
			initializeConfig()
			err := getCalculations()
			if err != nil {
				logrus.WithError(err).Fatal("get calculations command failed")
			}
		},
	}

	createCmd = &cobra.Command{
		Use:              "create",
		Short:            "Creates a calculation/bulks in the cluster.",
		TraverseChildren: true,
	}

	createCalcCmd = &cobra.Command{
		Use:   "calculation",
		Short: "Create a calculation in the cluster with specified logG and teff values.",
		Run: func(cmd *cobra.Command, args []string) {
			initializeConfig()
			err := createCalculation()
			if err != nil {
				logrus.WithError(err).Fatal("create calculation command failed")
			}
		},
	}

	createBulkCmd = &cobra.Command{
		Use:   "bulk",
		Short: "Creates a calculation bulk in the cluster using a .json file.",
		Run: func(cmd *cobra.Command, args []string) {
			initializeConfig()
			var err error
			if bulkFile == "" {
				logrus.WithError(err).Fatal("file to create a calculation bulk not specified")
			}
			err = createCalculationBulk()
			if err != nil {
				logrus.WithError(err).Fatal("create a calculation bulk command failed")
			}
		},
	}

	resultsCmd = &cobra.Command{
		Use:   "results",
		Short: "Downloads a calculation result by teff and logG or by calculation ID to a folder.",
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			initializeConfig()
			if teff == 0 && logG == 0 {
				err = getCalculationResultByID()
				if err != nil {
					logrus.WithError(err).Fatal("get result by id command failed")
				}
			} else if teff != 0 && logG != 0 {
				err = getCalculationResult()
				if err != nil {
					logrus.WithError(err).Fatal("get result by teff and logG command failed")
				}
			} else {
				logrus.WithError(err).Fatal("wrong usage of get results command")
			}
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVar(&url, "url", "", "URL to log into.")
	loginCmd.Flags().StringVar(&token, "token", "", "Token to login with.")

	rootCmd.AddCommand(getCmd)

	getCmd.AddCommand(calculationCmd)

	getCmd.AddCommand(calculationsCmd)

	getCmd.AddCommand(bulkCmd)

	getCmd.AddCommand(bulksCmd)

	getCmd.AddCommand(resultsCmd)
	resultsCmd.Flags().Float64Var(&teff, "teff", 0.0, "Teff value to download a calculation.")
	resultsCmd.Flags().Float64Var(&logG, "logG", 0.0, "logG value to download a calculation.")
	resultsCmd.Flags().StringVar(&path, "results-download-path", "", "Specified path to download the calculation to.")

	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createCalcCmd)
	createCalcCmd.Flags().Float64Var(&teff, "teff", 0.0, "Teff value to create a calculation.")
	createCalcCmd.Flags().Float64Var(&logG, "logG", 0.0, "LogG value to create a calculation.")

	createCmd.AddCommand(createBulkCmd)
	createBulkCmd.Flags().StringVar(&bulkFile, "bulk-file", "", "File in .json format to create a calculation bulk.")
}
