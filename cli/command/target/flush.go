/*
 *  Copyright (c) 2021 NetEase Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

/*
 * Project: CurveAdm
 * Created Date: 2023-11-20
 * Author: Caoxianfei
 */
package target

import (
	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	FLUSH_PLAYBOOK_STEPS = []int{
		playbook.FLUSH_TARGET,
	}
)

type flushOption struct {
	host   string
	target string
}

func NewFlushCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options flushOption

	cmd := &cobra.Command{
		Use:     "flush [OPTIONS]",
		Aliases: []string{"flush"},
		Short:   "flush target",
		Args:    cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.target = args[0]
			return runFlush(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.host, "host", "localhost", "Specify target host")

	return cmd
}

func genFlushPlaybook(curveadm *cli.CurveAdm,
	ccs []*configure.ClientConfig,
	options flushOption) (*playbook.Playbook, error) {
	steps := FLUSH_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: ccs,
			Options: map[string]interface{}{
				comm.KEY_TARGET_OPTIONS: bs.TargetOption{
					Host:   options.host,
					Target: options.target,
				},
			},
		})
	}
	return pb, nil
}

func runFlush(curveadm *cli.CurveAdm, options flushOption) error {
	pb, err := genFlushPlaybook(curveadm, []*configure.ClientConfig{configure.NewEmptyClientConfig()}, options)
	if err != nil {
		return err
	}

	// 3) run playground
	err = pb.Run()
	if err != nil {
		return err
	}

	// 4) print success prompt
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.GreenString("Flush target (target=%s) on %s success ^_^"),
		options.target, options.host)

	return nil
}

// for http service
func FlushSpdkTgt(curveadm *cli.CurveAdm, target, host string) error {
	options := flushOption{
		host:   host,
		target: target,
	}

	// generate list playbook
	pb, err := genFlushPlaybook(curveadm,
		[]*configure.ClientConfig{configure.NewEmptyClientConfig()},
		options)

	if err != nil {
		return err
	}

	// run playground
	err = pb.Run()
	if err != nil {
		return err
	}

	return nil
}
