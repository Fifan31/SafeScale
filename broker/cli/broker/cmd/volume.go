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

package cmd

import (
	"encoding/json"
	"fmt"

	pb "github.com/CS-SI/SafeScale/broker"
	"github.com/CS-SI/SafeScale/broker/client"
	"github.com/CS-SI/SafeScale/broker/utils"
	"github.com/CS-SI/SafeScale/providers/api"
	"github.com/urfave/cli"
)

//VolumeCmd volume command
var VolumeCmd = cli.Command{
	Name:  "volume",
	Usage: "volume COMMAND",
	Subcommands: []cli.Command{
		volumeList,
		volumeInspect,
		volumeDelete,
		volumeCreate,
		volumeAttach,
		volumeDetach,
	},
}

var volumeList = cli.Command{
	Name:  "list",
	Usage: "List available volumes",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "all",
			Usage: "List all Volumes on tenant (not only those created by SafeScale)",
		}},
	Action: func(c *cli.Context) error {
		resp, err := client.New().Volume.List(c.Bool("all"), 0)
		if err != nil {
			return fmt.Errorf("Could not get volume list: %v", err)
		}

		var volumes []*volumeDisplayable
		for _, vol := range resp.GetVolumes() {
			volumes = append(volumes, toDisplaybleVolume(vol))
		}
		out, _ := json.Marshal(volumes)
		fmt.Println(string(out))
		return nil
	},
}

var volumeInspect = cli.Command{
	Name:      "inspect",
	Usage:     "Inspect volume",
	ArgsUsage: "<Volume_name|Volume_ID>",
	Action: func(c *cli.Context) error {
		if c.NArg() != 1 {
			fmt.Println("Missing mandatory argument <Volume_name|Volume_ID>")
			cli.ShowSubcommandHelp(c)
			return fmt.Errorf("Volume name or ID required")
		}
		volumeInfo, err := client.New().Volume.Inspect(c.Args().First(), 0)
		if err != nil {
			return fmt.Errorf("Could not get volume '%s': %v", c.Args().First(), err)
		}

		out, _ := json.Marshal(toDisplaybleVolumeInfo(volumeInfo))
		fmt.Println(string(out))

		return nil
	},
}

var volumeDelete = cli.Command{
	Name:      "delete",
	Usage:     "Delete volume",
	ArgsUsage: "<Volume_name|Volume_ID>",
	Action: func(c *cli.Context) error {
		if c.NArg() != 1 {
			fmt.Println("Missing mandatory argument <Volume_name|Volume_ID>")
			cli.ShowSubcommandHelp(c)
			return fmt.Errorf("Volume name or ID required")
		}
		err := client.New().Volume.Delete(c.Args().First(), 0)
		if err != nil {
			return fmt.Errorf("Could not delete volume '%s': %v", c.Args().First(), err)
		}
		fmt.Printf("Volume '%s' deleted\n", c.Args().First())

		return nil
	},
}

var volumeCreate = cli.Command{
	Name:      "create",
	Usage:     "Create a volume",
	ArgsUsage: "<Volume_name>",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "size",
			Value: 10,
			Usage: "Size of the volume (in Go)",
		},
		cli.StringFlag{
			Name:  "speed",
			Value: "HDD",
			Usage: fmt.Sprintf("Allowed values: %s", getAllowedSpeeds()),
		},
	},
	Action: func(c *cli.Context) error {
		if c.NArg() != 1 {
			fmt.Println("Missing mandatory argument <Volume_name>")
			cli.ShowSubcommandHelp(c)
			return fmt.Errorf("Volume name required")
		}
		speed := c.String("speed")

		volSpeed, ok := pb.VolumeSpeed_value[speed]
		if !ok {
			return fmt.Errorf("Invalid speed '%s'", speed)
		}
		def := pb.VolumeDefinition{
			Name:  c.Args().First(),
			Size:  int32(c.Int("size")),
			Speed: pb.VolumeSpeed(volSpeed),
		}

		volume, err := client.New().Volume.Create(def, 0)
		if err != nil {
			return fmt.Errorf("Could not create volume '%s': %v", c.Args().First(), err)
		}
		out, _ := json.Marshal(toDisplaybleVolume(volume))
		fmt.Println(string(out))

		return nil
	},
}

var volumeAttach = cli.Command{
	Name:      "attach",
	Usage:     "Attach a volume to an host",
	ArgsUsage: "<Volume_name|Volume_ID>, <Host_name|Host_ID>",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "path",
			Value: api.DefaultVolumeMountPoint,
			Usage: "Mount point of the volume",
		},
		cli.StringFlag{
			Name:  "format",
			Value: "ext4",
			Usage: "Filesystem format",
		},
	},
	Action: func(c *cli.Context) error {
		if c.NArg() != 2 {
			fmt.Println("Missing mandatory argument <Volume_name> and/or <Host_name>")
			cli.ShowSubcommandHelp(c)
			return fmt.Errorf("Volume and Host name required")
		}
		def := pb.VolumeAttachment{
			Format:    c.String("format"),
			MountPath: c.String("path"),
			Host:      &pb.Reference{Name: c.Args().Get(1)},
			Volume:    &pb.Reference{Name: c.Args().Get(0)},
		}
		err := client.New().Volume.Attach(def, 0)
		if err != nil {
			return fmt.Errorf("could not attach volume '%s' to host '%s': %v", c.Args().Get(0), c.Args().Get(1), err)
		}
		fmt.Printf("Volume '%s' attached to host '%s'\n", c.Args().Get(0), c.Args().Get(1))

		return nil
	},
}

var volumeDetach = cli.Command{
	Name:      "detach",
	Usage:     "Detach a volume from an host",
	ArgsUsage: "<Volume_name|Volume_ID> <Host_name|Host_ID>",
	Action: func(c *cli.Context) error {
		if c.NArg() != 2 {
			fmt.Println("Missing mandatory argument <Volume_name> and/or <Host_name>")
			cli.ShowSubcommandHelp(c)
			return fmt.Errorf("volume and host names required")
		}
		err := client.New().Volume.Detach(c.Args().Get(0), c.Args().Get(1), 0)
		if err != nil {
			return fmt.Errorf("could not detach volume '%s' from host '%s': %v", c.Args().Get(0), c.Args().Get(1), err)
		}
		fmt.Printf("Volume '%s' detached from host '%s'\n", c.Args().Get(0), c.Args().Get(1))

		return nil
	},
}

type volumeInfoDisplayable struct {
	ID        string
	Name      string
	Speed     string
	Size      int32
	Host      string
	MountPath string
	Format    string
	Device    string
}

type volumeDisplayable struct {
	ID    string
	Name  string
	Speed string
	Size  int32
}

func toDisplaybleVolumeInfo(volumeInfo *pb.VolumeInfo) *volumeInfoDisplayable {
	return &volumeInfoDisplayable{
		volumeInfo.GetID(),
		volumeInfo.GetName(),
		pb.VolumeSpeed_name[int32(volumeInfo.GetSpeed())],
		volumeInfo.GetSize(),
		utils.GetReference(volumeInfo.GetHost()),
		volumeInfo.GetMountPath(),
		volumeInfo.GetFormat(),
		volumeInfo.GetDevice(),
	}
}

func toDisplaybleVolume(volumeInfo *pb.Volume) *volumeDisplayable {
	return &volumeDisplayable{
		volumeInfo.GetID(),
		volumeInfo.GetName(),
		pb.VolumeSpeed_name[int32(volumeInfo.GetSpeed())],
		volumeInfo.GetSize(),
	}
}

func getAllowedSpeeds() string {
	speeds := ""
	i := 0
	for k := range pb.VolumeSpeed_value {
		if i > 0 {
			speeds += ", "
		}
		speeds += k
		i++

	}
	return speeds
}
