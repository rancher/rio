package stringers

import (
	"strconv"

	"github.com/rancher/wrangler/pkg/kv"
)

func ParseUserGroup(user string, group string) (uid *int64, gid *int64, err error) {
	uidStr, gidStr := kv.Split(user, ":")
	if gidStr == "" {
		gidStr = group
	}

	if gidStr != "" {
		gidNum, err := strconv.ParseInt(gidStr, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		gid = &gidNum
	}

	if uidStr != "" {
		uidNum, err := strconv.ParseInt(uidStr, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		uid = &uidNum
	}

	return
}
