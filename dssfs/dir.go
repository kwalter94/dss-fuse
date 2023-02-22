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

type Dir struct {
	inode             uint64
	project           dssapi.Project
	files             map[string]*File
	filesAccessMutex  sync.RWMutex
	api               *dssapi.Client
	cacheExpiresAfter time.Time
}

func NewDir(project dssapi.Project, api *dssapi.Client) *Dir {
	return &Dir{
		inode:   nextInode(),
		api:     api,
		project: project,
		files:   make(map[string]*File),
	}
}

func (dir *Dir) Name() string {
	return dir.project.Name
}

func (dir *Dir) Inode() uint64 {
	return dir.inode
}

func (dir *Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = dir.inode
	attr.Mode = os.ModeDir | 0o555
	attr.Mtime = dir.project.DateModified()
	attr.Ctime = dir.project.DateCreated()
	attr.Uid = uint32(os.Getuid())
	attr.Gid = uint32(os.Getgid())

	return nil
}

func (dir *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Printf("Looking up file: %s", name)
	if err := dir.loadFiles(); err != nil {
		return nil, err
	}

	if file, exists := dir.files[name]; exists {
		log.Printf("File found: %s", name)
		return file, nil
	}

	return nil, syscall.ENOENT
}

func (dir *Dir) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Printf("Opening directory: %s", dir.project.Name)
	return dir, nil
}

func (dir *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	if err := dir.loadFiles(); err != nil {
		return nil, err
	}

	dirents := make([]fuse.Dirent, 0, len(dir.files))

	log.Printf("Attaching project `%s` recipes", dir.project.Name)

	for _, file := range dir.files {
		dirent := fuse.Dirent{
			Inode: file.Inode(),
			Type:  fuse.DT_File,
			Name:  file.Name(),
		}

		dirents = append(dirents, dirent)
	}

	return dirents, nil
}

func (dir *Dir) loadFiles() error {
	log.Printf("Loading project files: %s...", dir.project.Name)
	dir.filesAccessMutex.Lock()
	defer dir.filesAccessMutex.Unlock()

	if !time.Now().After(dir.cacheExpiresAfter) {
		log.Print("Using cached project files")
		return nil
	}

	dir.files = make(map[string]*File)

	recipes, err := dir.api.GetRecipes(dir.project)
	if err != nil {
		log.Printf("Error: Failed to fetch recipe: %v", err)
		return syscall.EIO
	}

	for _, recipe := range recipes {
		if !recipe.IsEditable() {
			log.Printf("Skipping non-editable recipe: %s", recipe.Type)
			continue
		}

		file := NewFile(recipe, dir.api)
		dir.files[recipe.Name] = file
	}

	dir.cacheExpiresAfter = time.Now().Add(DSS_DATA_CACHE_PERIOD)

	return nil
}
