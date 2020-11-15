# VSCode extension for ECAL

## Folder content

- `package.json` - manifest file
- `syntaxes/ecal.tmLanguage.json` - Text mate grammar file
- `language-configuration.json` - language configuration for VSCode

## Build the extention

To build the extention you need `npm` installed.

VSIX file can be build with `npm run package`

## Install the extension

The extention can be installed using a precompiled VSIX file which can be downloaded from here:

https://devt.de/krotik/ecal/releases

## Launch config for ecal projects

```
{
	"version": "0.2.0",
	"configurations": [
		{
			"type": "ecaldebug",
			"request": "launch",
			"name": "Debug ECAL script with ECAL Debug Server",

			"serverURL": "localhost:43806",
            "dir": "${workspaceFolder}",
			"executeOnEntry": true,
			"trace": false,
		}
	]
}
```

- serverURL: URL of the ECAL debug server.
- dir: Root directory for ECAL debug server.
- executeOnEntry: (optional) Execute the ECAL script on entry. If this is set to false then code needs to be manually started from the ECAL debug server console.
- trace: (optional) Enable tracing messages for debug adapter (useful when debugging the debugger).

## Developing the extension

In VSCode the extention can be launched and debugged using the included launch configuration. Press F5 to start a VS Code instance with ECAL support extention form the development code.
