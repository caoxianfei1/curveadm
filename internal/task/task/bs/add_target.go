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

package bs

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	TARGET_PREFIX      = "iqn.2016-06.io.spdk:"
	TGTD_PORT          = "3260"
	MALLOC             = "Malloc"
	GLOBAL_CACHE_NAME  = "Malloc_host_"
	GLOBAL             = "[global]"
	LOCAL_CACHE_PREFIX = "Malloc"
	LOCAL              = "[local]"
)

type TargetOption struct {
	Host            string
	User            string
	Volume          string
	Create          bool
	Size            int
	Tid             string
	Blocksize       uint64
	CacheSize       uint64
	HugePageMem     uint64
	CreateCacheDisk bool
	Target          string // delete spdk target
	WritePolicy     string
	DisableCache    bool
	Spdk            bool
}

type step2InsertTarget struct {
	curveadm *cli.CurveAdm
	options  TargetOption
}

func (s *step2InsertTarget) Execute(ctx *context.Context) error {
	curveadm := s.curveadm
	options := s.options
	portalInfo := fmt.Sprint(options.Host, ":", TGTD_PORT)
	volumeConv := strings.ReplaceAll(options.Volume, "/", "_")                              // volume: _test1_test2, cache: Malloc_test1_test2
	taregtNameSuffix := strings.TrimLeft(strings.ReplaceAll(options.Volume, "/", "-"), "-") // target: test1-test2
	target := fmt.Sprintf("%s%s", TARGET_PREFIX, taregtNameSuffix)
	volumeId := curveadm.GetVolumeId(options.Host, options.User, options.Volume)

	var cache string
	if !options.DisableCache {
		cache = fmt.Sprint(GLOBAL_CACHE_NAME, GLOBAL)
		if options.CreateCacheDisk {
			cache = fmt.Sprint(MALLOC, volumeConv, LOCAL)
		}
	}

	err := curveadm.Storage().InsertTarget(volumeId, target, options.Volume, portalInfo, cache)
	if err != nil {
		return errno.ERR_INSERT_TARGET_FAILED.E(err)
	}

	return nil
}

func checkAddSPDKTargetStatus(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if !*success {
			return errno.ERR_ADD_SPDK_TARGET_FAILED.S(*out)
		}
		if *out == "EXIST" {
			return task.ERR_SKIP_TASK
		}

		return nil
	}
}

func NewAddTargetTask(curveadm *cli.CurveAdm, cc *configure.ClientConfig) (*task.Task, error) {
	options := curveadm.MemStorage().Get(comm.KEY_TARGET_OPTIONS).(TargetOption)
	user, volume := options.User, options.Volume
	hc, err := curveadm.GetHost(options.Host)
	if err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("host=%s volume=%s", options.Host, volume)
	t := task.NewTask("Add Target", subname, hc.GetSSHConfig())

	// add step
	var output string
	var success bool
	containerId := DEFAULT_TGTD_CONTAINER_NAME
	targetScriptPath := "/curvebs/tools/sbin/add_spdk_target.sh"
	targetScript := scripts.TARGET
	cmd := fmt.Sprintf("bash %s %s %s %v %d %d %d %v %s %v %s %v",
		targetScriptPath,
		user,
		volume,
		options.Create,
		options.Size,
		options.Blocksize,
		options.CacheSize,
		options.Spdk,
		hc.GetHostname(),
		options.CreateCacheDisk,
		options.WritePolicy,
		options.DisableCache,
	)

	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.ID}} {{.Status}}'",
		Filter:      fmt.Sprintf("name=%s", DEFAULT_TGTD_CONTAINER_NAME),
		Out:         &output,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2CheckTgtdStatus{
		output: &output,
	})

	t.AddStep(&step.InstallFile{ // install target.sh
		Content:           &targetScript,
		ContainerId:       &containerId,
		ContainerDestPath: targetScriptPath,
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.ContainerExec{
		ContainerId: &containerId,
		Command:     cmd,
		Success:     &success,
		Out:         &output,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkAddSPDKTargetStatus(&success, &output),
	})
	t.AddStep(&step2InsertTarget{
		curveadm: curveadm,
		options:  options,
	})

	return t, nil
}
