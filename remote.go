// Copyright (c) TFG Co. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package nano

import (
	"fmt"
	"reflect"

	"github.com/lonnng/nano/logger"
	"github.com/lonnng/nano/protos"
	"github.com/lonnng/nano/route"
)

func processRemoteMessages(threadID int) {
	// TODO need to monitor stuff here to guarantee messages are not being dropped
	for req := range app.rpcServer.GetUnhandledRequestsChannel() {
		// TODO should deserializer be decoupled?
		log.Debugf("(%d) processing message %v", threadID, req.RequestID)
		switch {
		case req.Type == protos.RPCType_Sys:
			r, err := route.Decode(req.Route)
			if err != nil {
				// TODO answer rpc with an error
				continue
			}
			h, ok := handler.handlers[fmt.Sprintf("%s.%s", r.Service, r.Method)]
			if !ok {
				logger.Log.Warnf("nano/handler: %s not found(forgot registered?)", req.Route)
				// TODO answer rpc with an error
				continue
			}
			var data interface{}
			if h.IsRawArg {
				data = req.Data
			} else {
				data = reflect.New(h.Type.Elem()).Interface()
				err := app.serializer.Unmarshal(req.Data, data)
				if err != nil {
					logger.Log.Error("deserialize error", err.Error())
					return
				}
			}

			log.Debugf("SID=%d, Data=%s", req.SessionID, data)
			// backend session

			// need to create agent
			//handler.processMessage()
			// user request proxied from frontend server
		case req.Type == protos.RPCType_User:
			//TODO
			break
		}
	}

}
