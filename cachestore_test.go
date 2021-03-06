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

import (
	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"encoding/gob"
	"fmt"
	"github.com/icub3d/appenginetesting"
	"reflect"
	"testing"
)

var c = must(appenginetesting.NewContext(nil))

func must(c *appenginetesting.Context, err error) *appenginetesting.Context {
	if err != nil {
		panic(err)
	}
	return c
}

func init() {
	gob.Register(*new(Struct))
}

type Struct struct {
	I int
}

type PropertyLoadSaver struct {
	S string
}

func (p *PropertyLoadSaver) Load(c <-chan datastore.Property) error {
	if err := datastore.LoadStruct(p, c); err != nil {
		return err
	}
	p.S += ".load"
	return nil
}

func (p *PropertyLoadSaver) Save(c chan<- datastore.Property) error {
	defer close(c)
	c <- datastore.Property{
		Name:  "S",
		Value: p.S + ".save",
	}
	return nil
}

func TestWithStruct(t *testing.T) {
	src := Struct{I: 3}
	key := datastore.NewIncompleteKey(c, "Struct", nil)
	// Put
	key, err := Put(c, key, &src)
	if err != nil {
		t.Fatal(err)
	}
	// Get
	dst := *new(Struct)
	err = Get(c, key, &dst)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(src, dst) {
		t.Fatalf("expected=%#v actual=%#v", src, dst)
	}
	// Delete
	err = Delete(c, key)
	if err != nil {
		t.Fatal(err)
	}
	err = Get(c, key, &dst)
	if err != datastore.ErrNoSuchEntity {
		t.Fatal("expected=%#v actual=%#v", datastore.ErrNoSuchEntity, err)
	}
}

func TestWithStructArray(t *testing.T) {
	src := *new([]Struct)
	key := *new([]*datastore.Key)
	for i := 1; i < 11; i++ {
		src = append(src, Struct{I: i})
		key = append(key, datastore.NewIncompleteKey(c, "Struct", nil))
	}
	// PutMulti
	key, err := PutMulti(c, key, src)
	if err != nil {
		t.Fatal(err)
	}
	// GetMulti
	dst := make([]Struct, len(src))
	err = GetMulti(c, key, dst)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(src, dst) {
		t.Fatalf("expected=%#v actual=%#v", src, dst)
	}
	// DeleteMulti
	err = DeleteMulti(c, key)
	if err != nil {
		t.Fatal(err)
	}
	err = GetMulti(c, key, dst)
	if me, ok := err.(appengine.MultiError); ok {
		for _, e := range me {
			if e != datastore.ErrNoSuchEntity {
				t.Fatal(e)
			}
		}
	} else {
		t.Fatal(err)
	}
}

// TODO test []*S

// TODO test []I

func TestWithPropertyLoadSaver(t *testing.T) {
	src := PropertyLoadSaver{}
	key := datastore.NewIncompleteKey(c, "PropertyLoadSaver", nil)
	// Put
	key, err := Put(c, key, &src)
	if err != nil {
		t.Fatal(err)
	}
	// Get
	dst := *new(PropertyLoadSaver)
	err = Get(c, key, &dst)
	if err != nil {
		t.Fatal(err)
	}
	if dst.S != src.S+".save.load" {
		t.Fatalf("actual=%#v", dst.S)
	}
	// Delete
	err = Delete(c, key)
	if err != nil {
		t.Fatal(err)
	}
	err = Get(c, key, &dst)
	if err != datastore.ErrNoSuchEntity {
		t.Fatal("expected=%#v actual=%#v", datastore.ErrNoSuchEntity, err)
	}
}

func TestWithPropertyLoadSaverArray(t *testing.T) {
	src := *new([]PropertyLoadSaver)
	key := *new([]*datastore.Key)
	for i := 1; i < 11; i++ {
		src = append(src, PropertyLoadSaver{S: fmt.Sprint(i)})
		key = append(key, datastore.NewIncompleteKey(c, "PropertyLoadSaver", nil))
	}
	// PutMulti
	key, err := PutMulti(c, key, src)
	if err != nil {
		t.Fatal(err)
	}
	// GetMulti
	dst := make([]PropertyLoadSaver, len(src))
	err = GetMulti(c, key, dst)
	if err != nil {
		t.Fatal(err)
	}
	for i, d := range dst {
		if d.S != src[i].S+".save.load" {
			t.Fatalf("actual%#v", d.S)
		}
	}
	// DeleteMulti
	err = DeleteMulti(c, key)
	if err != nil {
		t.Fatal(err)
	}
	err = GetMulti(c, key, dst)
	if me, ok := err.(appengine.MultiError); ok {
		for _, e := range me {
			if e != datastore.ErrNoSuchEntity {
				t.Fatal(e)
			}
		}
	} else {
		t.Fatal(err)
	}
}

func TestGetFromMemcache(t *testing.T) {
	src := Struct{I: 3}
	key := datastore.NewIncompleteKey(c, "Struct", nil)
	// Put
	key, err := Put(c, key, &src)
	if err != nil {
		t.Fatal(err)
	}
	// remove from datastore
	err = datastore.Delete(c, key)
	if err != nil {
		t.Fatal(err)
	}
	// Get
	dst := *new(Struct)
	err = Get(c, key, &dst)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(src, dst) {
		t.Fatalf("expected=%#v actual=%#v", src, dst)
	}
	// Delete
	err = Delete(c, key)
	if err != nil {
		t.Fatal(err)
	}
	err = Get(c, key, &dst)
	if err != datastore.ErrNoSuchEntity {
		t.Fatal("expected=%#v actual=%#v", datastore.ErrNoSuchEntity, err)
	}
}

func TestGetFromDatastore(t *testing.T) {
	src := Struct{I: 3}
	key := datastore.NewIncompleteKey(c, "Struct", nil)
	// Put
	key, err := Put(c, key, &src)
	if err != nil {
		t.Fatal(err)
	}
	// remove from memcache
	err = memcache.Delete(c, key.Encode())
	if err != nil {
		t.Fatal(err)
	}
	// Get
	dst := *new(Struct)
	err = Get(c, key, &dst)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(src, dst) {
		t.Fatalf("expected=%#v actual=%#v", src, dst)
	}
	// Delete
	err = Delete(c, key)
	if err != nil {
		t.Fatal(err)
	}
	err = Get(c, key, &dst)
	if err != datastore.ErrNoSuchEntity {
		t.Fatal("expected=%#v actual=%#v", datastore.ErrNoSuchEntity, err)
	}
}
