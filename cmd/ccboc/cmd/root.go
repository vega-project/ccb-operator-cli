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
			"ccboc get calculation calc-1881i9dh5zvnllip (Gets the calculation with id='calc-1881i9dh5zvnllip')\n" +
			"ccboc get calculations (Gets all active calculations)\n" +
			"ccboc --teff=10000 --logG=4.0 get results (Downloads the result of a calculation with teff=10000 and LogG=4.0)\n" +
			"ccboc get results calc-1881i9dh5zvnllip (Downloads the result of a calculation with id='calc-1881i9dh5zvnllip')\n" +
			"ccboc get bulks (Gets all calculation bulks)\n" +
			"ccboc get bulk bulk-2bw55pr5p37dasdl (Gets the calculation bulk with id='2bw55pr5p37dasdl')\n" +
			"ccboc get workerpools (Gets all the workerpools)\n" +
			"ccboc create bulk --bulk-file=<bulk-input-file.json> (Creates a calculation bulk from a file)\n",
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
		Short:            "Get an object - calculation/bulk/workerpool.",
		TraverseChildren: true,
	}

	phaseCmd = &cobra.Command{
		Use:   "phase",
		Short: "Get a phase of ",
		Run: func(cmd *cobra.Command, args []string) {
			initializeConfig()
			err := getCalculationPhase()
			if err != nil {
				logrus.WithError(err).Fatal("get phase <bulk-id> command failed")
			}
		},
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

	workerPoolsCmd = &cobra.Command{
		Use:   "workerpools",
		Short: "Get all workerpools.",
		Run: func(cmd *cobra.Command, args []string) {
			initializeConfig()
			err := getWorkerPools()
			if err != nil {
				logrus.WithError(err).Fatal("get workerpools command failed")
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
		Short:            "Create a bulk in the cluster.",
		TraverseChildren: true,
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

	getCmd.AddCommand(phaseCmd)

	getCmd.AddCommand(bulksCmd)

	getCmd.AddCommand(workerPoolsCmd)

	getCmd.AddCommand(resultsCmd)
	resultsCmd.Flags().Float64Var(&teff, "teff", 0.0, "Teff value to download a calculation.")
	resultsCmd.Flags().Float64Var(&logG, "logG", 0.0, "logG value to download a calculation.")
	resultsCmd.Flags().StringVar(&path, "results-download-path", "", "Specified path to download the calculation to.")

	rootCmd.AddCommand(createCmd)

	createCmd.AddCommand(createBulkCmd)
	createBulkCmd.Flags().StringVar(&bulkFile, "bulk-file", "", "File in .json format to create a calculation bulk.")
}
