# ccboc
CLI tool for ccb-operator


```
Examples of usage:
ccboc get calculation calc-1881i9dh5zvnllip (Gets the calculation with id='calc-1881i9dh5zvnllip')
ccboc get calculations (Gets all active calculations)
ccboc --teff=10000 --logG=4.0 get results (Downloads the result of a calculation with teff=10000 and LogG=4.0)
ccboc get results calc-1881i9dh5zvnllip (Downloads the result of a calculation with id='calc-1881i9dh5zvnllip')
ccboc get bulks (Gets all calculation bulks)
ccboc get bulk bulk-2bw55pr5p37dasdl (Gets the calculation bulk with id='2bw55pr5p37dasdl')
ccboc get workerpools (Gets all the workerpools)
ccboc create bulk --bulk-file=<bulk-input-file.json> (Creates a calculation bulk from a file)

Usage:
  ccboc [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  create      Create a bulk in the cluster.
  get         Get an object - calculation/bulk/workerpool.
  help        Help about any command
  login       Login to api server using the provided url and token. Also it generates the configuration file (default path is $HOME/.config/ccbo/config)

Flags:
  -h, --help   help for ccboc

Use "ccboc [command] --help" for more information about a command.
```
