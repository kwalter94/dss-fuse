package main

import (
	"log"
	"os"
	"time"

	"github.com/kwalter94/dss-fuse/dssapi"
	"github.com/kwalter94/dss-fuse/dssfs"
)

var DssConnectTimeout time.Duration

func init() {
	DssConnectTimeout = time.Second * 30
}

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Expected a directory name to mount filesystem on")
	}

	dssConfig, err := dssapi.LoadUserConfig()
	if err != nil {
		panic(err)
	}

	instance := dssConfig.GetDefaultInstance()
	if instance == nil {
		panic("No default instance specified in user config!")
	}

	dssClient, err := dssapi.NewDssClient(instance.Url, instance.ApiKey, DssConnectTimeout)
	if err != nil {
		panic(err)
	}

	log.Printf("Connected to DSS: %s", instance.Url)
	fs := dssfs.NewFS(dssClient)
	if err := fs.MountAndServe(os.Args[1]); err != nil {
		panic(err)
	}
}
