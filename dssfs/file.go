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

type File struct {
	inode   uint64
	recipe  dssapi.Recipe
	api     *dssapi.Client
	handle  uint64
	content []byte
}

// Don't want multiple files being loaded from Dataiku at the same time
var fileLoadLock sync.Mutex

func init() {
	fileLoadLock = sync.Mutex{}
}

func NewFile(recipe dssapi.Recipe, api *dssapi.Client) *File {
	return &File{
		inode:   nextInode(),
		recipe:  recipe,
		api:     api,
		content: []byte{},
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
	if err := file.load(); err != nil {
		return err
	}

	attr.Inode = file.inode
	attr.Mode = 0o644
	attr.Ctime = file.recipe.CreatedOn()
	attr.Mtime = file.recipe.ModifiedOn()
	attr.Size = uint64(len(file.content))
	attr.Uid = uint32(os.Getuid())
	attr.Gid = uint32(os.Getgid())

	return nil
}

func (file *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Printf("Opening file (%s)", file.Name())
	if len(file.content) == 0 {
		if err := file.load(); err != nil {
			return nil, err
		}
	}

	if file.handle == 0 {
		file.handle = nextFileHandle()
	}

	resp.Handle = fuse.HandleID(file.handle)

	return file, nil
}

func (file *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	log.Printf("Closing file (%s)", file.Name())
	if file.handle == 0 || file.handle != uint64(req.Handle) {
		return syscall.ENOTSUP
	}

	if req.ReleaseFlags&fuse.ReleaseFlush != 0 {
		if err := file.save(); err != nil {
			return err
		}
	}

	file.handle = 0
	file.content = []byte{}

	return nil
}

func (file *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	log.Printf("Reading file(%s): %d bytes from %d", file.Name(), req.Size, req.Offset)
	if file.handle == 0 {
		return syscall.ENOTSUP
	}

	content := file.content[req.Offset:]

	n := req.Size
	if n > len(content) {
		n = len(content)
	}

	resp.Data = content[0:n]

	return nil
}

func (file *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Printf("Saving file (%s)", file.Name())
	// TODO: Remove me when ready to go crazy!
	log.Fatal("Error: Write operation not allowed!!!")

	if file.handle == 0 || file.handle != uint64(req.Handle) {
		return syscall.ENOTSUP
	}

	if req.Offset >= int64(len(file.content)) {
		file.content = append(file.content, req.Data...)
	} else {
		file.content = append(file.content[0:req.Offset], req.Data...)
	}

	resp.Size = len(req.Data)

	if err := file.save(); err != nil {
		return err
	}

	return nil
}

func (file *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	if file.handle == 0 || file.handle != uint64(req.Handle) {
		return syscall.ENOTSUP
	}

	if err := file.save(); err != nil {
		return err
	}

	return nil
}

func (file *File) load() error {
	log.Printf("Loading file: %s", file.Name())
	fileLoadLock.Lock()
	defer fileLoadLock.Unlock()

	content, err := file.api.GetRecipePayload(file.recipe)
	if err != nil {
		log.Printf("Error: Failed to retrieve read file (%s): %v", file.Name(), err)
		return syscall.EIO
	}

	file.content = []byte(content)
	log.Printf("Loaded %d bytes from file: %s", len(file.content), file.Name())

	return nil
}

func (file *File) save() error {
	if err := file.api.SaveRecipePayload(file.recipe, string(file.content)); err != nil {
		log.Printf("Error: Failed to save file(%s): %v", file.Name(), err)
		return syscall.EIO
	}

	return nil
}

var fileHandlesCount uint64 = 1
var fileHandlesMutex = sync.Mutex{}

func nextFileHandle() uint64 {
	fileHandlesMutex.Lock()
	defer fileHandlesMutex.Unlock()

	handle := fileHandlesCount
	fileHandlesCount++

	return handle
}
