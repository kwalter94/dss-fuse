package dssfs

import (
	"context"

	"bazil.org/fuse"
	"github.com/kwalter94/dss-fuse/dssapi"
)

type File struct {
	inode  uint64
	recipe dssapi.Recipe
	api    *dssapi.Client
}

func NewFile(recipe dssapi.Recipe, api *dssapi.Client) *File {
	return &File{
		inode:  nextInode(),
		recipe: recipe,
		api:    api,
	}
}

func (file *File) Name() string {
	return file.recipe.Name
}

func (file *File) Inode() uint64 {
	return file.inode
}

func (file *File) ReloadRecipe(recipe dssapi.Recipe) {
	file.recipe = recipe
}

func (file *File) Attr(ctx context.Context, attr *fuse.Attr) error {
	attr.Inode = file.inode
	attr.Mode = 0o644
	attr.Ctime = file.recipe.CreatedOn()
	attr.Mtime = file.recipe.ModifiedOn()

	return nil
}
