package dssfs

import "sync"

var inodeCount uint64 = 2 // 1 is reserved for root
var inodeMutex = sync.Mutex{}

func nextInode() uint64 {
	inodeMutex.Lock()
	defer inodeMutex.Unlock()

	inode := inodeCount
	inodeCount += 1

	return inode
}
