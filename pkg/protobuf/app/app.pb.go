// Code generated by protoc-gen-go.
// source: pkg/protobuf/app/app.proto
// DO NOT EDIT!

/*
Package app is a generated protocol buffer package.

It is generated from these files:
	pkg/protobuf/app/app.proto

It has these top-level messages:
	CreateRequest
	Empty
*/
package app

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type CreateRequest struct {
	Name        string                   `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	Team        string                   `protobuf:"bytes,2,opt,name=team" json:"team,omitempty"`
	ProcessType string                   `protobuf:"bytes,3,opt,name=process_type,json=processType" json:"process_type,omitempty"`
	Limits      *CreateRequest_Limits    `protobuf:"bytes,4,opt,name=limits" json:"limits,omitempty"`
	AutoScale   *CreateRequest_AutoScale `protobuf:"bytes,5,opt,name=auto_scale,json=autoScale" json:"auto_scale,omitempty"`
}

func (m *CreateRequest) Reset()                    { *m = CreateRequest{} }
func (m *CreateRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateRequest) ProtoMessage()               {}
func (*CreateRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *CreateRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *CreateRequest) GetTeam() string {
	if m != nil {
		return m.Team
	}
	return ""
}

func (m *CreateRequest) GetProcessType() string {
	if m != nil {
		return m.ProcessType
	}
	return ""
}

func (m *CreateRequest) GetLimits() *CreateRequest_Limits {
	if m != nil {
		return m.Limits
	}
	return nil
}

func (m *CreateRequest) GetAutoScale() *CreateRequest_AutoScale {
	if m != nil {
		return m.AutoScale
	}
	return nil
}

type CreateRequest_Limits struct {
	Default        []*CreateRequest_Limits_LimitRangeQuantity `protobuf:"bytes,1,rep,name=default" json:"default,omitempty"`
	DefaultRequest []*CreateRequest_Limits_LimitRangeQuantity `protobuf:"bytes,2,rep,name=default_request,json=defaultRequest" json:"default_request,omitempty"`
}

func (m *CreateRequest_Limits) Reset()                    { *m = CreateRequest_Limits{} }
func (m *CreateRequest_Limits) String() string            { return proto.CompactTextString(m) }
func (*CreateRequest_Limits) ProtoMessage()               {}
func (*CreateRequest_Limits) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

func (m *CreateRequest_Limits) GetDefault() []*CreateRequest_Limits_LimitRangeQuantity {
	if m != nil {
		return m.Default
	}
	return nil
}

func (m *CreateRequest_Limits) GetDefaultRequest() []*CreateRequest_Limits_LimitRangeQuantity {
	if m != nil {
		return m.DefaultRequest
	}
	return nil
}

type CreateRequest_Limits_LimitRangeQuantity struct {
	Quantity string `protobuf:"bytes,1,opt,name=quantity" json:"quantity,omitempty"`
	Resource string `protobuf:"bytes,2,opt,name=resource" json:"resource,omitempty"`
}

func (m *CreateRequest_Limits_LimitRangeQuantity) Reset() {
	*m = CreateRequest_Limits_LimitRangeQuantity{}
}
func (m *CreateRequest_Limits_LimitRangeQuantity) String() string { return proto.CompactTextString(m) }
func (*CreateRequest_Limits_LimitRangeQuantity) ProtoMessage()    {}
func (*CreateRequest_Limits_LimitRangeQuantity) Descriptor() ([]byte, []int) {
	return fileDescriptor0, []int{0, 0, 0}
}

func (m *CreateRequest_Limits_LimitRangeQuantity) GetQuantity() string {
	if m != nil {
		return m.Quantity
	}
	return ""
}

func (m *CreateRequest_Limits_LimitRangeQuantity) GetResource() string {
	if m != nil {
		return m.Resource
	}
	return ""
}

type CreateRequest_AutoScale struct {
	CpuTargetUtilization int32 `protobuf:"varint,1,opt,name=cpu_target_utilization,json=cpuTargetUtilization" json:"cpu_target_utilization,omitempty"`
	Max                  int32 `protobuf:"varint,2,opt,name=max" json:"max,omitempty"`
	Min                  int32 `protobuf:"varint,3,opt,name=min" json:"min,omitempty"`
}

func (m *CreateRequest_AutoScale) Reset()                    { *m = CreateRequest_AutoScale{} }
func (m *CreateRequest_AutoScale) String() string            { return proto.CompactTextString(m) }
func (*CreateRequest_AutoScale) ProtoMessage()               {}
func (*CreateRequest_AutoScale) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 1} }

func (m *CreateRequest_AutoScale) GetCpuTargetUtilization() int32 {
	if m != nil {
		return m.CpuTargetUtilization
	}
	return 0
}

func (m *CreateRequest_AutoScale) GetMax() int32 {
	if m != nil {
		return m.Max
	}
	return 0
}

func (m *CreateRequest_AutoScale) GetMin() int32 {
	if m != nil {
		return m.Min
	}
	return 0
}

type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func init() {
	proto.RegisterType((*CreateRequest)(nil), "app.CreateRequest")
	proto.RegisterType((*CreateRequest_Limits)(nil), "app.CreateRequest.Limits")
	proto.RegisterType((*CreateRequest_Limits_LimitRangeQuantity)(nil), "app.CreateRequest.Limits.LimitRangeQuantity")
	proto.RegisterType((*CreateRequest_AutoScale)(nil), "app.CreateRequest.AutoScale")
	proto.RegisterType((*Empty)(nil), "app.Empty")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for App service

type AppClient interface {
	Create(ctx context.Context, in *CreateRequest, opts ...grpc.CallOption) (*Empty, error)
}

type appClient struct {
	cc *grpc.ClientConn
}

func NewAppClient(cc *grpc.ClientConn) AppClient {
	return &appClient{cc}
}

func (c *appClient) Create(ctx context.Context, in *CreateRequest, opts ...grpc.CallOption) (*Empty, error) {
	out := new(Empty)
	err := grpc.Invoke(ctx, "/app.App/Create", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for App service

type AppServer interface {
	Create(context.Context, *CreateRequest) (*Empty, error)
}

func RegisterAppServer(s *grpc.Server, srv AppServer) {
	s.RegisterService(&_App_serviceDesc, srv)
}

func _App_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AppServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/app.App/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AppServer).Create(ctx, req.(*CreateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _App_serviceDesc = grpc.ServiceDesc{
	ServiceName: "app.App",
	HandlerType: (*AppServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _App_Create_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/protobuf/app/app.proto",
}

func init() { proto.RegisterFile("pkg/protobuf/app/app.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 362 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x9c, 0x52, 0x3b, 0x6f, 0xe2, 0x40,
	0x10, 0x96, 0x31, 0x36, 0xc7, 0x70, 0x2f, 0xad, 0x4e, 0x27, 0x9f, 0x75, 0x05, 0x47, 0xe5, 0xe2,
	0x04, 0x3a, 0x2e, 0x5d, 0x2a, 0x14, 0x25, 0x15, 0x4d, 0x36, 0x50, 0x5b, 0x8b, 0x33, 0xa0, 0x55,
	0xfc, 0x58, 0xec, 0x59, 0x29, 0xce, 0x3f, 0xca, 0x0f, 0x4c, 0x1f, 0xed, 0x7a, 0x21, 0x8a, 0x48,
	0x9a, 0x14, 0x96, 0xbf, 0xfd, 0x5e, 0x23, 0xcf, 0x1a, 0x62, 0x75, 0xb7, 0x9b, 0xa9, 0xba, 0xa2,
	0x6a, 0xa3, 0xb7, 0x33, 0xa1, 0x94, 0x79, 0xa6, 0x96, 0x60, 0xbe, 0x50, 0x6a, 0xf2, 0xd8, 0x87,
	0x2f, 0x17, 0x35, 0x0a, 0x42, 0x8e, 0x7b, 0x8d, 0x0d, 0x31, 0x06, 0xfd, 0x52, 0x14, 0x18, 0x79,
	0x63, 0x2f, 0x19, 0x72, 0x8b, 0x0d, 0x47, 0x28, 0x8a, 0xa8, 0xd7, 0x71, 0x06, 0xb3, 0x3f, 0xf0,
	0x59, 0xd5, 0x55, 0x86, 0x4d, 0x93, 0x52, 0xab, 0x30, 0xf2, 0xad, 0x36, 0x72, 0xdc, 0xaa, 0x55,
	0xc8, 0xfe, 0x41, 0x98, 0xcb, 0x42, 0x52, 0x13, 0xf5, 0xc7, 0x5e, 0x32, 0x9a, 0xff, 0x9a, 0x9a,
	0xe9, 0xaf, 0xc6, 0x4d, 0x97, 0xd6, 0xc0, 0x9d, 0x91, 0x9d, 0x03, 0x08, 0x4d, 0x55, 0xda, 0x64,
	0x22, 0xc7, 0x28, 0xb0, 0xb1, 0xdf, 0x6f, 0xc4, 0x16, 0x9a, 0xaa, 0x1b, 0xe3, 0xe1, 0x43, 0x71,
	0x80, 0xf1, 0x93, 0x07, 0x61, 0xd7, 0xc7, 0xae, 0x60, 0x70, 0x8b, 0x5b, 0xa1, 0x73, 0x8a, 0xbc,
	0xb1, 0x9f, 0x8c, 0xe6, 0x7f, 0xdf, 0x9d, 0xdd, 0xbd, 0xb8, 0x28, 0x77, 0x78, 0xad, 0x45, 0x49,
	0x92, 0x5a, 0x7e, 0x08, 0xb3, 0x35, 0x7c, 0x73, 0x30, 0xad, 0xbb, 0x54, 0xd4, 0xfb, 0x40, 0xdf,
	0x57, 0x57, 0xe2, 0x9c, 0xf1, 0x12, 0xd8, 0xa9, 0x8b, 0xc5, 0xf0, 0x69, 0xef, 0xb0, 0x5b, 0xff,
	0xf1, 0x6c, 0xb4, 0x1a, 0x9b, 0x4a, 0xd7, 0x19, 0xba, 0x6b, 0x38, 0x9e, 0x63, 0x84, 0xe1, 0x71,
	0x1f, 0xec, 0x0c, 0x7e, 0x66, 0x4a, 0xa7, 0x24, 0xea, 0x1d, 0x52, 0xaa, 0x49, 0xe6, 0xf2, 0x41,
	0x90, 0xac, 0x4a, 0x5b, 0x19, 0xf0, 0x1f, 0x99, 0xd2, 0x2b, 0x2b, 0xae, 0x5f, 0x34, 0xf6, 0x1d,
	0xfc, 0x42, 0xdc, 0xdb, 0xe6, 0x80, 0x1b, 0x68, 0x19, 0x59, 0xda, 0x6b, 0x35, 0x8c, 0x2c, 0x27,
	0x03, 0x08, 0x2e, 0x0b, 0x45, 0xed, 0x7c, 0x06, 0xfe, 0x42, 0x29, 0x96, 0x40, 0xd8, 0x7d, 0x3f,
	0x63, 0xa7, 0xcb, 0x88, 0xc1, 0x72, 0x36, 0xb0, 0x09, 0xed, 0x1f, 0xf7, 0xff, 0x39, 0x00, 0x00,
	0xff, 0xff, 0xc1, 0x39, 0x7d, 0xf6, 0x8f, 0x02, 0x00, 0x00,
}
