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

package cluster

import (
	"encoding/json"

	"github.com/lonnng/nano/logger"
)

// Server struct
type Server struct {
	ID   string
	Type string
	Data map[string]string
}

// NewServer ctor
func NewServer(id, serverType string, data ...map[string]string) *Server {
	d := make(map[string]string)
	if len(data) > 0 {
		d = data[0]
	}
	return &Server{
		ID:   id,
		Type: serverType,
		Data: d,
	}
}

// GetDataAsJSONString returns data as a json string
func (s *Server) GetDataAsJSONString() string {
	if s.Data == nil {
		return "{}"
	}
	str, err := json.Marshal(s.Data)
	if err != nil {
		logger.Log.Errorf("error getting server data as json: %s", err.Error())
		return ""
	}
	return string(str)
}
