// Code generated by "stringer -type=Level levels.go"; DO NOT EDIT

package eslog

import "fmt"

const _Level_name = "DebugInfoWarnErrorFatal"

var _Level_index = [...]uint8{0, 5, 9, 13, 18, 23}

func (i Level) String() string {
	if i < 0 || i >= Level(len(_Level_index)-1) {
		return fmt.Sprintf("Level(%d)", i)
	}
	return _Level_name[_Level_index[i]:_Level_index[i+1]]
}
