/*
   Copyright (c) 2019 AT&T Intellectual Property.
   Copyright (c) 2018-2019 Nokia.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package sdlgo

import (
	"reflect"
	"strings"

	"gerrit.oran-osc.org/r/ric-plt/sdlgo/internal/sdlgoredis"
)

type Idatabase interface {
	MSet(pairs ...interface{}) error
	MGet(keys []string) ([]interface{}, error)
	Close() error
	Del(keys []string) error
	Keys(key string) ([]string, error)
}

type SdlInstance struct {
	NameSpace string
	NsPrefix  string
	Idatabase
}

func Create(NameSpace string) *SdlInstance {
	db := sdlgoredis.Create()
	s := SdlInstance{
		NameSpace: NameSpace,
		NsPrefix:  "{" + NameSpace + "},",
		Idatabase: db,
	}

	return &s
}

func (s *SdlInstance) Close() error {
	return s.Close()
}

func (s *SdlInstance) setNamespaceToKeys(pairs ...interface{}) []interface{} {
	var retVal []interface{}
	for i, v := range pairs {
		if i%2 == 0 {
			reflectType := reflect.TypeOf(v)
			switch reflectType.Kind() {
			case reflect.Slice:
				x := reflect.ValueOf(v)
				for i2 := 0; i2 < x.Len(); i2++ {
					if i2%2 == 0 {
						retVal = append(retVal, s.NsPrefix+x.Index(i2).Interface().(string))
					} else {
						retVal = append(retVal, x.Index(i2).Interface())
					}
				}
			case reflect.Array:
				x := reflect.ValueOf(v)
				for i2 := 0; i2 < x.Len(); i2++ {
					if i2%2 == 0 {
						retVal = append(retVal, s.NsPrefix+x.Index(i2).Interface().(string))
					} else {
						retVal = append(retVal, x.Index(i2).Interface())
					}
				}
			default:
				retVal = append(retVal, s.NsPrefix+v.(string))
			}
		} else {
			retVal = append(retVal, v)
		}
	}
	return retVal
}

func (s *SdlInstance) Set(pairs ...interface{}) error {
	if len(pairs) == 0 {
		return nil
	}

	keyAndData := s.setNamespaceToKeys(pairs...)
	err := s.MSet(keyAndData...)
	return err
}

func (s *SdlInstance) Get(keys []string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	if len(keys) == 0 {
		return m, nil
	}

	var keysWithNs []string
	for _, v := range keys {
		keysWithNs = append(keysWithNs, s.NsPrefix+v)
	}
	val, err := s.MGet(keysWithNs)
	if err != nil {
		return m, err
	}
	for i, v := range val {
		m[keys[i]] = v
	}
	return m, err
}

func (s *SdlInstance) SetIf(key string, oldData, newData interface{}) {
	panic("SetIf not implemented\n")
}

func (s *SdlInstance) SetIfiNotExists(key string, data interface{}) {
	panic("SetIfiNotExists not implemented\n")
}

func (s *SdlInstance) Remove(keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	var keysWithNs []string
	for _, v := range keys {
		keysWithNs = append(keysWithNs, s.NsPrefix+v)
	}
	err := s.Del(keysWithNs)
	return err
}

func (s *SdlInstance) RemoveIf(key string, data interface{}) {
	panic("RemoveIf not implemented\n")
}

func (s *SdlInstance) GetAll() ([]string, error) {
	keys, err := s.Keys(s.NsPrefix + "*")
	var retVal []string = nil
	if err != nil {
		return retVal, err
	}
	for _, v := range keys {
		retVal = append(retVal, strings.Split(v, s.NsPrefix)[1])
	}
	return retVal, err
}

func (s *SdlInstance) RemoveAll() error {
	keys, err := s.Keys(s.NsPrefix + "*")
	if err != nil {
		return err
	}
	if keys != nil {
		err = s.Del(keys)
	}
	return err
}
