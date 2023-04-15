package interactive

import "strings"

type HistoryMgr struct {
	list []string
	pos  int
	last string
}

func (mgr *HistoryMgr) Push(cmd string) {
	trimmed := strings.TrimSpace(cmd)
	if len(trimmed) == 0 {
		return
	}
	mgr.list = append(mgr.list, trimmed)
	mgr.pos = 0
}

func (mgr *HistoryMgr) GetPrevious(curr string) string {
	if len(mgr.list) == 0 || mgr.pos == len(mgr.list) {
		return ""
	}

	if mgr.pos == 0 {
		mgr.last = curr
	}

	cmd := mgr.list[len(mgr.list)-mgr.pos-1]
	if mgr.pos != len(mgr.list)-1 {
		mgr.pos++
	}
	return cmd
}

func (mgr *HistoryMgr) GetNext() string {
	if len(mgr.list) == 0 {
		return ""
	}

	if mgr.pos == 0 {
		return mgr.last
	}

	mgr.pos--
	return mgr.list[len(mgr.list)-mgr.pos-1]
}
