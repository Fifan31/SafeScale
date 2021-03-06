/*
 * Copyright 2018, CS Systemes d'Information, http://www.c-s.fr
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package Complexity

import (
	"fmt"
	"strings"
)

//go:generate stringer -type=Enum

//Enum represents the complexity of a cluster
type Enum int

const (

	//Dev is the simplest mode of cluster
	Dev Enum = 1
	//Normal allows the cluster to be resistant to 1 master failure
	Normal Enum = 3
	//Volume allows the cluster to be resistant to 2 master failures and is sized for high volume of agents
	Volume Enum = 5
)

//FromString returns a Complexity.Enum corresponding to String
func FromString(complexity string) (Enum, error) {
	lowered := strings.ToLower(complexity)
	if lowered == "dev" {
		return Dev, nil
	}
	if lowered == "normal" {
		return Normal, nil
	}
	if lowered == "volume" {
		return Volume, nil
	}
	return 0, fmt.Errorf("incorrect complexity '%s'", complexity)
}
