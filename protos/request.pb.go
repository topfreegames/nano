// Code generated by protoc-gen-go. DO NOT EDIT.
// source: request.proto

/*
Package protos is a generated protocol buffer package.

It is generated from these files:
	request.proto

It has these top-level messages:
	Request
*/
package protos

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type RPCType int32

const (
	RPCType_Sys  RPCType = 0
	RPCType_User RPCType = 1
)

var RPCType_name = map[int32]string{
	0: "Sys",
	1: "User",
}
var RPCType_value = map[string]int32{
	"Sys":  0,
	"User": 1,
}

func (x RPCType) String() string {
	return proto.EnumName(RPCType_name, int32(x))
}
func (RPCType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Request struct {
	Route      string  `protobuf:"bytes,1,opt,name=Route" json:"Route,omitempty"`
	SessionID  int64   `protobuf:"varint,2,opt,name=SessionID" json:"SessionID,omitempty"`
	RequestID  uint64  `protobuf:"varint,3,opt,name=RequestID" json:"RequestID,omitempty"`
	FrontendID string  `protobuf:"bytes,4,opt,name=FrontendID" json:"FrontendID,omitempty"`
	Type       RPCType `protobuf:"varint,5,opt,name=Type,enum=protos.RPCType" json:"Type,omitempty"`
	Data       []byte  `protobuf:"bytes,6,opt,name=Data,proto3" json:"Data,omitempty"`
}

func (m *Request) Reset()                    { *m = Request{} }
func (m *Request) String() string            { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()               {}
func (*Request) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Request) GetRoute() string {
	if m != nil {
		return m.Route
	}
	return ""
}

func (m *Request) GetSessionID() int64 {
	if m != nil {
		return m.SessionID
	}
	return 0
}

func (m *Request) GetRequestID() uint64 {
	if m != nil {
		return m.RequestID
	}
	return 0
}

func (m *Request) GetFrontendID() string {
	if m != nil {
		return m.FrontendID
	}
	return ""
}

func (m *Request) GetType() RPCType {
	if m != nil {
		return m.Type
	}
	return RPCType_Sys
}

func (m *Request) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*Request)(nil), "protos.Request")
	proto.RegisterEnum("protos.RPCType", RPCType_name, RPCType_value)
}

func init() { proto.RegisterFile("request.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 199 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x8f, 0x3d, 0x4e, 0xc6, 0x30,
	0x0c, 0x40, 0x31, 0xcd, 0xd7, 0x52, 0x8b, 0x9f, 0xca, 0x62, 0xc8, 0x50, 0xa1, 0x08, 0x96, 0x88,
	0xa1, 0x03, 0x1c, 0x81, 0x08, 0xa9, 0x1b, 0x72, 0xe1, 0x00, 0x45, 0x64, 0x60, 0x69, 0x4a, 0x92,
	0x0e, 0xbd, 0x17, 0x07, 0x44, 0x4d, 0x2a, 0x95, 0xc9, 0xf6, 0x7b, 0xb6, 0x65, 0xe3, 0x95, 0xb7,
	0x3f, 0x8b, 0x0d, 0xb1, 0x9b, 0xbd, 0x8b, 0x8e, 0xca, 0x14, 0xc2, 0xfd, 0x2f, 0x60, 0xc5, 0xd9,
	0xd0, 0x2d, 0x9e, 0xd8, 0x2d, 0xd1, 0x4a, 0x50, 0xa0, 0x6b, 0xce, 0x05, 0xb5, 0x58, 0x0f, 0x36,
	0x84, 0x6f, 0x37, 0xf5, 0x46, 0x9e, 0x2b, 0xd0, 0x05, 0x1f, 0x60, 0xb3, 0xfb, 0x78, 0x6f, 0x64,
	0xa1, 0x40, 0x0b, 0x3e, 0x00, 0xdd, 0x21, 0xbe, 0x7a, 0x37, 0x45, 0x3b, 0x7d, 0xf5, 0x46, 0x8a,
	0xb4, 0xf6, 0x1f, 0xa1, 0x07, 0x14, 0xef, 0xeb, 0x6c, 0xe5, 0x49, 0x81, 0xbe, 0x7e, 0xba, 0xc9,
	0xb7, 0x85, 0x8e, 0xdf, 0x5e, 0x36, 0xcc, 0x49, 0x12, 0xa1, 0x30, 0x63, 0x1c, 0x65, 0xa9, 0x40,
	0x5f, 0x72, 0xca, 0x1f, 0x5b, 0xac, 0xf6, 0x26, 0xaa, 0xb0, 0x18, 0xd6, 0xd0, 0x9c, 0xd1, 0x05,
	0x8a, 0x8f, 0x60, 0x7d, 0x03, 0x9f, 0xf9, 0xb9, 0xe7, 0xbf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x16,
	0xd4, 0x6c, 0xb4, 0xf4, 0x00, 0x00, 0x00,
}
