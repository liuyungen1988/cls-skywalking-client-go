package cls_skywalking_client_go

import (
	"sync"

	"github.com/petermattis/goid"
)

var (
	contexts = map[int64]interface{}{}
	rwm      sync.RWMutex
)

// Set 设置一个 context
func SetContext(context interface{}) {
	if context == nil {
		return
	}
	goID := getGoID()
	rwm.Lock()
	defer rwm.Unlock()

	contexts[goID] = context
}

// Get 返回设置的 context
func GetContext() interface{} {
	goID := getGoID()
	rwm.RLock()
	defer rwm.RUnlock()

	return contexts[goID]
}

// Delete 删除设置的 RequestID
func DeleteContext() {
	goID := getGoID()
	rwm.Lock()
	defer rwm.Unlock()

	delete(contexts, goID)
}

func getGoID() int64 {
	return goid.Get()
}
