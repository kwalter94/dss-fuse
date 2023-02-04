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

type Dir struct {
	inode            uint64
	project          dssapi.Project
	files            map[string]*File
	filesAccessMutex sync.RWMutex
	api              *dssapi.Client
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
	return dir.project.ProjectKey
}

func (dir *Dir) Inode() uint64 {
	return dir.inode
}

func (dir *Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = dir.inode
	attr.Mode = os.ModeDir | 0o555
	attr.Mtime = dir.project.DateModified()
	attr.Ctime = dir.project.DateCreated()

	return nil
}

func (dir *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	dir.filesAccessMutex.RLock()
	defer dir.filesAccessMutex.RUnlock()

	if file := dir.getFile(name); file != nil {
		return file, nil
	}

	return nil, syscall.ENOENT
}

func (dir *Dir) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Printf("Opening directory: %s", dir.project.Name)
	return dir, nil
}

func (dir *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	dir.filesAccessMutex.Lock()
	defer dir.filesAccessMutex.Unlock()

	recipes, err := dir.api.GetRecipes(dir.project)
	if err != nil {
		log.Printf("Error: Failed to fetch recipe: %v", err)
		return nil, syscall.EIO
	}

	files := make(map[string]*File)
	dirents := make([]fuse.Dirent, 0, len(recipes))

	log.Printf("Attaching project `%s` recipes", dir.project.Name)

	for _, recipe := range recipes {
		if !recipe.IsEditable() {
			log.Printf("Skipping non-editable recipe: %s", recipe.Type)
			continue
		}

		file := dir.getFile(recipe.Name)

		if file == nil {
			file = NewFile(recipe, dir.api)
		} else {
			file.ReloadRecipe(recipe)
		}

		dirent := fuse.Dirent{
			Inode: file.Inode(),
			Type:  fuse.DT_File,
			Name:  recipe.Name,
		}

		files[recipe.Name] = file
		dirents = append(dirents, dirent)
	}

	dir.files = files

	return dirents, nil
}

func (dir *Dir) ReloadProject(project dssapi.Project) {
	dir.project = project
}

func (dir *Dir) getFile(name string) *File {
	if file, exists := dir.files[name]; exists {
		return file
	}

	return nil
}
