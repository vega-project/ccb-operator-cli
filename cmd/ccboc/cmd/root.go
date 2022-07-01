package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	url            string
	token          string
	bulkFile       string
	workerPoolName string

	rootCmd = &cobra.Command{
		Use:   "ccboc",
		Short: "CLI tool for integrating with the active calculations by communicating with the API server (see https://github.com/vega-project/ccb-operator/tree/master/cmd/apiserver)",
		Long: "Examples of usage:" + "\n" +
			"ccboc get calculation calc-1881i9dh5zvnllip (Gets the calculation with id='calc-1881i9dh5zvnllip')\n" +
			"ccboc get calculations (Gets all active calculations)\n" +
			"ccboc get bulks (Gets all calculation bulks)\n" +
			"ccboc get bulk bulk-2bw55pr5p37dasdl (Gets the calculation bulk with id='2bw55pr5p37dasdl')\n" +
			"ccboc get workerpools (Gets all the workerpools)\n" +
			"ccboc create bulk --bulk-file=<bulk-input-file.json> (Creates a calculation bulk from a file)\n" +
			"ccboc create workerpool --name=vega-project\n" +
			"ccboc get workerpool workerpool-vega-project\n" +
			"ccboc delete workerpool workerpool-vega-project\n" +
			"ccboc delete bulk bulk-vega-project",
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

	deleteCmd = &cobra.Command{
		Use:              "delete",
		Short:            "Delete an object - bulk/workerpool",
		TraverseChildren: true,
	}

	deleteWorkerPoolCmd = &cobra.Command{
		Use:   "workerpool",
		Short: "Delete a workerpool by an ID",
		Run: func(cmd *cobra.Command, args []string) {
			initializeConfig()
			if err := deleteWorkerPool(); err != nil {
				logrus.WithError(err).Fatal("delete workerpool <workerpool-id> command failed")
			}
		},
	}

	deleteCalculationBulkCmd = &cobra.Command{
		Use:   "bulk",
		Short: "Delete a calculation bulk by an ID",
		Run: func(cmd *cobra.Command, args []string) {
			initializeConfig()
			if err := deleteCalculationBulk(); err != nil {
				logrus.WithError(err).Fatal("delete calculation bulk <bulk-id> command failed")
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

	workerPoolCmd = &cobra.Command{
		Use:   "workerpool",
		Short: "Get a workerpool by an ID",
		Run: func(cmd *cobra.Command, args []string) {
			initializeConfig()
			if err := getWorkerPoolByName(); err != nil {
				logrus.WithError(err).Fatal("get workerpool by name command failed")
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
		Short:            "Create a bulk/workerpool object in the cluster.",
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

	createWorkerPoolCmd = &cobra.Command{
		Use:   "workerpool",
		Short: "Creates a workerpool in the cluster with custom name",
		Run: func(cmd *cobra.Command, args []string) {
			initializeConfig()
			if workerPoolName == "" {
				logrus.Fatal("name of the workerpool was not specified")
			}
			var err error
			if err = createWorkerPool(); err != nil {
				logrus.WithError(err).Fatal("create a workerpool command failed")
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

	getCmd.AddCommand(workerPoolsCmd)

	getCmd.AddCommand(workerPoolCmd)

	rootCmd.AddCommand(createCmd)

	createCmd.AddCommand(createBulkCmd)
	createBulkCmd.Flags().StringVar(&bulkFile, "bulk-file", "", "File in .json format to create a calculation bulk.")

	createCmd.AddCommand(createWorkerPoolCmd)
	createWorkerPoolCmd.Flags().StringVar(&workerPoolName, "name", "", "Name of the workerpool.")

	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteWorkerPoolCmd)
	deleteCmd.AddCommand(deleteCalculationBulkCmd)

}
