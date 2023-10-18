/*
 *  Copyright (c) 2022 NetEase Inc.
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
	DEFAULT_TGTD_CONTAINER_NAME = "curvebs-target-daemon"
)

func checkStartIscsid(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if *success {
			return nil
		}
		return errno.ERR_START_SPDK_TARGET_FAILED.S(*out)
	}
}

func checkStartSPDKStatus(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if *success {
			return nil
		}
		return errno.ERR_START_SPDK_TARGET_FAILED.S(*out)
	}
}

func checkSetSPDKStatus(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if *success {
			return nil
		}
		return errno.ERR_SET_SPDK_TARGET_FAILED.S(*out)
	}
}

type (
	step2CheckTargetDaemonStatus struct {
		host   string
		status *string
	}
)

func (s *step2CheckTargetDaemonStatus) Execute(ctx *context.Context) error {
	if strings.HasPrefix(*s.status, "Up") {
		return task.ERR_SKIP_TASK
	} else if len(*s.status) == 0 {
		return nil
	}

	return errno.ERR_OLD_TARGET_DAEMON_IS_ABNORMAL.
		F("host=%s", s.host)
}

func NewStartTargetDaemonTask(curveadm *cli.CurveAdm, cc *configure.ClientConfig) (*task.Task, error) {
	options := curveadm.MemStorage().Get(comm.KEY_TARGET_OPTIONS).(TargetOption)
	hc, err := curveadm.GetHost(options.Host)
	if err != nil {
		return nil, err
	}

	// add step to task
	var success bool
	var status, containerId, out string
	containerName := DEFAULT_TGTD_CONTAINER_NAME
	hostname := containerName
	host2addr := fmt.Sprintf("%s:%s", hostname, hc.GetHostname())
	// for spdk
	startSPDKScript := scripts.START_SPDK
	startSPDKScriptPath := "/curvebs/tools/sbin/start_spdk.sh"
	setSPDKScript := scripts.SET_SPDK
	setSPDKScriptPath := "/curvebs/tools/sbin/set_spdk.sh"
	startSPDKCmd := fmt.Sprintf("bash %s", startSPDKScriptPath)
	setSPDKCmd := fmt.Sprintf("bash %s %s %d %d",
		setSPDKScriptPath, hc.GetHostname(), options.CacheSize, options.HugePageMem)
	toolsConf := fmt.Sprintf(FORMAT_TOOLS_CONF, cc.GetClusterMDSAddr())

	// new task
	subname := fmt.Sprintf("host=%s image=%s", options.Host, cc.GetContainerImage())
	t := task.NewTask("Start Target Daemon", subname, hc.GetSSHConfig())

	// add step to task
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.Status}}'",
		Filter:      fmt.Sprintf("name=%s", DEFAULT_TGTD_CONTAINER_NAME),
		Out:         &status,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2CheckTargetDaemonStatus{ // skip if target-daemon exist and healthy
		status: &status,
	})
	t.AddStep(&step.PullImage{
		Image:       cc.GetContainerImage(),
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.CreateContainer{
		Image:       cc.GetContainerImage(),
		AddHost:     []string{host2addr},
		Envs:        []string{"LD_PRELOAD=/usr/local/lib/libjemalloc.so"},
		Hostname:    hostname,
		Command:     "--role nebd",
		Name:        containerName,
		Pid:         "host",
		Privileged:  true,
		Volumes:     getVolumes(cc),
		Out:         &containerId,
		ExecOptions: curveadm.ExecOptions(),
	})
	for _, filename := range []string{"client.conf", "nebd-server.conf"} {
		t.AddStep(&step.SyncFile{
			ContainerSrcId:    &containerId,
			ContainerSrcPath:  "/curvebs/conf/" + filename,
			ContainerDestId:   &containerId,
			ContainerDestPath: "/curvebs/nebd/conf/" + filename,
			KVFieldSplit:      CLIENT_CONFIG_DELIMITER,
			Mutate:            newMutate(cc, CLIENT_CONFIG_DELIMITER),
			ExecOptions:       curveadm.ExecOptions(),
		})
	}
	t.AddStep(&step.SyncFile{ // sync nebd-client config
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  "/curvebs/conf/nebd-client.conf",
		ContainerDestId:   &containerId,
		ContainerDestPath: "/etc/nebd/nebd-client.conf",
		KVFieldSplit:      CLIENT_CONFIG_DELIMITER,
		Mutate:            newMutate(cc, CLIENT_CONFIG_DELIMITER),
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.SyncFile{ // sync client configuration for tgtd
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  "/curvebs/conf/client.conf",
		ContainerDestId:   &containerId,
		ContainerDestPath: "/etc/curve/client.conf",
		KVFieldSplit:      CLIENT_CONFIG_DELIMITER,
		Mutate:            newMutate(cc, CLIENT_CONFIG_DELIMITER),
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.InstallFile{ // install tools.conf
		Content:           &toolsConf,
		ContainerId:       &containerId,
		ContainerDestPath: "/etc/curve/tools.conf",
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.StartContainer{
		ContainerId: &containerId,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.ContainerExec{
		ContainerId: &containerId,
		Command:     "iscsid start",
		Success:     &success,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkStartIscsid(&success, &out),
	})

	if options.Spdk {
		t.AddStep(&step.InstallFile{ // install start_spdk.sh
			Content:           &startSPDKScript,
			ContainerId:       &containerId,
			ContainerDestPath: startSPDKScriptPath,
			ExecOptions:       curveadm.ExecOptions(),
		})
		t.AddStep(&step.InstallFile{ // install set_spdk.sh
			Content:           &setSPDKScript,
			ContainerId:       &containerId,
			ContainerDestPath: setSPDKScriptPath,
			ExecOptions:       curveadm.ExecOptions(),
		})
		t.AddStep(&step.ContainerExec{
			ContainerId: &containerId,
			Command:     startSPDKCmd,
			Daemon:      true,
			Success:     &success,
			Out:         &out,
			ExecOptions: curveadm.ExecOptions(),
		})
		t.AddStep(&step.Lambda{
			Lambda: checkStartSPDKStatus(&success, &out),
		})
		t.AddStep(&step.ContainerExec{
			ContainerId: &containerId,
			Command:     setSPDKCmd,
			Success:     &success,
			Out:         &out,
			ExecOptions: curveadm.ExecOptions(),
		})
		t.AddStep(&step.Lambda{
			Lambda: checkSetSPDKStatus(&success, &out),
		})
	} else {
		t.AddStep(&step.ContainerExec{
			Command:     "tgtd -f &",
			ContainerId: &containerId,
			ExecOptions: curveadm.ExecOptions(),
		})
	}
	return t, nil
}
