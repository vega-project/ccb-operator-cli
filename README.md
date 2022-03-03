# ccboc
CLI tool for ccb-operator


```
NAME:
   ccboc - CLI tool for integrating with the active calculations by communicating with the API server (see https://github.com/vega-project/ccb-operator/tree/master/cmd/apiserver)

USAGE:
   
  ccboc --teff=10000 --logG=4.0 create (Creates a calculation with teff=10000 and LogG=4.0)
  ccboc get calculation calc-1881i9dh5zvnllip (Gets the calculation with id='calc-1881i9dh5zvnllip')
  ccboc get calculations (Gets all active calculations)
  ccboc --teff=10000 --logG=4.0 get results (Downloads the result of a calculation with teff=10000 and LogG=4.0)
  ccboc get results calc-1881i9dh5zvnllip (Downloads the result of a calculation with id='calc-1881i9dh5zvnllip')
  ccboc get bulks (Gets all calculation bulks)
  ccboc get bulk bulk-2bw55pr5p37dasdl (Gets the calculation bulk with id='2bw55pr5p37dasdl')
  ccboc get workerpools (Gets all the workerpools)
  ccboc create bulk --bulk-file=<bulk-input-file.json> (Creates a calculation bulk from a file)
      

COMMANDS:
   create, c  Creates a calculation using the values from --teff and --logG flags
   get, g     Usage: `get calculation CALCID`, or `get calculations` to get all active calculations, or `get results CALCID` to download a calculation's result by CALCID, or `TEFF LOGG get results`, or `get bulks` to get all calculation bulks, or `get bulk BULKID` to get a calculation bulk by BULKID., or `get workerpools` to get all workerpools
   login, l   Login to api server using the provided url and token. Also it generates the configuration file (default path is $HOME/.config/ccbo/config)
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --logG value             specifies the logG value when creating a calculation. (default: 0)
   --teff value             specifies the teff value when creating a calculation. (default: 0)
   --token value, -t value  Token for authenticating to the API server.
   --url value, -u value    URL of the API server.
   --results-download-path  Specifies a path to the downloaded calculation results
   --bulk-file              Specify a .json file to build a calculation bulk from
   --help, -h               show help (default: false)
```
