// Code generated by protoc-gen-go. DO NOT EDIT.
// source: models/vpp/abf/abf.proto

package vpp_abf

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

// ABF defines ACL based forwarding.
type ABF struct {
	Index                uint32                   `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	AclName              string                   `protobuf:"bytes,2,opt,name=acl_name,json=aclName,proto3" json:"acl_name,omitempty"`
	AttachedInterfaces   []*ABF_AttachedInterface `protobuf:"bytes,3,rep,name=attached_interfaces,json=attachedInterfaces,proto3" json:"attached_interfaces,omitempty"`
	ForwardingPaths      []*ABF_ForwardingPath    `protobuf:"bytes,4,rep,name=forwarding_paths,json=forwardingPaths,proto3" json:"forwarding_paths,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *ABF) Reset()         { *m = ABF{} }
func (m *ABF) String() string { return proto.CompactTextString(m) }
func (*ABF) ProtoMessage()    {}
func (*ABF) Descriptor() ([]byte, []int) {
	return fileDescriptor_50fc84daf012435f, []int{0}
}

func (m *ABF) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ABF.Unmarshal(m, b)
}
func (m *ABF) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ABF.Marshal(b, m, deterministic)
}
func (m *ABF) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ABF.Merge(m, src)
}
func (m *ABF) XXX_Size() int {
	return xxx_messageInfo_ABF.Size(m)
}
func (m *ABF) XXX_DiscardUnknown() {
	xxx_messageInfo_ABF.DiscardUnknown(m)
}

var xxx_messageInfo_ABF proto.InternalMessageInfo

func (m *ABF) GetIndex() uint32 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *ABF) GetAclName() string {
	if m != nil {
		return m.AclName
	}
	return ""
}

func (m *ABF) GetAttachedInterfaces() []*ABF_AttachedInterface {
	if m != nil {
		return m.AttachedInterfaces
	}
	return nil
}

func (m *ABF) GetForwardingPaths() []*ABF_ForwardingPath {
	if m != nil {
		return m.ForwardingPaths
	}
	return nil
}

// List of interfaces attached to the ABF
type ABF_AttachedInterface struct {
	InputInterface       string   `protobuf:"bytes,1,opt,name=input_interface,json=inputInterface,proto3" json:"input_interface,omitempty"`
	Priority             uint32   `protobuf:"varint,2,opt,name=priority,proto3" json:"priority,omitempty"`
	IsIpv6               bool     `protobuf:"varint,3,opt,name=is_ipv6,json=isIpv6,proto3" json:"is_ipv6,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ABF_AttachedInterface) Reset()         { *m = ABF_AttachedInterface{} }
func (m *ABF_AttachedInterface) String() string { return proto.CompactTextString(m) }
func (*ABF_AttachedInterface) ProtoMessage()    {}
func (*ABF_AttachedInterface) Descriptor() ([]byte, []int) {
	return fileDescriptor_50fc84daf012435f, []int{0, 0}
}

func (m *ABF_AttachedInterface) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ABF_AttachedInterface.Unmarshal(m, b)
}
func (m *ABF_AttachedInterface) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ABF_AttachedInterface.Marshal(b, m, deterministic)
}
func (m *ABF_AttachedInterface) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ABF_AttachedInterface.Merge(m, src)
}
func (m *ABF_AttachedInterface) XXX_Size() int {
	return xxx_messageInfo_ABF_AttachedInterface.Size(m)
}
func (m *ABF_AttachedInterface) XXX_DiscardUnknown() {
	xxx_messageInfo_ABF_AttachedInterface.DiscardUnknown(m)
}

var xxx_messageInfo_ABF_AttachedInterface proto.InternalMessageInfo

func (m *ABF_AttachedInterface) GetInputInterface() string {
	if m != nil {
		return m.InputInterface
	}
	return ""
}

func (m *ABF_AttachedInterface) GetPriority() uint32 {
	if m != nil {
		return m.Priority
	}
	return 0
}

func (m *ABF_AttachedInterface) GetIsIpv6() bool {
	if m != nil {
		return m.IsIpv6
	}
	return false
}

// List of forwarding paths added to the ABF policy (via)
type ABF_ForwardingPath struct {
	NextHopIp            string   `protobuf:"bytes,1,opt,name=next_hop_ip,json=nextHopIp,proto3" json:"next_hop_ip,omitempty"`
	InterfaceName        string   `protobuf:"bytes,2,opt,name=interface_name,json=interfaceName,proto3" json:"interface_name,omitempty"`
	Weight               uint32   `protobuf:"varint,3,opt,name=weight,proto3" json:"weight,omitempty"`
	Preference           uint32   `protobuf:"varint,4,opt,name=preference,proto3" json:"preference,omitempty"`
	Dvr                  bool     `protobuf:"varint,5,opt,name=dvr,proto3" json:"dvr,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ABF_ForwardingPath) Reset()         { *m = ABF_ForwardingPath{} }
func (m *ABF_ForwardingPath) String() string { return proto.CompactTextString(m) }
func (*ABF_ForwardingPath) ProtoMessage()    {}
func (*ABF_ForwardingPath) Descriptor() ([]byte, []int) {
	return fileDescriptor_50fc84daf012435f, []int{0, 1}
}

func (m *ABF_ForwardingPath) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ABF_ForwardingPath.Unmarshal(m, b)
}
func (m *ABF_ForwardingPath) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ABF_ForwardingPath.Marshal(b, m, deterministic)
}
func (m *ABF_ForwardingPath) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ABF_ForwardingPath.Merge(m, src)
}
func (m *ABF_ForwardingPath) XXX_Size() int {
	return xxx_messageInfo_ABF_ForwardingPath.Size(m)
}
func (m *ABF_ForwardingPath) XXX_DiscardUnknown() {
	xxx_messageInfo_ABF_ForwardingPath.DiscardUnknown(m)
}

var xxx_messageInfo_ABF_ForwardingPath proto.InternalMessageInfo

func (m *ABF_ForwardingPath) GetNextHopIp() string {
	if m != nil {
		return m.NextHopIp
	}
	return ""
}

func (m *ABF_ForwardingPath) GetInterfaceName() string {
	if m != nil {
		return m.InterfaceName
	}
	return ""
}

func (m *ABF_ForwardingPath) GetWeight() uint32 {
	if m != nil {
		return m.Weight
	}
	return 0
}

func (m *ABF_ForwardingPath) GetPreference() uint32 {
	if m != nil {
		return m.Preference
	}
	return 0
}

func (m *ABF_ForwardingPath) GetDvr() bool {
	if m != nil {
		return m.Dvr
	}
	return false
}

func init() {
	proto.RegisterType((*ABF)(nil), "vpp.abf.ABF")
	proto.RegisterType((*ABF_AttachedInterface)(nil), "vpp.abf.ABF.AttachedInterface")
	proto.RegisterType((*ABF_ForwardingPath)(nil), "vpp.abf.ABF.ForwardingPath")
}

func init() { proto.RegisterFile("models/vpp/abf/abf.proto", fileDescriptor_50fc84daf012435f) }

var fileDescriptor_50fc84daf012435f = []byte{
	// 385 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x52, 0x5d, 0x8b, 0xd3, 0x40,
	0x14, 0x25, 0x66, 0xb7, 0x1f, 0x77, 0x69, 0x77, 0x1d, 0x45, 0xc7, 0x0a, 0xa5, 0x08, 0x62, 0x5f,
	0x4c, 0x40, 0xa1, 0x08, 0x3e, 0xb5, 0x0f, 0xc5, 0xbe, 0xa8, 0xcc, 0xa3, 0x2f, 0xe1, 0x26, 0x99,
	0x24, 0x03, 0xed, 0xcc, 0x75, 0x32, 0x9b, 0x5d, 0x7f, 0x8e, 0x7f, 0xce, 0xdf, 0x21, 0x19, 0x63,
	0xb6, 0xd1, 0x87, 0x40, 0xce, 0x99, 0xc3, 0x3d, 0xe7, 0x5e, 0x0e, 0xf0, 0x93, 0xc9, 0xe5, 0xb1,
	0x8e, 0x1b, 0xa2, 0x18, 0xd3, 0xa2, 0xfd, 0x22, 0xb2, 0xc6, 0x19, 0x36, 0x6e, 0x88, 0x22, 0x4c,
	0x8b, 0x57, 0xbf, 0x42, 0x08, 0xb7, 0xbb, 0x3d, 0x7b, 0x0a, 0x97, 0x4a, 0xe7, 0xf2, 0x9e, 0x07,
	0xab, 0x60, 0x3d, 0x13, 0x7f, 0x00, 0x7b, 0x01, 0x13, 0xcc, 0x8e, 0x89, 0xc6, 0x93, 0xe4, 0x8f,
	0x56, 0xc1, 0x7a, 0x2a, 0xc6, 0x98, 0x1d, 0x3f, 0xe3, 0x49, 0xb2, 0x2f, 0xf0, 0x04, 0x9d, 0xc3,
	0xac, 0x92, 0x79, 0xa2, 0xb4, 0x93, 0xb6, 0xc0, 0x4c, 0xd6, 0x3c, 0x5c, 0x85, 0xeb, 0xab, 0x77,
	0xcb, 0xa8, 0x9b, 0x1f, 0x6d, 0x77, 0xfb, 0x68, 0xdb, 0xe9, 0x0e, 0x7f, 0x65, 0x82, 0xe1, 0xbf,
	0x54, 0xcd, 0xf6, 0x70, 0x53, 0x18, 0x7b, 0x87, 0x36, 0x57, 0xba, 0x4c, 0x08, 0x5d, 0x55, 0xf3,
	0x0b, 0x3f, 0xed, 0xe5, 0x60, 0xda, 0xbe, 0x17, 0x7d, 0x45, 0x57, 0x89, 0xeb, 0x62, 0x80, 0xeb,
	0xc5, 0x77, 0x78, 0xfc, 0x9f, 0x21, 0x7b, 0x03, 0xd7, 0x4a, 0xd3, 0xad, 0x7b, 0x88, 0xea, 0x17,
	0x9d, 0x8a, 0xb9, 0xa7, 0x1f, 0x84, 0x0b, 0x98, 0x90, 0x55, 0xc6, 0x2a, 0xf7, 0xc3, 0x6f, 0x3c,
	0x13, 0x3d, 0x66, 0xcf, 0x61, 0xac, 0xea, 0x44, 0x51, 0xb3, 0xe1, 0xe1, 0x2a, 0x58, 0x4f, 0xc4,
	0x48, 0xd5, 0x07, 0x6a, 0x36, 0x8b, 0x9f, 0x01, 0xcc, 0x87, 0xb1, 0xd8, 0x12, 0xae, 0xb4, 0xbc,
	0x77, 0x49, 0x65, 0x28, 0x51, 0xd4, 0x99, 0x4d, 0x5b, 0xea, 0x93, 0xa1, 0x03, 0xb1, 0xd7, 0x30,
	0xef, 0xa3, 0x9c, 0xdf, 0x77, 0xd6, 0xb3, 0xfe, 0xca, 0xcf, 0x60, 0x74, 0x27, 0x55, 0x59, 0x39,
	0xef, 0x38, 0x13, 0x1d, 0x62, 0x4b, 0x00, 0xb2, 0xb2, 0x90, 0x56, 0xea, 0x4c, 0xf2, 0x0b, 0xff,
	0x76, 0xc6, 0xb0, 0x1b, 0x08, 0xf3, 0xc6, 0xf2, 0x4b, 0x1f, 0xb3, 0xfd, 0xdd, 0x7d, 0xf8, 0xb6,
	0x29, 0x95, 0xab, 0x6e, 0xd3, 0x28, 0x33, 0xa7, 0xf8, 0xa8, 0x4a, 0x74, 0xa6, 0x2d, 0xc6, 0x5b,
	0x2c, 0xa5, 0x76, 0x31, 0x92, 0x8a, 0x87, 0x6d, 0xf9, 0xd8, 0x10, 0x25, 0x98, 0x16, 0xe9, 0xc8,
	0x57, 0xe6, 0xfd, 0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0x45, 0x72, 0xed, 0x59, 0x4e, 0x02, 0x00,
	0x00,
}
