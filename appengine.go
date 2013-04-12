/*
This file is part of cachestore.

cachestore is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

cachestore is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with cachestore.  If not, see <http://www.gnu.org/licenses/>.
*/

package cachestore

// This file contains code copied from appengine.datastore for memcache key generation for the -multi methods.

import (
	"appengine"
	"appengine/datastore"
	"reflect"
)

const (
	multiArgTypeInvalid multiArgType = iota
	multiArgTypePropertyLoadSaver
	multiArgTypeStruct
	multiArgTypeStructPtr
	multiArgTypeInterface
)

type multiArgType int

var (
	typeOfPropertyLoadSaver = reflect.TypeOf((*datastore.PropertyLoadSaver)(nil)).Elem()
	typeOfPropertyList      = reflect.TypeOf(datastore.PropertyList(nil))
)

// checkMultiArg checks that v has type []S, []*S, []I, or []P, for some struct
// type S, for some interface type I, or some non-interface non-pointer type P
// such that P or *P implements PropertyLoadSaver.
//
// It returns what category the slice's elements are, and the reflect.Type
// that represents S, I or P.
//
// As a special case, PropertyList is an invalid type for v.
func checkMultiArg(v reflect.Value) (m multiArgType, elemType reflect.Type) {
	if v.Kind() != reflect.Slice {
		return multiArgTypeInvalid, nil
	}
	if v.Type() == typeOfPropertyList {
		return multiArgTypeInvalid, nil
	}
	elemType = v.Type().Elem()
	if reflect.PtrTo(elemType).Implements(typeOfPropertyLoadSaver) {
		return multiArgTypePropertyLoadSaver, elemType
	}
	switch elemType.Kind() {
	case reflect.Struct:
		return multiArgTypeStruct, elemType
	case reflect.Interface:
		return multiArgTypeInterface, elemType
	case reflect.Ptr:
		elemType = elemType.Elem()
		if elemType.Kind() == reflect.Struct {
			return multiArgTypeStructPtr, elemType
		}
	}
	return multiArgTypeInvalid, nil
}

// multiValid is a batch version of Key.valid. It returns an error, not a
// []bool.
func multiValid(key []*datastore.Key) error {
	invalid := false
	for _, k := range key {
		if !valid(k) {
			invalid = true
			break
		}
	}
	if !invalid {
		return nil
	}
	err := make(appengine.MultiError, len(key))
	for i, k := range key {
		if !valid(k) {
			err[i] = datastore.ErrInvalidKey
		}
	}
	return err
}

// valid returns whether the key is valid.
func valid(k *datastore.Key) bool {
	if k == nil {
		return false
	}
	for ; k != nil; k = k.Parent() {
		if k.Kind() == "" || k.AppID() == "" {
			return false
		}
		if k.StringID() != "" && k.IntID() != 0 {
			return false
		}
		if k.Parent() != nil {
			if k.Parent().Incomplete() {
				return false
			}
			if k.Parent().AppID() != k.AppID() { // namespace is unexported and can't be checked
				return false
			}
		}
	}
	return true
}
