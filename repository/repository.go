/*--------------------------------------------------------------------------

 smoke-hub-appengine - messaging relay for webrtc.

 The MIT License (MIT)

 Copyright (c) 2016 Haydn Paterson (sinclair) <haydn.developer@gmail.com>

 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included in
 all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 THE SOFTWARE.
 
---------------------------------------------------------------------------*/

package repository

import "appengine"
import "appengine/datastore"
import "encoding/base64"
import "crypto/rand"


type Repository interface {
    GetDhcpOrdinal ()              (int64, error)
    SetDhcpOrdinal (ordinal int64) (error)
    GetSecretKey   ()              ([]byte, error)
}

//-----------------------------------------------------
// helper for creating keys on demand
//-----------------------------------------------------
func GenerateRandomBytes(length int) ([]byte, error) {
    b := make([]byte, length)
    _, err := rand.Read(b)
    if err != nil {
        return nil, err
    }
    return b, nil
}

//-----------------------------------------------------
// internally cached secret key. 
//-----------------------------------------------------
var CACHED_SECRET_KEY []byte = nil

// DHCP datastore record.
type DHCP struct {
  Ordinal int64
}

// SECRET datastore record.
type SECRET struct {
  Value string
}

type AppEngineRepository struct {
  context appengine.Context
}
func (repository AppEngineRepository) GetDhcpOrdinal() (int64, error) {
  var key = datastore.NewKey(repository.context, "DHCP", "0", 0, nil)
  var record = new(DHCP)
  var err = datastore.Get(repository.context, key, record)
  if err != nil {
    record.Ordinal = 0  
  }
  return record.Ordinal, nil
}
func (repository AppEngineRepository) SetDhcpOrdinal(ordinal int64) (error) {
  var key = datastore.NewKey(repository.context, "DHCP", "0", 0, nil)
  var record = new(DHCP)
  record.Ordinal = ordinal
  if _, err := datastore.Put(repository.context, key, record); err != nil {
    return err  
  }
  return nil
}
// gets the aes secret key. If the key does not exist, this 
// function will create with a random key.
func (repository AppEngineRepository) GetSecretKey() ([]byte, error) {
  // check the cache for the key.
	if CACHED_SECRET_KEY != nil {
		return CACHED_SECRET_KEY, nil
	}
	var key    = datastore.NewKey(repository.context, "SECRET", "0", 0, nil)
	var record = new(SECRET)
	if err := datastore.Get(repository.context, key, record); err != nil {
    if bytes, err := GenerateRandomBytes(32); err != nil {
			return nil, err
		} else {
      record.Value = base64.URLEncoding.EncodeToString(bytes)
      if _, err := datastore.Put(repository.context, key, record); err != nil {
        return nil, err
      } else {
        CACHED_SECRET_KEY = bytes
        return CACHED_SECRET_KEY, nil
      }
    }
	} else {
		if bytes, err := base64.URLEncoding.DecodeString(record.Value); err != nil {
			return nil, err 
		} else {
      CACHED_SECRET_KEY = bytes
      return CACHED_SECRET_KEY, nil
    }
	}
}

// creates a new appengine datastore backed store.
func NewAppEngineRepository(context appengine.Context) * AppEngineRepository {
  var store = new(AppEngineRepository)
  store.context = context
  return store
}