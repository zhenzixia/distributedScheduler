
package electionmanager

import (
	"flag"
	"github.com/coreos/go-etcd/etcd"
)

const (
	EtcdErrorCodeNotFound     = 100
	EtcdErrorCodeIndexCleared = 401
)

var (
	EtcdErrorNotFound = &etcd.EtcdError{ErrorCode: EtcdErrorCodeNotFound}
)

// IsEtcdNotFound checks if the err is a not found error.
func IsEtcdNotFound(err error) bool {
	etcdErr, ok := err.(*etcd.EtcdError)
	return ok && etcdErr != nil && etcdErr.ErrorCode == EtcdErrorCodeNotFound
}

// IsEtcdWatchStoppedByUser checks if err indicates watch stopped by user.
func IsEtcdWatchStoppedByUser(err error) bool {
	return etcd.ErrWatchStoppedByUser == err
}

func IsEtcdEventIndexCleared(err error) bool {
	etcdErr, ok := err.(*etcd.EtcdError)
	return ok && etcdErr != nil && etcdErr.ErrorCode == EtcdErrorCodeIndexCleared
}

// Etcd client interface.
type Registry interface {
	// Add a new file with a random etcd-generated key under the given path
	AddChild(key string, value string, ttl uint64) (*etcd.Response, error)
	// Get gets the file or directory associated with the given key.
	// If the key points to a directory, files and directories under
	// it will be returned in sorted or unsorted order, depending on
	// the sort flag.
	// If recursive is set to false, contents under child directories
	// will not be returned.
	// If recursive is set to true, all the contents will be returned.
	Get(key string, sort, recursive bool) (*etcd.Response, error)
	// Set sets the given key to the given value.
	// It will create a new key value pair or replace the old one.
	// It will not replace a existing directory.
	Set(key string, value string, ttl uint64) (*etcd.Response, error)
	// Update updates the given key to the given value. It succeeds only if the given key
	// already exists.
	Update(key string, value string, ttl uint64) (*etcd.Response, error)
	// Create creates a file with the given value under the given key. It succeeds
	// only if the given key does not yet exist.
	Create(key string, value string, ttl uint64) (*etcd.Response, error)
	// CreateInOrder creates a file with a key that's guaranteed to be higher than other
	// keys in the given directory. It is useful for creating queues.
	CreateInOrder(dir string, value string, ttl uint64) (*etcd.Response, error)
	// CreateDir create a driectory. It succeeds only if the given key
	// does not yet exist.
	CreateDir(key string, ttl uint64) (*etcd.Response, error)
	// Compare and swap only if prevValue & prevIndex match
	CompareAndSwap(key string, value string, ttl uint64, prevValue string,
		prevIndex uint64) (*etcd.Response, error)
	// Delete deletes the given key.
	// When recursive set to false, if the key points to a
	// directory the method will fail.
	// When recursive set to true, if the key points to a file,
	// the file will be deleted; if the key points to a directory,
	// then everything under the directory (including all child directories)
	// will be deleted.
	Delete(key string, recursive bool) (*etcd.Response, error)
	// If recursive is set to true the watch returns the first change under the
	// given prefix since the given index.
	// If recursive is set to false the watch returns the first change to the
	// given key since the given index.
	// To watch for the latest change, set waitIndex = 0.
	// If a receiver channel is given, it will be a long-term watch. Watch will
	// block at the channel. After someone receives the channel, it will go on
	// to watch that prefix. If a stop channel is given, the client can close
	// long-term watch using the stop channel.
	Watch(prefix string, waitIndex uint64, recursive bool,
		receiver chan *etcd.Response, stop chan bool) (*etcd.Response, error)
}

var etcdServer = flag.String("etcd-server", "http://127.0.0.1:2379",
	"Etcd service location")

func NewEtcdRegistry() Registry {
	return etcd.NewClient([]string{*etcdServer})
}
