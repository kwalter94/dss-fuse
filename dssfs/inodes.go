package dssfs

var inodeCount uint64 = 2

func nextInode() uint64 {
	inode := inodeCount

	inodeCount += 1

	return inode
}
