// obs-switch is a small CLI to switch program scenes in OBS Studio using the
// OBS WebSocket API.
package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/scenes"
	"github.com/andreykaipov/goobs/api/typedefs"
	"github.com/spf13/cobra"
)

//go:embed obs-switch.go
var code string

var version = "0.1.0"

var server string
var password string

var rootCmd = &cobra.Command{
	Use:          "obs-switch",
	Version:      version,
	Short:        "A remote control for switching scenes in OBS",
	SilenceUsage: true,
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"obs-version"},
	Short:   "Retrieve and print the OBS and obs-websocket version as json or text",
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := cmd.Flags().GetString("format")
		if err != nil {
			return err
		}
		if !(f == "json" || f == "text") {
			return errors.New("format must be json or text")
		}
		opts := []goobs.Option{}
		if len(password) > 0 {
			opts = []goobs.Option{goobs.WithPassword(password)}
		}
		obs, err := goobs.New(server, opts...)
		if err != nil {
			return err
		}
		defer obs.Disconnect()
		version, err := obs.General.GetVersion()
		if err != nil {
			return err
		}
		switch f {
		case "json":
			obj := struct {
				ObsVersion          string `json:"obs"`
				ObsWebsocketVersion string `json:"websocket"`
			}{
				ObsVersion:          version.ObsVersion,
				ObsWebsocketVersion: version.ObsWebSocketVersion,
			}
			j, err := json.Marshal(obj)
			if err != nil {
				return err
			}
			fmt.Println(string(j))
		case "text":
			fmt.Printf("OBS version: %s\nOBS websocket version: %s\n", version.ObsVersion, version.ObsWebSocketVersion)
		}
		return nil
	},
}

func reverseScenes(s []*typedefs.Scene) []*typedefs.Scene {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

var sceneCmd = &cobra.Command{
	Use: "scene [-t | -tb | sceneNumber]",
	Example: "    Switch to the first scene: obs-switch scene 0\n" +
		"     Switch one scene forward: obs-switch scene -t\n" +
		"Go back to the previous scene: obs-switch scene -tb",
	DisableFlagsInUseLine: true,
	Short:                 "Switch program scene",
	RunE: func(cmd *cobra.Command, args []string) error {
		toggle, err := cmd.Flags().GetBool("toggle")
		if err != nil {
			return err
		}
		backwards, err := cmd.Flags().GetBool("backwards")
		if err != nil {
			return err
		}
		opts := []goobs.Option{}
		if len(password) > 0 {
			opts = []goobs.Option{goobs.WithPassword(password)}
		}
		obs, err := goobs.New(server, opts...)
		if err != nil {
			return err
		}
		defer obs.Disconnect()
		sl, err := obs.Scenes.GetSceneList()
		if err != nil {
			return err
		}
		if len(sl.Scenes) == 0 {
			return errors.New("no scenes to switch between")
		}
		if !toggle || (toggle && !backwards) {
			reverseScenes(sl.Scenes)
		}
		if toggle {
			scene, err := obs.Scenes.GetCurrentProgramScene()
			if err != nil {
				return err
			}
			nextSceneIndex := 0
			for i, s := range sl.Scenes {
				if s.SceneName == scene.CurrentProgramSceneName {
					nextSceneIndex = (i + 1) % len(sl.Scenes)
					break
				}
			}
			_, err = obs.Scenes.SetCurrentProgramScene(&scenes.SetCurrentProgramSceneParams{
				SceneName: sl.Scenes[nextSceneIndex].SceneName,
			})
			if err != nil {
				return err
			}
		} else {
			if len(args) == 0 || len(args) > 1 {
				return cmd.Help()
			}
			idx, err := strconv.Atoi(args[0])
			if err != nil {
				cmd.SilenceUsage = false
				return errors.New("argument must be an integer")
			}
			if idx >= len(sl.Scenes) {
				cmd.SilenceUsage = false
				return fmt.Errorf("index out-of-bounds, there are only %d scenes", len(sl.Scenes))
			}
			_, err = obs.Scenes.SetCurrentProgramScene(&scenes.SetCurrentProgramSceneParams{
				SceneName: sl.Scenes[idx].SceneName,
			})
			if err != nil {
				return err
			}
		}
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List scenes",
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := cmd.Flags().GetString("format")
		if err != nil {
			return err
		}
		if !(f == "json" || f == "text") {
			return errors.New("format must be json or text")
		}
		opts := []goobs.Option{}
		if len(password) > 0 {
			opts = []goobs.Option{goobs.WithPassword(password)}
		}
		obs, err := goobs.New(server, opts...)
		if err != nil {
			return err
		}
		defer obs.Disconnect()
		sl, err := obs.Scenes.GetSceneList()
		if err != nil {
			return err
		}
		if len(sl.Scenes) == 0 {
			return errors.New("no scenes to list")
		}
		reverseScenes(sl.Scenes)
		switch f {
		case "json":
			type Scene struct {
				SceneIndex int    `json:"index"`
				SceneName  string `json:"name"`
			}
			type ScenesSlice struct {
				Scenes []Scene `json:"scenes"`
			}
			obj := ScenesSlice{}
			for _, s := range sl.Scenes {
				obj.Scenes = append(obj.Scenes, Scene{SceneIndex: s.SceneIndex, SceneName: s.SceneName})
			}
			j, err := json.Marshal(obj)
			if err != nil {
				return err
			}
			fmt.Println(string(j))
		case "text":
			for _, s := range sl.Scenes {
				fmt.Println(s.SceneName)
			}
		}
		return nil
	},
}

var dumpCodeCmd = &cobra.Command{
	Use:     "dump-code",
	Aliases: []string{"dump", "code"},
	Short:   "Dump the code of this program",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(code)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&server, "server", "s", "localhost:4455", "`address` of obs-websocket to connect to")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "P", "", "websocket server password")

	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().StringP("format", "f", "json", "`format` to output, json or text")

	rootCmd.AddCommand(sceneCmd)
	sceneCmd.Flags().BoolP("toggle", "t", false, "toggle through all scenes")
	sceneCmd.Flags().BoolP("backwards", "b", false, "use with -t, toggle backwards instead of forward")

	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringP("format", "f", "json", "`format` to output, json or text")

	rootCmd.AddCommand(dumpCodeCmd)
}

func main() {
	Execute()
}
