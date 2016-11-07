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

package hub

import (
    "fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"appengine"
	"appengine/channel"
	"dhcp"
	"repository"
	"encryption"
)

// api error constants.
const (
    InternalServerError              = 600
    ConnectAddressAllocationError    = 700
    ConnectChannelInitializeError    = 701
    ConnectIdentitySerializeError    = 702
    ConnectEncryptionError           = 703
    ForwardHttpStreamError           = 800
    ForwardDeserializeError          = 801
    ForwardDecryptionError           = 802
    ForwardDeserializeIdentityError  = 803
    ForwardIdentityVerificationError = 804
    ForwardSerializeError            = 805   
)
var errorText = map[int16] string {
    InternalServerError              : "internal server error.",
    ConnectAddressAllocationError    : "unable to allocate address.",
    ConnectChannelInitializeError    : "unable to initialize data channel.",
    ConnectEncryptionError           : "unable to encrypt identity.",
    ForwardHttpStreamError           : "unable to read from http input stream.",
    ForwardDeserializeError          : "unable to deserialize user request.",
    ForwardDecryptionError           : "unable to decrypt user identity",
    ForwardDeserializeIdentityError  : "unable to deserialize identity",
    ForwardIdentityVerificationError : "unable to verify user identity.",
    ForwardSerializeError            : "unable to serialize forwarded message.",
}

type Error struct {
    Code    int16        `json:"code"`
    Message string       `json:"message"`
}
type RequestError struct {
    Error   Error        `json:"error"`
}
type RequestOk struct {
    Data    interface {} `json:"data"`
}

func init() {
    http.Handle("/connect", Cors(http.HandlerFunc(connect)))
    http.Handle("/forward", Cors(http.HandlerFunc(forward)))
}

// cross origin middleware.
func Cors(next http.Handler) http.Handler {
    fc := func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Origin, Accept, X-Requested-With, Content-Type")
        if r.Method == "OPTIONS" {
            w.WriteHeader(200)
            w.Write([]byte(""))
            return
        }
        next.ServeHTTP(w, r)
    }
    return http.HandlerFunc(fc)
}

// writes a standard api json error on the given response.
func WriteError (w http.ResponseWriter, code int16) {
    output := RequestError {
        Error: Error {
            Code    : code,
            Message : errorText[code],
        },
    }
    if json, err := json.MarshalIndent(output, "", " "); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(500)
        w.Write([]byte(fmt.Sprintf("{error:{ \"code\": %d, \"message\": \"%s\" }}", 
            InternalServerError, 
            errorText[InternalServerError],
        )))
    } else {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(500)
        w.Write(json)
    }
}

// writes a standard api json ok on the given response.
func WriteOk (w http.ResponseWriter, data interface {}) {
    output := RequestOk { 
        Data: data,
    }
    if json, err := json.MarshalIndent(output, "", " "); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(500)
        w.Write([]byte(fmt.Sprintf("{error:{ \"code\": %d, \"message\": \"%s\" }}", 
            InternalServerError, 
            errorText[InternalServerError],
        )))
    } else {    
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(200)
        w.Write(json)
    }
}

// connection identity. generated on connect
// and passed on forward, this struct is encrypted
// between client and server and used to verify
// the identity of the user forwarding messages.
type Identity struct {
    RemoteAddr string `json:"remoteAddr"`
    Address    string `json:"address"`
}

type ConnectResponse struct {
    Channel   string `json:"channel"`
    Identity  string `json:"identity"`
    Address   string `json:"address"`
}

// creates a new connection to this hub. 
func connect (w http.ResponseWriter, r *http.Request) {
    context    := appengine.NewContext(r)
    repository := repository.NewAppEngineRepository  (context)
    allocator  := dhcp.NewVirtualAddressAllocator    (repository)
    encryption := encryption.NewAesEncryptionProvider(repository)

    // allocate new address.
    if address, err := allocator.Next(); err != nil {
        WriteError(w, ConnectAddressAllocationError)
    } else {

        // create new channel.
        if channel_token, err := channel.Create(context, address); err != nil {
            WriteError(w, ConnectChannelInitializeError)
        } else {

            // create identity for user.
            if identity, err := json.Marshal( Identity {RemoteAddr: r.RemoteAddr, Address: address}); err != nil {
                WriteError(w, ConnectIdentitySerializeError)
            } else {

                // encrypt the user identity.
                if identity_token, err := encryption.Encrypt(string(identity)); err != nil {
                    WriteError(w, ConnectEncryptionError)
                } else {

                    // respond.
                    WriteOk(w, ConnectResponse { 
                        Channel : channel_token, 
                        Identity: identity_token,
                        Address : address,
                    })
                }
            }
        }
    }
}

type ForwardRequest struct {
    Identity string  `json:"identity"`
    To       string  `json:"to"`
    Data     string  `json:"data"`
}
type ForwardResponse struct {
    Ok        bool   `json:"ok"`
}
type ForwardOutput struct {
    From     string `json:"from"`
    To       string `json:"to"`
    Data     string `json:"data"`
}

// forwards a request onto another user connected to the hub.
func forward(w http.ResponseWriter, r *http.Request) {
    context    := appengine.NewContext(r)
    repository := repository.NewAppEngineRepository  (context)
    encryption := encryption.NewAesEncryptionProvider(repository)
    
    // read http content.
    defer r.Body.Close()
    if content, err := ioutil.ReadAll(r.Body); err != nil {
        WriteError(w, ForwardHttpStreamError)
    } else {

        // deserialize message.
        var request ForwardRequest
        if err := json.Unmarshal(content, &request); err != nil {
            WriteError(w, ForwardDeserializeError)
        } else {

            // decrypt identity token.
            if identity_token, err := encryption.Decrypt(request.Identity); err != nil {
                 WriteError(w, ForwardDecryptionError)
            } else {
                
                // deserialize identity from token.
                var identity Identity
                if err := json.Unmarshal([]byte(identity_token), &identity); err != nil {
                    WriteError(w, ForwardDeserializeIdentityError)
                } else {

                    // validate request and identity remote address.
                    if identity.RemoteAddr != r.RemoteAddr {
                        WriteError(w, ForwardIdentityVerificationError)
                    } else {

                        // create forwarded message.
                        message := ForwardOutput { 
                            From   : identity.Address, 
                            To     : request.To,
                            Data   : request.Data,
                        }
                        if output, err := json.Marshal(message); err != nil {
                            WriteError(w, ForwardSerializeError)
                        } else {

                            // emit to channel and respond ok.
                            channel.Send(context, request.To, string(output))
                            WriteOk(w, ForwardResponse {  Ok: true, })
                        }
                    }
                }
            }
        }
    }
}

