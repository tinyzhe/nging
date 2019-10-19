/*
   Nging is a toolbox for webmasters
   Copyright (C) 2018-present  Wenhui Shen <swh@admpub.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package upload

import (
	"sort"
	"strings"

	"github.com/webx-top/echo"

	"github.com/admpub/nging/application/registry/upload/table"
)

type FieldInfo struct {
	Key     string
	Name    string
	checker Checker
}

func (info *FieldInfo) AddChecker(checker Checker) *FieldInfo {
	if checker == nil {
		return info
	}
	if info.checker == nil {
		info.checker = checker
		return info
	}
	oldChecker := info.checker
	info.checker = func(ctx echo.Context, tbl table.TableInfoStorer) (subdir string, name string, err error) {
		subdir, name, err = oldChecker(ctx, tbl)
		if err != nil {
			return
		}
		subdir2, name2, err2 := checker(ctx, tbl)
		if err2 != nil {
			err = err2
		}
		if len(subdir2) > 0 {
			subdir = subdir2
		}
		if len(name2) > 0 {
			name = name2
		}
		return
	}
	return info
}

// SubdirInfo 子目录信息
type SubdirInfo struct {
	Allowed     bool
	Key         string
	Name        string
	NameEN      string
	Description string

	tableName  string
	fieldInfos map[string]*FieldInfo
	checker    Checker
}

func NewSubdirInfo(key, name string, checkers ...Checker) *SubdirInfo {
	var checker Checker
	if len(checkers) > 0 {
		checker = checkers[0]
	}
	return &SubdirInfo{
		Allowed: true,
		Key:     key,
		Name:    name,
		checker: checker,
	}
}

func (i *SubdirInfo) CopyFrom(other *SubdirInfo) *SubdirInfo {
	i.Allowed = other.Allowed
	if len(other.Key) > 0 {
		i.Key = other.Key
	}
	if len(other.Name) > 0 {
		i.Name = other.Name
	}
	if len(other.NameEN) > 0 {
		i.NameEN = other.NameEN
	}
	if len(other.Description) > 0 {
		i.Description = other.Description
	}
	if len(other.tableName) > 0 {
		i.tableName = other.tableName
	}
	if other.checker != nil {
		if i.checker != nil {
			oldChecker := i.checker
			i.checker = func(ctx echo.Context, tbl table.TableInfoStorer) (subdir string, name string, err error) {
				subdir, name, err = oldChecker(ctx, tbl)
				if err != nil {
					return
				}
				subdir2, name2, err2 := other.checker(ctx, tbl)
				if err2 != nil {
					err = err2
				}
				if len(subdir2) > 0 {
					subdir = subdir2
				}
				if len(name2) > 0 {
					name = name2
				}
				return
			}
		} else {
			i.checker = other.checker
		}
	}
	if len(other.fieldInfos) > 0 {
		if i.fieldInfos == nil {
			i.fieldInfos = make(map[string]*FieldInfo)
		}
		for k, in := range other.fieldInfos {
			i.fieldInfos[k] = in
		}
	}
	return i
}

func (i *SubdirInfo) String() string {
	if len(i.Name) > 0 {
		return i.Name
	}
	return i.Key
}

func (i *SubdirInfo) TableName() string {
	if len(i.tableName) == 0 {
		r := strings.SplitN(i.Key, `-`, 2)
		switch len(r) {
		case 2:
			i.SetFieldName(r[1])
			fallthrough
		case 1:
			i.tableName = r[0]
		}
	}
	return i.tableName
}

func (i *SubdirInfo) ValidFieldName(fieldName string) bool {
	if i.fieldInfos == nil {
		return false
	}
	_, ok := i.fieldInfos[fieldName]
	return ok
}

func (i *SubdirInfo) FieldNames() []string {
	fieldNames := make([]string, len(i.fieldInfos))
	if i.fieldInfos == nil {
		return fieldNames
	}
	var index int
	for fieldName := range i.fieldInfos {
		fieldNames[index] = fieldName
		index++
	}
	sort.Strings(fieldNames)
	return fieldNames
}

func (i *SubdirInfo) FieldInfos() []echo.KV {
	fieldNames := i.FieldNames()
	r := make([]echo.KV, len(fieldNames))
	for index, fieldName := range fieldNames {
		info := i.fieldInfos[fieldName]
		r[index] = echo.KV{K: fieldName, V: info.Name}
	}
	return r
}

func (i *SubdirInfo) SetTableName(tableName string) *SubdirInfo {
	i.tableName = tableName
	return i
}

func (i *SubdirInfo) SetTable(tableName string, fieldNames ...string) *SubdirInfo {
	i.tableName = tableName
	if len(fieldNames) > 0 {
		i.SetFieldName(fieldNames...)
	}
	return i
}

func (i *SubdirInfo) SetFieldName(fieldNames ...string) *SubdirInfo {
	i.fieldInfos = map[string]*FieldInfo{}
	for _, fieldName := range fieldNames {
		i.AddFieldName(fieldName)
	}
	return i
}

func (i *SubdirInfo) parseFieldInfo(field string) (fieldName string, fieldText string) {
	r := strings.SplitN(field, ":", 2)
	switch len(r) {
	case 2:
		fieldText = r[1]
		fallthrough
	case 1:
		fieldName = r[0]
	}
	return
}

func (i *SubdirInfo) AddFieldName(fieldName string, checkers ...Checker) *SubdirInfo {
	var (
		fieldText string
		checker   Checker
	)
	if len(checkers) > 0 {
		checker = checkers[0]
	}
	fieldName, fieldText = i.parseFieldInfo(fieldName)
	info, ok := i.fieldInfos[fieldName]
	if !ok {
		i.fieldInfos[fieldName] = &FieldInfo{
			Key:     fieldName,
			Name:    fieldText,
			checker: checker,
		}
	} else {
		if len(fieldText) > 0 {
			info.Name = fieldText
		}
		info.AddChecker(checker)
	}
	return i
}

func (i *SubdirInfo) GetNameEN() string {
	if len(i.NameEN) > 0 {
		return i.NameEN
	}
	return i.Name
}

func (i *SubdirInfo) SetChecker(checker Checker, fieldNames ...string) *SubdirInfo {
	if len(fieldNames) > 0 {
		if i.fieldInfos == nil {
			i.fieldInfos = make(map[string]*FieldInfo)
		}
		info, ok := i.fieldInfos[fieldNames[0]]
		if !ok {
			panic(`not found: ` + i.Key + `.` + fieldNames[0])
		}
		info.AddChecker(checker)
		return i
	}
	i.checker = checker
	return i
}

func (i *SubdirInfo) Checker() Checker {
	return i.checker
}

func (i *SubdirInfo) MustChecker() Checker {
	if i.checker != nil {
		return i.checker
	}
	return func(ctx echo.Context, tab table.TableInfoStorer) (subdir string, name string, err error) {
		tab.SetTableName(i.TableName())
		if !i.ValidFieldName(tab.FieldName()) {
			err = table.ErrInvalidFieldName
		}
		checker := i.fieldInfos[tab.FieldName()].checker
		if checker == nil {
			checker = DefaultChecker
		}
		subdir, name, err = checker(ctx, tab)
		return
	}
}

func (i *SubdirInfo) SetAllowed(allowed bool) *SubdirInfo {
	i.Allowed = allowed
	return i
}

func (i *SubdirInfo) SetName(name string) *SubdirInfo {
	i.Name = name
	return i
}

func (i *SubdirInfo) SetNameEN(nameEN string) *SubdirInfo {
	i.NameEN = nameEN
	return i
}

func (i *SubdirInfo) SetDescription(description string) *SubdirInfo {
	i.Description = description
	return i
}
