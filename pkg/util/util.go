/*
==================================================================================
  Copyright (c) 2019 AT&T Intellectual Property.
  Copyright (c) 2019 Nokia

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
==================================================================================
*/

package util

import (
	"bytes"
	"errors"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"strings"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/appmgr/pkg/appmgr"
)

var execCommand = exec.Command

func Exec(args string) (out []byte, err error) {
	cmd := execCommand("/bin/sh", "-c", args)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	appmgr.Logger.Info("Running command: %s ", cmd.Args)
	for i := 0; i < viper.GetInt("helm.retry"); i++ {
		if err = cmd.Run(); err != nil {
			appmgr.Logger.Error("Command failed: %v - %s, retrying", err.Error(), stderr.String())
			time.Sleep(time.Duration(2) * time.Second)
			continue
		}
		break
	}

	if err == nil && !strings.HasSuffix(os.Args[0], ".test") {
		appmgr.Logger.Info("command success: %s", stdout.String())
		return stdout.Bytes(), nil
	}

	return stdout.Bytes(), errors.New(stderr.String())
}

var HelmExec = func(args string) (out []byte, err error) {
	return Exec(strings.Join([]string{"helm", args}, " "))
}

var KubectlExec = func(args string) (out []byte, err error) {
	return Exec(strings.Join([]string{"kubectl", args}, " "))
}
