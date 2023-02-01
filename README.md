# obs-switch

A small CLI written in Go to switch program scenes in
[OBS Studio](https://obsproject.com/) using OBS Websocket (included in OBS since
v28).

```console
$ obs-switch 
A remote control for switching scenes in OBS

Usage:
  obs-switch [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  dump-code   Dump the code of this program
  help        Help about any command
  list        List scenes
  scene       Switch program scene
  version     Retrieve and print the OBS and obs-websocket version as json or text

Flags:
  -h, --help              help for obs-switch
  -P, --password string   websocket server password
  -s, --server address    address of obs-websocket to connect to (default "localhost:4455")
  -v, --version           version for obs-switch

Use "obs-switch [command] --help" for more information about a command.

$ obs-switch scene -h
Switch program scene

Usage:
  obs-switch scene [-t | -tb | sceneNumber]

Examples:
    Switch to the first scene: obs-switch scene 0
     Switch one scene forward: obs-switch scene -t
Go back to the previous scene: obs-switch scene -tb

Flags:
  -b, --backwards   use with -t, toggle backwards instead of forward
  -h, --help        help for scene
  -t, --toggle      toggle through all scenes

Global Flags:
  -P, --password string   websocket server password
  -s, --server address    address of obs-websocket to connect to (default "localhost:4455")
```

## Install

```console
$ go install github.com/sa6mwa/obs-switch@latest
```

Or use the `Makefile`...

```console
$ make
go build -o obs-switch -ldflags=-s .

$ sudo make install
install obs-switch /usr/local/bin/obs-switch
```
