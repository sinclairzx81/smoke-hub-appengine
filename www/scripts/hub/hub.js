//-------------------------------------------------------
// smoke-io appengine messaging hub 
// client script example.
// sinclair 2016
//-------------------------------------------------------

var hub = hub || {}

hub.http = {
  get: function (endpoint, callback) {
      let xhr = new XMLHttpRequest()
      xhr.open("GET", endpoint)
      xhr.addEventListener("readystatechange", function() {
        if (xhr.readyState === XMLHttpRequest.DONE) {
          switch (xhr.status) {
            case 200:
              callback(JSON.parse(xhr.responseText))
              break;
          }
        }
      })
      xhr.send()
  },
  post: function (endpoint, data, callback) {
      var xhr = new XMLHttpRequest()
      xhr.open("POST", endpoint)
      xhr.setRequestHeader('Content-type', 'application/json')
      xhr.addEventListener("readystatechange", function () {
        if (xhr.readyState === XMLHttpRequest.DONE) {
          switch (xhr.status) {
            case 200:
              callback(JSON.parse(xhr.responseText));
              break;
          }
        }
      })
      xhr.send(JSON.stringify(data))
  }
}

hub.client = function (endpoint, resolve) {
    hub.http.get("./connect", function(response) {
      var listeners  = {}
      var connection = response.data
      var channel    = new goog.appengine.Channel(connection.channel)
      var socket     = channel.open()
      // socket on message.
      socket.onmessage = function (message) {
        listeners["message"] = listeners["message"] || []
        listeners["message"].forEach(function (callback) {
          callback(JSON.parse(message.data))
        })
      }
      // socket on error.
      socket.onerror = function (e) {
        listeners["error"] = listeners["error"] || []
        listeners["error"].forEach(function (callback) {
          callback(e)
        })
      }
      // socket on close.
      socket.onclose = function () {
        listeners["close"] = listeners["close"] || []
        listeners["close"].forEach(function (callback) {
          callback()
        })
      }
      // socket on open
      socket.onopen = function () {
        resolve({
          address: function() {
            return connection.address
          },
          send: function (to, data) {
            hub.http.post("./forward", {
              identity : connection.identity,
              to       : to,
              data     : data
            }, function() {})
          },
          on: function (event, callback) {
            listeners[event] = listeners[event] || []
            listeners[event].push(callback)
          }
        })
      }
    })
}