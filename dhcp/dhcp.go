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

package dhcp

import "bytes"
import "strconv"
import "repository"

// computes the conical row major for the given
// ordinal. returns an array of spatial indices
// that constitutes an address in the address 
// space. 
func conical (ordinal int64) [6]int64 {
  var bounds = [6]int64 { 256, 256, 256, 256, 256, 256 }
  var output = [6]int64 { 0, 0, 0, 0, 0, 0 }
  var extent = [1]int64 { 1 }
  for i := 0; i < len(output); i++ {
    if(i > 0) { extent[0] *= bounds[i - 1] }
    output[i] = ordinal / extent[0] % bounds[i]
  }
  return output
}

// formats the given ordinal as a address string,
// returns a IP like 6 component vector string 
// given to users on connect.
func format(ordinal int64) string {
  var address = conical(ordinal)
  var buffer bytes.Buffer
  for i := 0; i < len(address); i++ {
    buffer.WriteString(strconv.FormatInt(address[i], 10))
    if(i != len(address) - 1) {
      buffer.WriteString(".")
    }
  }
  return buffer.String()
}

type AddressAllocator interface {
  Next() (string, error)
}
type VirtualAddressAllocator struct {
  repository repository.Repository
}
// returns the next address in this space.
func (allocator VirtualAddressAllocator) Next() (string, error) {
  if result, err := allocator.repository.GetDhcpOrdinal(); err != nil {
    return "", err
  } else {
    address := format(result)
    result += 1
    if err := allocator.repository.SetDhcpOrdinal(result); err != nil {
      return "", err
    } else {
      return address, nil
    }
  }
}
// creates a new virtual address allocator.
func NewVirtualAddressAllocator(repository repository.Repository) * VirtualAddressAllocator {
  allocator := new(VirtualAddressAllocator)
  allocator.repository = repository
  return allocator
}