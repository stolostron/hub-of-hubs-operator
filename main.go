/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/component-base/cli"
	"k8s.io/component-base/logs"
	_ "k8s.io/component-base/logs/json/register"
	"open-cluster-management.io/addon-framework/pkg/version"

	hohagent "github.com/stolostron/hub-of-hubs-operator/pkg/agent"
	constants "github.com/stolostron/hub-of-hubs-operator/pkg/constants"
	hohmanager "github.com/stolostron/hub-of-hubs-operator/pkg/manager"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	logsOptions := logs.NewOptions()
	command := newCommand(logsOptions)
	logsOptions.AddFlags(command.Flags())
	code := cli.Run(command)
	os.Exit(code)
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func newCommand(logsOptions *logs.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addon",
		Short: "hub-of-hubs operator",
		Run: func(cmd *cobra.Command, args []string) {
			if err := logsOptions.ValidateAndApply(); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}

	if v := version.Get().String(); len(v) == 0 {
		cmd.Version = "<unknown>"
	} else {
		cmd.Version = v
	}

	cmd.AddCommand(hohmanager.NewControllerCommand())
	cmd.AddCommand(hohagent.NewAgentCommand(constants.HoHOperatorName))

	return cmd
}
