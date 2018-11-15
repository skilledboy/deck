// Copyright © 2018 Harry Bagdi <harrybagdi@gmail.com>

package cmd

import (
	"io/ioutil"
	"strconv"

	"gopkg.in/yaml.v2"

	"github.com/hbagdi/deck/diff"
	"github.com/hbagdi/deck/dump"
	"github.com/hbagdi/deck/state"
	"github.com/hbagdi/go-kong/kong"
	"github.com/spf13/cobra"
)

var kongStateFile string

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		target, err := state.NewKongState()
		if err != nil {
			return err
		}
		current, err := state.NewKongState()
		if err != nil {
			return err
		}
		client, err := kong.NewClient(nil, nil)
		if err != nil {
			return err
		}
		// client.SetDebugMode(true)
		services, err := dump.GetAllServices(client)
		if err != nil {
			return err
		}
		for _, service := range services {
			var s state.Service
			s.Service = *service
			err := current.AddService(s)
			if err != nil {
				return err
			}
		}
		routes, err := dump.GetAllRoutes(client)
		if err != nil {
			return err
		}
		for _, route := range routes {
			var r state.Route
			r.Route = *route
			err := current.AddRoute(r)
			if err != nil {
				return err
			}
		}
		servicesFromLocalState, err := readFile(kongStateFile)
		if err != nil {
			return err
		}
		// targetServices := make([]state.Service, 3)

		for i, s := range servicesFromLocalState {
			// TODO add override logic
			// TODO add support for file based defaults
			// TODO add counter abstraction to use everywhere in this file
			s.ID = kong.String("placeholder" + strconv.Itoa(i))
			// s.Name = kong.String("name" + strconv.Itoa(i))
			// s.Host = kong.String("host" + strconv.Itoa(i))
			// s.Port = kong.Int(80)
			// s.ConnectTimeout = kong.Int(60000)
			// s.Protocol = kong.String("http")
			// s.WriteTimeout = kong.Int(1000)
			// s.ReadTimeout = kong.Int(60000)
			// s.Retries = kong.Int(5)
			// s.Host = kong.String("host" + strconv.Itoa(i))

			// use a set to check for duplicate services (services with same name)
			err := target.AddService(s)
			if err != nil {
				return err
			}
		}
		// var route state.Route
		// route.Paths = kong.StringSlice("/foo")
		// route.Service = &kong.Service{
		// 	Name: servicesFromLocalState[0].Name,
		// }
		// target.AddRoute(route)
		s, _ := diff.NewSyncer(current, target)
		gDelete, gCreateUpdate, err := s.Diff()
		if err != nil {
			return err
		}
		err = s.Solve(gDelete, client)
		if err != nil {
			return err
		}
		err = s.Solve(gCreateUpdate, client)
		if err != nil {
			return err
		}
		return nil
	},
}

func readFile(kongStateFile string) ([]state.Service, error) {
	type State struct {
		Services []state.Service `yaml:"services"`
	}
	var s State
	b, err := ioutil.ReadFile(kongStateFile)
	err = yaml.Unmarshal(b, &s)
	if err != nil {
		return nil, err
	}
	return s.Services, nil
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	syncCmd.Flags().StringVarP(&kongStateFile, "state", "s", "", "Kong configuration directory or file")
}