package dssfs

import (
	"log"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	dssApi "github.com/kwalter94/dss-fuse/dssapi"
)

type FS struct {
	api *dssApi.Client
}

func NewFS(api *dssApi.Client) *FS {
	return &FS{api}
}

func (dss *FS) Root() (fs.Node, error) {
	root, err := NewRootDir(dss.api)
	if err != nil {
		return nil, err
	}

	return root, nil
}

func (dssfs *FS) MountAndServe(path string) error {
	log.Printf("Mounting dssfs at %s", path)

	conn, err := fuse.Mount(path, fuse.FSName("dss"), fuse.Subtype("dssfs"))
	if err != nil {
		return err
	}
	defer conn.Close()

	fs.Serve(conn, dssfs)

	return nil
}
