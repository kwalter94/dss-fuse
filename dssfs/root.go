package dssfs

import (
	"context"
	"log"
	"os"
	"sync"
	"syscall"
	"time"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"github.com/kwalter94/dss-fuse/dssapi"
)

var DSS_DATA_CACHE_PERIOD time.Duration = time.Minute * 10

type Root struct {
	api               *dssapi.Client
	dirs              map[string]*Dir
	dirsAccessMutex   sync.RWMutex
	cacheExpiresAfter time.Time
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
	log.Printf("Looking up directory: %s", name)
	if err := root.loadDirs(); err != nil {
		return nil, err
	}

	if dir, exists := root.dirs[name]; exists {
		log.Printf("Found directory: %s", name)
		return dir, nil
	}

	log.Printf("Error: Directory not found: %s", name)

	return nil, syscall.ENOENT
}

func (root *Root) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Printf("Opening root directory")
	return root, nil
}

func (root *Root) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	if err := root.loadDirs(); err != nil {
		return nil, err
	}

	dirents := make([]fuse.Dirent, 0, len(root.dirs))

	for _, dir := range root.dirs {
		dirents = append(dirents, fuse.Dirent{
			Inode: dir.Inode(),
			Type:  fuse.DT_Dir,
			Name:  dir.Name(),
		})
	}

	return dirents, nil
}

func (root *Root) loadDirs() error {
	log.Print("Loading root directory...")
	root.dirsAccessMutex.Lock()
	defer root.dirsAccessMutex.Unlock()

	if !time.Now().After(root.cacheExpiresAfter) {
		log.Print("Using cached projects")
		return nil
	}

	projects, err := root.api.GetProjects()
	if err != nil {
		log.Printf("Error: Failed to load projects: %v", err)
		return syscall.EIO
	}

	for _, project := range projects {
		dir := NewDir(project, root.api)
		root.dirs[dir.Name()] = dir
	}

	root.cacheExpiresAfter = time.Now().Add(DSS_DATA_CACHE_PERIOD)
	log.Printf("Loaded %d projects", len(projects))

	return nil
}
