package electionmanager

import (
	"errors"
	"sync"
	"time"

	"github.com/coreos/go-etcd/etcd"
	"github.com/golang/glog"
)

const kRetrySleep time.Duration = 100 // milliseconds

// Various event types for the events channel.
type MasterEventType int

const (
	MasterAdded MasterEventType = iota // this node has the lock.
	MasterDeleted
)

// MasterEvent represents a single event sent on the events channel.
type MasterEvent struct {
	Type   MasterEventType // event type
	Master string          // identity of the lock holder
}

// Interface used by the etcd master lock clients.
type MasterInterface interface {
	// Start the election and attempt to acquire the lock. If acquired, the
	// lock is refreshed periodically based on the ttl.
	Start()

	// Returns the event channel used by the etcd lock.
	EventsChan() <-chan MasterEvent

	// Method to get the current lockholder. Returns "" if free.
	GetHolder() string
}

// Internal structure to represent an etcd lock.
type electionManager struct {
	sync.Mutex
	client             Registry         // etcd interface
	name               string           // name of the lock
	id                 string           // identity of the lockholder
	ttl                uint64           // ttl of the lock
	lastObservedLeader string           // Lock holder
	watchStopCh        chan bool        // To stop the watch
	eventsCh           chan MasterEvent // channel to send lock ownership updates
	stoppedCh          chan bool        // channel that waits for acquire to finish
	refreshStopCh      chan bool        // channel used to stop the refresh routine
	//holding            bool             // whether this node is holding the lock
	modifiedIndex      uint64           // valid only when this node is holding the lock
}

// Method to create a new etcd lock.
func NewMaster(client Registry, name string, id string, ttl uint64) (MasterInterface, error) {
	// client is mandatory. Min ttl is 5 seconds.
	if client == nil || ttl < 5 {
		return nil, errors.New("Invalid args")
	}

	return &electionManager{client: client, name: name, id: id, ttl: ttl,
		lastObservedLeader:        "",
		watchStopCh:   make(chan bool, 1),
		eventsCh:      make(chan MasterEvent, 1),
		stoppedCh:     make(chan bool, 1),
		refreshStopCh: make(chan bool, 1),
		modifiedIndex: 0,
	}, nil
}

// Method to start the attempt to acquire the lock.
func (e *electionManager) Start() {
	go func() {
		// If acquire returns without error, exit. If not, acquire
		// crashed and needs to be called again.
		for {
			if err := e.acquire(); err == nil {
				glog.Infof("Got the lock, about to break..")
				break
			}
		}
	}()
}

// Method to get the event channel used by the etcd lock.
func (e *electionManager) EventsChan() <-chan MasterEvent {
	return e.eventsCh
}

// Method to get the lockholder.
func (e *electionManager) GetHolder() string {
	e.Lock()
	defer e.Unlock()
	return e.lastObservedLeader
}

// Method to acquire the lock. It launches another goroutine to refresh the ttl
// if successful in acquiring the lock.
func (e *electionManager) acquire() (ret error) {

	var resp *etcd.Response
	var err error

	for {
		//1. Try to get the lock
		resp, err = e.client.Get(e.name, false, false)
		//2. If the lock is empty, try to acquire the lock. There are two possibilities:
		//	1) It is the first time start the electonManager
		//	2) The last ttl is expired, and the lock is deleted by etcd.
		if err != nil {
			if IsEtcdNotFound(err) {
				// Try to acquire the lock.
				glog.Infof("Trying to acquire lock %s", e.name)
				resp, err = e.client.Create(e.name, e.id, e.ttl)
				if err != nil {
					glog.Infof("Failed to acquire the lock... err: %s", err.Error())
					// Failed to acquire the lock.
					continue
				}
			} else {
				glog.Infof("Failed to get lock %s, error: %v",
					e.name, err)
				time.Sleep(kRetrySleep * time.Millisecond)
				continue
			}
		}

		glog.Infof("---- Found the leader node:  %v ----\n", resp.Node.Value)
		//3. If the lock is not empty, check if current node is leader
		//	if current node is leader, start ttl updater if not started,
		//	if current node is not leader, stop ttl updater and jobController (no matter if it was running or not in this node)
		if resp.Node.Value == e.id {
			glog.Infof("Current node is a leader... id: %v", e.id)
			//Current node is a leader
			//last observed leader is not the current node. That means current node just became the leader.
			//Do: 1) send an event to event channel to start the jobController 2) start the ttl updater.
			if e.lastObservedLeader != e.id {
				e.eventsCh <- MasterEvent{Type: MasterAdded,
					Master: e.id}
				go e.ttlUpdater()
			}
			//if last observed leader equals to current node id, do nothing.
		} else {
			glog.Infof("Current node is a slave...")
			glog.Infof("Last observed leader is %v ...", e.lastObservedLeader)
			//Current node is a slave
			//Last observed leader is the current node, that mean current node just lost the lock, need to stop the ttl updater
			if e.lastObservedLeader == e.id {
				glog.Infof("about to send stop signal to eventsCh")
				e.eventsCh <- MasterEvent{Type: MasterDeleted,
					Master: resp.Node.Value}
				glog.Infof("about to send stop signal to refreshStopCH")
				e.refreshStopCh <- true
			}
			//Last observed leader is not the current node, do nothing, continue to wait for the leader's ttl expiration.
		}

		//4. Keep watch the lock for modification.
		//	There are two possible changes: 1) ttl is updated by the updater.
		//					2) ttl expired, and lock is deleted by etcd


		// Record the new master and modified index.
		e.Lock()
		e.lastObservedLeader = resp.Node.Value
		e.Unlock()
		e.modifiedIndex = resp.Node.ModifiedIndex

		var prevIndex uint64

		// Intent is to start the watch using EtcdIndex. Sometimes, etcd
		// is returning EtcdIndex lower than ModifiedIndex. In such
		// cases, use ModifiedIndex to set the watch.
		// TODO: Change this code when etcd behavior changes.
		if resp.EtcdIndex < resp.Node.ModifiedIndex {
			prevIndex = resp.Node.ModifiedIndex + 1
		} else {
			prevIndex = resp.EtcdIndex + 1
		}

		// Start watching for changes to lock.
		resp, err = e.client.Watch(e.name, prevIndex, false, nil, e.watchStopCh)

		if IsEtcdWatchStoppedByUser(err) {
	 		glog.Infof("Watch for lock %s stopped by user", e.name)
		} else if err != nil {
			// Log only if its not too old event index error.
			if !IsEtcdEventIndexCleared(err) {
				glog.Errorf("Failed to watch lock %s, error %v",
					e.name, err)
			}
		}
	}

	return nil
}

// Method to refresh the lock. It refreshes the ttl at ttl*4/10 interval.
func (e *electionManager) ttlUpdater() {
	glog.Infof("ttl updater is called...")
	for {
		select {
		case <-e.refreshStopCh:
			glog.Infof("Stopping refresh for lock %s", e.name)
			// Lock released.
			return
		case <-time.After(time.Second * time.Duration(e.ttl*4/10)):
			glog.Infof("Refresh ttl for %v", e.id)
			// Uses CompareAndSwap to protect against the case where a
			// watch is received with a "delete" and refresh routine is
			// still running.
			if resp, err := e.client.CompareAndSwap(e.name, e.id, e.ttl,
				e.id, e.modifiedIndex); err != nil {
				// Failure here could mean that some other node
				// acquired the lock. Should also get a watch
				// notification if that happens and this go routine
				// is stopped there.
				glog.Errorf("Failed to set the ttl for lock %s with "+
					"error: %s", e.name, err.Error())
			} else {
				e.modifiedIndex = resp.Node.ModifiedIndex
			}
		}
	}
}
