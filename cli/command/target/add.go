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
 * Created Date: 2022-02-08
 * Author: Jingli Chen (Wine93)
 */

package target

import (
	"strings"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/cli/command/client"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	ADD_PLAYBOOK_STEPS = []int{
		//playbook.CREATE_VOLUME,
		playbook.ADD_TARGET,
	}
)

type addOptions struct {
	image       string
	host        string
	size        string
	create      bool
	blocksize   string
	cachesize   string
	spdk        bool
	createcache bool
	writepolicy string
	usecache    bool
}

func checkAddOptions(curveadm *cli.CurveAdm, options addOptions) error {
	if _, _, err := client.ParseImage(options.image); err != nil {
		return err
	} else if _, err = client.ParseSize(options.size); err != nil {
		return err
	} else if _, err = client.ParseBlockSize(options.blocksize); err != nil {
		return err
	} else if _, err = client.ParseCacheSize(options.cachesize); err != nil {
		return err
	} else if _, err = client.ParseWritePolicy(options.writepolicy); err != nil {
		return err
	}
	return nil
}

func NewAddCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options addOptions

	cmd := &cobra.Command{
		Use:   "add USER:VOLUME [OPTIONS]",
		Short: "Add a target of CurveBS",
		Args:  cliutil.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			options.image = args[0]
			return checkAddOptions(curveadm, options)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			options.image = args[0]
			return runAdd(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.host, "host", "localhost", "Specify target host")
	flags.BoolVar(&options.create, "create", false, "Create volume iff not exist")
	flags.StringVar(&options.size, "size", "10GiB", "Specify volume size")
	flags.StringVar(&options.blocksize, "blocksize", "4096B", "Specify volume blocksize")
	flags.StringVar(&options.cachesize, "cachesize", "64MB", "Specify cachesize MB")
	flags.BoolVar(&options.spdk, "spdk", false, "create iscsi spdk target")
	flags.BoolVar(&options.createcache, "createcache", false, "create cache disk for current target")
	flags.BoolVar(&options.usecache, "usecache", true, "use cache disk or not")
	flags.StringVar(&options.writepolicy, "writepolicy", "", "Specify write policy of cache, default is ''")
	return cmd
}

func genAddPlaybook(curveadm *cli.CurveAdm,
	ccs []*configure.ClientConfig,
	options addOptions) (*playbook.Playbook, error) {
	user, name, _ := client.ParseImage(options.image)
	size, _ := client.ParseSize(options.size)
	blocksize, _ := client.ParseBlockSize(options.blocksize)
	cachesize, _ := client.ParseCacheSize(options.cachesize)
	writepolicy, _ := client.ParseWritePolicy(options.writepolicy)
	steps := ADD_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: ccs,
			Options: map[string]interface{}{
				comm.KEY_TARGET_OPTIONS: bs.TargetOption{
					Host:            options.host,
					User:            user,
					Volume:          name,
					Size:            size,
					Create:          options.create,
					Blocksize:       blocksize,
					CacheSize:       cachesize,
					CreateCacheDisk: options.createcache,
					WritePolicy:     writepolicy,
					Usecache:        options.usecache,
					Spdk:            options.spdk,
				},
			},
		})
	}
	return pb, nil
}

func runAdd(curveadm *cli.CurveAdm, options addOptions) error {
	// 2) generate map playbook
	pb, err := genAddPlaybook(curveadm, []*configure.ClientConfig{configure.NewEmptyClientConfig()}, options)
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
	curveadm.WriteOutln(color.GreenString("Add target (%s) to %s success ^_^"),
		options.image, options.host)
	return nil
}

// for http service spdk
func AddSpdkTgt(curveadm *cli.CurveAdm,
	image, host, size, cacheSize, blockSize string,
	createImage, createCache bool,
	writePolicy string,
	useCache bool) error {
	// new ClientConfig object
	defaultSize, defaultBlockSize, defaultCacheSize := "10GiB", "4096B", "64MB"
	if strings.TrimSpace(size) != "" {
		defaultSize = size
	}
	if strings.TrimSpace(blockSize) != "" {
		defaultBlockSize = blockSize
	}
	if strings.TrimSpace(cacheSize) != "" {
		defaultCacheSize = cacheSize
	}
	defaultCreate, defaultCreateCache := false, false
	if createImage {
		defaultCreate = true
	}

	if createCache {
		defaultCreateCache = true
	}

	defaultWritePolicy := ""
	if strings.TrimSpace(writePolicy) != "" {
		defaultWritePolicy = writePolicy
	}

	defaultUseCache := true
	if !useCache {
		defaultUseCache = false
	}

	options := addOptions{
		image:       image,
		host:        host,
		size:        defaultSize,
		create:      defaultCreate,
		cachesize:   defaultCacheSize,
		createcache: defaultCreateCache,
		blocksize:   defaultBlockSize,
		writepolicy: defaultWritePolicy,
		usecache:    defaultUseCache,
		spdk:        true,
	}
	err := checkAddOptions(curveadm, options)
	if err != nil {
		return err
	}

	// 2) generate map playbook
	pb, err := genAddPlaybook(curveadm, []*configure.ClientConfig{configure.NewEmptyClientConfig()}, options)
	if err != nil {
		return err
	}

	// 3) run playground
	err = pb.Run()
	if err != nil {
		return err
	}

	return nil
}
