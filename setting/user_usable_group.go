package setting

import (
	"encoding/json"
	"fmt"
	"one-api/common"
	"one-api/lang"
)

var userUsableGroups = map[string]string{
	"default": lang.T(nil, "group.default"),
	"vip":     lang.T(nil, "group.vip"),
}

func GetUserUsableGroupsCopy() map[string]string {
	copyUserUsableGroups := make(map[string]string)
	for k, v := range userUsableGroups {
		copyUserUsableGroups[k] = v
	}
	return copyUserUsableGroups
}

func UserUsableGroups2JSONString() string {
	jsonBytes, err := json.Marshal(userUsableGroups)
	if err != nil {
		common.SysError(fmt.Sprintf(lang.T(nil, "group.error.marshal"), err.Error()))
	}
	return string(jsonBytes)
}

func UpdateUserUsableGroupsByJSONString(jsonStr string) error {
	userUsableGroups = make(map[string]string)
	return json.Unmarshal([]byte(jsonStr), &userUsableGroups)
}

func GetUserUsableGroups(userGroup string) map[string]string {
	groupsCopy := GetUserUsableGroupsCopy()
	if userGroup == "" {
		if _, ok := groupsCopy["default"]; !ok {
			groupsCopy["default"] = "default"
		}
	}
	// 如果userGroup不在UserUsableGroups中，返回UserUsableGroups + userGroup
	if _, ok := groupsCopy[userGroup]; !ok {
		groupsCopy[userGroup] = lang.T(nil, "group.user")
	}
	// 如果userGroup在UserUsableGroups中，返回UserUsableGroups
	return groupsCopy
}

func GroupInUserUsableGroups(groupName string) bool {
	_, ok := userUsableGroups[groupName]
	return ok
}
