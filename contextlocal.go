package cls_skywalking_client_go

import (
	"github.com/labstack/echo/v4"
	"github.com/petermattis/goid"
	"sync"
	"time"
)

var (
	contexts = map[int64]echo.Context{}
	rwm      sync.RWMutex
)

// Set 设置一个 context
func SetContext(context echo.Context) {
	if context == nil {
		return
	}
	goID := getGoID()
	rwm.Lock()
	defer rwm.Unlock()

	context.Set("time", time.Now())
	contexts[goID] = context
}

// Get 返回设置的 context
func GetContext() echo.Context {
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

func ClearContextAtRegularTime() {
	t := time.NewTicker(120 * time.Second)
	defer t.Stop()
	for {
		<-t.C
		doClearContextAtRegularTime()
		t.Reset(120 * time.Second)
	}
}

func doClearContextAtRegularTime() {
	rwm.Lock()
	defer rwm.Unlock()
	sm, _ := time.ParseDuration("-2m")
	timeBefore := time.Now().Add(sm)

	for k, v := range contexts {
		contextTime := v.Get("time").(time.Time)
		if contextTime.Unix() < timeBefore.Unix() {
			delete(contexts, k)
		}
	}

}


