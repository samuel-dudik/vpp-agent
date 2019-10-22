// Code generated by protoc-gen-go. DO NOT EDIT.
// source: models/linux/linux.proto

package linux

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	interfaces "github.com/ligato/vpp-agent/api/models/linux/interfaces"
	l3 "github.com/ligato/vpp-agent/api/models/linux/l3"
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

type ConfigData struct {
	Interfaces           []*interfaces.Interface `protobuf:"bytes,10,rep,name=interfaces,proto3" json:"interfaces,omitempty"`
	ArpEntries           []*l3.ARPEntry          `protobuf:"bytes,20,rep,name=arp_entries,json=arpEntries,proto3" json:"arp_entries,omitempty"`
	Routes               []*l3.Route             `protobuf:"bytes,21,rep,name=routes,proto3" json:"routes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *ConfigData) Reset()         { *m = ConfigData{} }
func (m *ConfigData) String() string { return proto.CompactTextString(m) }
func (*ConfigData) ProtoMessage()    {}
func (*ConfigData) Descriptor() ([]byte, []int) {
	return fileDescriptor_0350c4d5cece02f2, []int{0}
}

func (m *ConfigData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfigData.Unmarshal(m, b)
}
func (m *ConfigData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfigData.Marshal(b, m, deterministic)
}
func (m *ConfigData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfigData.Merge(m, src)
}
func (m *ConfigData) XXX_Size() int {
	return xxx_messageInfo_ConfigData.Size(m)
}
func (m *ConfigData) XXX_DiscardUnknown() {
	xxx_messageInfo_ConfigData.DiscardUnknown(m)
}

var xxx_messageInfo_ConfigData proto.InternalMessageInfo

func (m *ConfigData) GetInterfaces() []*interfaces.Interface {
	if m != nil {
		return m.Interfaces
	}
	return nil
}

func (m *ConfigData) GetArpEntries() []*l3.ARPEntry {
	if m != nil {
		return m.ArpEntries
	}
	return nil
}

func (m *ConfigData) GetRoutes() []*l3.Route {
	if m != nil {
		return m.Routes
	}
	return nil
}

type Notification struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Notification) Reset()         { *m = Notification{} }
func (m *Notification) String() string { return proto.CompactTextString(m) }
func (*Notification) ProtoMessage()    {}
func (*Notification) Descriptor() ([]byte, []int) {
	return fileDescriptor_0350c4d5cece02f2, []int{1}
}

func (m *Notification) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Notification.Unmarshal(m, b)
}
func (m *Notification) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Notification.Marshal(b, m, deterministic)
}
func (m *Notification) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Notification.Merge(m, src)
}
func (m *Notification) XXX_Size() int {
	return xxx_messageInfo_Notification.Size(m)
}
func (m *Notification) XXX_DiscardUnknown() {
	xxx_messageInfo_Notification.DiscardUnknown(m)
}

var xxx_messageInfo_Notification proto.InternalMessageInfo

func init() {
	proto.RegisterType((*ConfigData)(nil), "linux.ConfigData")
	proto.RegisterType((*Notification)(nil), "linux.Notification")
}

func init() { proto.RegisterFile("models/linux/linux.proto", fileDescriptor_0350c4d5cece02f2) }

var fileDescriptor_0350c4d5cece02f2 = []byte{
	// 250 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x8f, 0x3f, 0x4b, 0x34, 0x31,
	0x10, 0x87, 0x79, 0x79, 0xf1, 0x8a, 0x39, 0x51, 0x08, 0x0a, 0xcb, 0xd9, 0xc8, 0x35, 0xda, 0x98,
	0xc8, 0xad, 0xdd, 0xd9, 0xf8, 0xe7, 0x0a, 0x1b, 0x91, 0x94, 0x36, 0x32, 0xb7, 0x66, 0xd7, 0x81,
	0x5c, 0x26, 0x64, 0x67, 0x45, 0x3f, 0x91, 0x5f, 0x53, 0xcc, 0xae, 0xac, 0x56, 0x67, 0x33, 0x4c,
	0x32, 0xcf, 0x93, 0xcc, 0x0f, 0x8a, 0x0d, 0x3f, 0x3b, 0xdf, 0x1a, 0x4f, 0xa1, 0x7b, 0xeb, 0xab,
	0x8e, 0x89, 0x85, 0xd5, 0x4e, 0x3e, 0xcc, 0x2e, 0x3d, 0x35, 0x28, 0x6c, 0x5e, 0x63, 0x3c, 0xc3,
	0xc6, 0x05, 0x31, 0x18, 0xc9, 0xfc, 0xb2, 0x28, 0x88, 0x4b, 0x35, 0x56, 0xae, 0x1d, 0xdb, 0xfe,
	0x91, 0x99, 0xde, 0x6e, 0xfb, 0xd2, 0x60, 0x8a, 0x03, 0x7f, 0xfe, 0x27, 0x3e, 0x71, 0x27, 0xc3,
	0x0f, 0xf3, 0x8f, 0x7f, 0x00, 0x37, 0x1c, 0x6a, 0x6a, 0x6e, 0x51, 0x50, 0x2d, 0x01, 0xc6, 0x75,
	0x0a, 0x38, 0xfe, 0x7f, 0x3a, 0x5d, 0x1c, 0xe9, 0x3e, 0xd7, 0x38, 0xd0, 0x77, 0xdf, 0xad, 0xfd,
	0x81, 0xab, 0x12, 0xa6, 0x98, 0xe2, 0x93, 0x0b, 0x92, 0xc8, 0xb5, 0xc5, 0x41, 0xb6, 0xd5, 0x60,
	0xfb, 0x52, 0x5f, 0xd9, 0x87, 0x55, 0x90, 0xf4, 0x6e, 0x01, 0x53, 0x5c, 0xf5, 0x94, 0x3a, 0x81,
	0x49, 0xde, 0xa7, 0x2d, 0x0e, 0x33, 0xbf, 0x3f, 0xf2, 0xf6, 0xeb, 0xde, 0x0e, 0xe3, 0xf9, 0x1e,
	0xec, 0xde, 0xb3, 0x50, 0x4d, 0x15, 0x0a, 0x71, 0xb8, 0xbe, 0x78, 0x5c, 0x34, 0x24, 0x2f, 0xdd,
	0x5a, 0x57, 0xbc, 0x31, 0x5b, 0x83, 0x2f, 0x73, 0x5d, 0x4f, 0x72, 0xec, 0xf2, 0x33, 0x00, 0x00,
	0xff, 0xff, 0xa3, 0xbc, 0xd3, 0x29, 0xb9, 0x01, 0x00, 0x00,
}
