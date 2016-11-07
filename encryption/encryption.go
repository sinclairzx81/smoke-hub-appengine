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

package encryption

import "io"
import "encoding/base64"
import "crypto/rand"
import "crypto/aes"
import "crypto/cipher"
import "repository"
import "errors"

type EncryptionProvider interface {
  // encrypts the given plain text input, returns base64 result.
  Encrypt(input string) (string, error)
  // decrypts the given base64 input, returns plain text result.
  Decrypt(input string) (string, error)
}

type Aes256EncryptionProvider struct {
  repository repository.Repository
}

// encrypts the given plain text input, returns base64 result.
func (provider Aes256EncryptionProvider) Encrypt(input string) (string, error) {
  var bytes = []byte(input)
  if key, err := provider.repository.GetSecretKey(); err != nil {
    return "", err
  } else {
    if block, err := aes.NewCipher(key); err != nil {
      return "", err
    } else {
      ciphertext := make([]byte, aes.BlockSize + len(bytes))
      iv         := ciphertext[:aes.BlockSize]
      if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
      } else {
        cfb := cipher.NewCFBEncrypter(block, iv)
        cfb.XORKeyStream(ciphertext [aes.BlockSize:], bytes)
        return base64.URLEncoding.EncodeToString(ciphertext), nil
      }
    }
  }
}

// decrypts the given base64 input, returns plain text result.
func (provider Aes256EncryptionProvider) Decrypt(input string) (string, error) {
  if len(input) == 0 {
    return "", errors.New("can not decrypt input with 0 length.")
  }
  if bytes, err := base64.URLEncoding.DecodeString(input); err != nil {
    return "", err
  } else {
    if key, err := provider.repository.GetSecretKey(); err != nil {
      return "", err
    } else {
      if block, err := aes.NewCipher(key); err != nil {
        return "", err
      } else {
        iv    := bytes[:aes.BlockSize]
        bytes := bytes[aes.BlockSize:]
        cfb   := cipher.NewCFBDecrypter(block, iv)
        cfb.XORKeyStream(bytes, bytes)
        return string(bytes), nil
      }
    }
  }
}

// creates a new aes encryption proovder.
func NewAesEncryptionProvider (repository repository.Repository) * Aes256EncryptionProvider {
  var provider = new(Aes256EncryptionProvider)
  provider.repository = repository
  return provider
}