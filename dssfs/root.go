package dssfs

import (
	"context"
	"log"
	"os"
	"sync"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/kwalter94/dss-fuse/dssapi"
)

type Root struct {
	api             *dssapi.Client
	dirs            map[string]*Dir
	dirsAccessMutex sync.RWMutex
}

func NewRootDir(api *dssapi.Client) (*Root, error) {
	return &Root{api: api, dirs: make(map[string]*Dir)}, nil
}

func (root *Root) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = 1
	attr.Mode = os.ModeDir | 0o555
	attr.Uid = uint32(os.Getuid())
	attr.Gid = uint32(os.Getgid())

	return nil
}

func (root *Root) Lookup(ctx context.Context, name string) (fs.Node, error) {
	root.dirsAccessMutex.RLock()
	defer root.dirsAccessMutex.RUnlock()

	log.Printf("Looking up directory: %s", name)

	if dir := root.getDir(name); dir != nil {
		return dir, nil
	}

	return nil, syscall.ENOENT
}

func (root *Root) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Printf("Opening root directory")
	return root, nil
}

func (root *Root) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	root.dirsAccessMutex.Lock()
	defer root.dirsAccessMutex.Unlock()

	projects, err := root.api.GetProjects()
	if err != nil {
		log.Printf("Error: Failed to load projects: %v", err)
		return nil, syscall.EIO
	}

	dirents := make([]fuse.Dirent, 0, len(projects))

	for _, project := range projects {
		dir := root.getDir(project.Name)

		if dir == nil {
			dir = NewDir(project, root.api)
		} else {
			dir.ReloadProject(project)
		}

		root.dirs[project.Name] = dir
		dirents = append(dirents, fuse.Dirent{Inode: dir.Inode(), Type: fuse.DT_Dir, Name: project.Name})
	}

	log.Printf("Loaded %d projects... %v", len(dirents), dirents[:5])

	return dirents, nil
}

func (root *Root) getDir(name string) *Dir {
	if dir, exists := root.dirs[name]; exists {
		return dir
	}

	return nil
}
