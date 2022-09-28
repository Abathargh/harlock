package interactive

import "strings"

type HistoryMgr struct {
	list []string
	pos  int
	last []rune
}

func (mgr *HistoryMgr) Push(cmd string) {
	if len(strings.TrimSpace(cmd)) == 0 {
		return
	}
	mgr.list = append(mgr.list, cmd)
	mgr.pos = 0
}

// TODO IDEA: must keep state of current yet not returned line
// since the line object gets updated to print the data

func (mgr *HistoryMgr) GetPrevious() string {
	if len(mgr.list) == 0 || mgr.pos == len(mgr.list) {
		return ""
	}
	cmd := mgr.list[len(mgr.list)-mgr.pos-1]
	if mgr.pos != len(mgr.list)-1 {
		mgr.pos++
	}
	return cmd
}

func (mgr *HistoryMgr) GetNext() string {
	if len(mgr.list) == 0 || mgr.pos == 0 {
		return ""
	}
	mgr.pos--
	return mgr.list[len(mgr.list)-mgr.pos-1]
}
