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

package commands

import (
	"context"
	"fmt"
	"log"

	pb "github.com/CS-SI/SafeScale/broker"
	services "github.com/CS-SI/SafeScale/broker/daemon/services"

	google_protobuf "github.com/golang/protobuf/ptypes/empty"
)

// broker ssh connect vm2
// broker ssh run vm2 -c "uname -a"
// broker ssh copy /file/test.txt vm1://tmp
// broker ssh copy vm1:/file/test.txt /tmp

//SSHServiceServer SSH service server grpc
type SSHServiceServer struct{}

//Run executes an ssh command an a VM
func (s *SSHServiceServer) Run(ctx context.Context, in *pb.SshCommand) (*pb.SshResponse, error) {
	log.Printf("Ssh run called")
	if GetCurrentTenant() == nil {
		return nil, fmt.Errorf("No tenant set")
	}

	service := services.NewSSHService(currentTenant.client)
	out, err := service.Run(in.GetVM().GetName(), in.GetCommand())
	if err != nil {
		return nil, err
	}

	log.Println("End ssh run")
	return &pb.SshResponse{
		Status: 0,
		Output: out,
		Err:    "",
	}, nil
}

//Copy copy file from/to a VM
func (s *SSHServiceServer) Copy(ctx context.Context, in *pb.SshCopyCommand) (*google_protobuf.Empty, error) {
	log.Printf("Ssh copy called")
	if GetCurrentTenant() == nil {
		return nil, fmt.Errorf("No tenant set")
	}

	service := services.NewSSHService(currentTenant.client)
	err := service.Copy(in.GetSource(), in.GetDestination())
	if err != nil {
		return nil, err
	}

	log.Println("End ssh copy")
	return &google_protobuf.Empty{}, nil
}
