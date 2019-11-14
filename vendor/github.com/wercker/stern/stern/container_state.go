//   Copyright 2016 Wercker Holding BV
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package stern

import (
	"errors"

	v1 "k8s.io/api/core/v1"
)

type ContainerState []string

const (
	RUNNING    = "running"
	WAITING    = "waiting"
	TERMINATED = "terminated"
)

func NewContainerState(stateConfig []string) (ContainerState, error) {
	var containerState []string
	for _, p := range stateConfig {
		if p == RUNNING || p == WAITING || p == TERMINATED {
			containerState = append(containerState, p)
		}
	}
	if len(containerState) == 0 {
		return []string{}, errors.New("containerState should include 'running', 'waiting', or 'terminated'")
	}
	return containerState, nil
}

func (stateConfig ContainerState) Match(containerState v1.ContainerState) bool {
	if containerState.Running != nil && stateConfig.has(RUNNING) {
		return true
	}
	if containerState.Waiting != nil && stateConfig.has(WAITING) {
		return true
	}
	if containerState.Terminated != nil && stateConfig.has(TERMINATED) {
		return true
	}
	return false
}

func (stateConfig ContainerState) has(state string) bool {
	for _, s := range stateConfig {
		if s == state {
			return true
		}
	}
	return false
}
