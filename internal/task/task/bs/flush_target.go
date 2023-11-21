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
 * Created Date: 2022-11-20
 * Author: Caoxianfei
 */

package bs

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/common"
	client "github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

func checkFlushTarget(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if !*success {
			return errno.ERR_FLUSH_TARGET_FAILED.S(*out)
		}

		return nil
	}
}

func NewFlushTargetTask(curveadm *cli.CurveAdm, cc *client.ClientConfig) (*task.Task, error) {
	options := curveadm.MemStorage().Get(common.KEY_TARGET_OPTIONS).(TargetOption)
	hc, err := curveadm.GetHost(options.Host)
	if err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("hostname=%s target=%s", hc.GetHostname(), options.Target)
	t := task.NewTask("Flush Target", subname, hc.GetSSHConfig())

	targetScript := scripts.FLUSH_SPDK
	targetScriptPath := "/curvebs/tools/sbin/flush_target.sh"
	cmd := fmt.Sprintf("bash %s %s %s",
		targetScriptPath,
		options.Target,
		hc.GetHostname(),
	)
	// add step
	var output string
	var success bool
	containerId := DEFAULT_TGTD_CONTAINER_NAME
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
	t.AddStep(&step.InstallFile{ // install flush_spdk.sh
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
		Lambda: checkFlushTarget(&success, &output),
	})

	return t, nil
}
