// Code generated by protoc-gen-go. DO NOT EDIT.
// source: registry.proto

package regrpc

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_41af05d40a615591, []int{0}
}

func (m *Empty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Empty.Unmarshal(m, b)
}
func (m *Empty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Empty.Marshal(b, m, deterministic)
}
func (m *Empty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Empty.Merge(m, src)
}
func (m *Empty) XXX_Size() int {
	return xxx_messageInfo_Empty.Size(m)
}
func (m *Empty) XXX_DiscardUnknown() {
	xxx_messageInfo_Empty.DiscardUnknown(m)
}

var xxx_messageInfo_Empty proto.InternalMessageInfo

type Session struct {
	Sessionid            string   `protobuf:"bytes,1,opt,name=sessionid,proto3" json:"sessionid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Session) Reset()         { *m = Session{} }
func (m *Session) String() string { return proto.CompactTextString(m) }
func (*Session) ProtoMessage()    {}
func (*Session) Descriptor() ([]byte, []int) {
	return fileDescriptor_41af05d40a615591, []int{1}
}

func (m *Session) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Session.Unmarshal(m, b)
}
func (m *Session) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Session.Marshal(b, m, deterministic)
}
func (m *Session) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Session.Merge(m, src)
}
func (m *Session) XXX_Size() int {
	return xxx_messageInfo_Session.Size(m)
}
func (m *Session) XXX_DiscardUnknown() {
	xxx_messageInfo_Session.DiscardUnknown(m)
}

var xxx_messageInfo_Session proto.InternalMessageInfo

func (m *Session) GetSessionid() string {
	if m != nil {
		return m.Sessionid
	}
	return ""
}

type Domain struct {
	Sessionid            string   `protobuf:"bytes,1,opt,name=sessionid,proto3" json:"sessionid,omitempty"`
	Name                 string   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Domain) Reset()         { *m = Domain{} }
func (m *Domain) String() string { return proto.CompactTextString(m) }
func (*Domain) ProtoMessage()    {}
func (*Domain) Descriptor() ([]byte, []int) {
	return fileDescriptor_41af05d40a615591, []int{2}
}

func (m *Domain) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Domain.Unmarshal(m, b)
}
func (m *Domain) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Domain.Marshal(b, m, deterministic)
}
func (m *Domain) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Domain.Merge(m, src)
}
func (m *Domain) XXX_Size() int {
	return xxx_messageInfo_Domain.Size(m)
}
func (m *Domain) XXX_DiscardUnknown() {
	xxx_messageInfo_Domain.DiscardUnknown(m)
}

var xxx_messageInfo_Domain proto.InternalMessageInfo

func (m *Domain) GetSessionid() string {
	if m != nil {
		return m.Sessionid
	}
	return ""
}

func (m *Domain) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type Status struct {
	ReturnCode           int32    `protobuf:"varint,1,opt,name=return_code,json=returnCode,proto3" json:"return_code,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Status) Reset()         { *m = Status{} }
func (m *Status) String() string { return proto.CompactTextString(m) }
func (*Status) ProtoMessage()    {}
func (*Status) Descriptor() ([]byte, []int) {
	return fileDescriptor_41af05d40a615591, []int{3}
}

func (m *Status) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Status.Unmarshal(m, b)
}
func (m *Status) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Status.Marshal(b, m, deterministic)
}
func (m *Status) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Status.Merge(m, src)
}
func (m *Status) XXX_Size() int {
	return xxx_messageInfo_Status.Size(m)
}
func (m *Status) XXX_DiscardUnknown() {
	xxx_messageInfo_Status.DiscardUnknown(m)
}

var xxx_messageInfo_Status proto.InternalMessageInfo

func (m *Status) GetReturnCode() int32 {
	if m != nil {
		return m.ReturnCode
	}
	return 0
}

func init() {
	proto.RegisterType((*Empty)(nil), "Empty")
	proto.RegisterType((*Session)(nil), "Session")
	proto.RegisterType((*Domain)(nil), "Domain")
	proto.RegisterType((*Status)(nil), "Status")
}

func init() { proto.RegisterFile("registry.proto", fileDescriptor_41af05d40a615591) }

var fileDescriptor_41af05d40a615591 = []byte{
	// 250 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x90, 0x3f, 0x4b, 0xc4, 0x40,
	0x10, 0xc5, 0x13, 0xf1, 0x92, 0xdc, 0xdc, 0x21, 0x38, 0xd5, 0x11, 0x04, 0x75, 0x2d, 0xfc, 0x53,
	0x2c, 0xa2, 0x9d, 0x76, 0x7a, 0x87, 0xcd, 0x55, 0x49, 0x67, 0x23, 0xf1, 0x32, 0x84, 0x05, 0x93,
	0x09, 0xbb, 0x13, 0x30, 0x5f, 0xc7, 0x4f, 0x2a, 0xec, 0xe6, 0xd4, 0xce, 0xee, 0xf1, 0xde, 0xdb,
	0x9d, 0xdf, 0x0c, 0x1c, 0x59, 0x6a, 0x8c, 0x13, 0x3b, 0xea, 0xde, 0xb2, 0xb0, 0x4a, 0x61, 0xb6,
	0x69, 0x7b, 0x19, 0xd5, 0x25, 0xa4, 0x25, 0x39, 0x67, 0xb8, 0xc3, 0x13, 0x98, 0xbb, 0x20, 0x4d,
	0xbd, 0x8a, 0xcf, 0xe2, 0xab, 0x79, 0xf1, 0x6b, 0xa8, 0x07, 0x48, 0xd6, 0xdc, 0x56, 0xe6, 0x9f,
	0x1e, 0x22, 0x1c, 0x76, 0x55, 0x4b, 0xab, 0x03, 0x1f, 0x78, 0xad, 0xae, 0x21, 0x29, 0xa5, 0x92,
	0xc1, 0xe1, 0x29, 0x2c, 0x2c, 0xc9, 0x60, 0xbb, 0xb7, 0x1d, 0xd7, 0xe4, 0x5f, 0xcf, 0x0a, 0x08,
	0xd6, 0x33, 0xd7, 0x74, 0xf7, 0x15, 0x43, 0x56, 0x4c, 0xac, 0x78, 0x0e, 0x8b, 0x2d, 0x37, 0xa6,
	0x2b, 0x47, 0x27, 0xd4, 0x62, 0xa2, 0x3d, 0x73, 0x9e, 0xe9, 0x09, 0x59, 0x45, 0x78, 0x03, 0xc7,
	0x2f, 0x24, 0x9b, 0xcf, 0xde, 0x58, 0xaa, 0x03, 0xa0, 0xc3, 0x9f, 0x42, 0x9e, 0xea, 0xe0, 0xa9,
	0xe8, 0x36, 0x46, 0x05, 0xcb, 0x35, 0x7d, 0x90, 0xd0, 0xb4, 0xc8, 0x3e, 0xcc, 0x53, 0x1d, 0xf0,
	0x54, 0x84, 0x17, 0xb0, 0xdc, 0x72, 0xc3, 0x83, 0x4c, 0x33, 0xff, 0x7e, 0xb5, 0x2f, 0x3d, 0xc1,
	0x6b, 0xa6, 0x1f, 0x2d, 0x35, 0xb6, 0xdf, 0xbd, 0x27, 0xfe, 0xa0, 0xf7, 0xdf, 0x01, 0x00, 0x00,
	0xff, 0xff, 0xa0, 0xad, 0x03, 0xe6, 0x62, 0x01, 0x00, 0x00,
}