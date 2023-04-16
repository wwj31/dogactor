// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: actor_msg.proto

package actor_msg

import (
	context "context"
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

func (a *ActorMessage) Reset()         { *a = ActorMessage{} }
func (a *ActorMessage) String() string { return proto.CompactTextString(a) }
func (*ActorMessage) ProtoMessage()    {}
func (*ActorMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_b5bf8dd4e82efef3, []int{0}
}
func (a *ActorMessage) XXX_Unmarshal(b []byte) error {
	return a.Unmarshal(b)
}
func (a *ActorMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ActorMessage.Marshal(b, a, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := a.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (a *ActorMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ActorMessage.Merge(a, src)
}
func (a *ActorMessage) XXX_Size() int {
	return a.Size()
}
func (a *ActorMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_ActorMessage.DiscardUnknown(a)
}

var xxx_messageInfo_ActorMessage proto.InternalMessageInfo

func (a *ActorMessage) GetSourceId() string {
	if a != nil {
		return a.SourceId
	}
	return ""
}

func (a *ActorMessage) GetTargetId() string {
	if a != nil {
		return a.TargetId
	}
	return ""
}

func (a *ActorMessage) GetRequestId() string {
	if a != nil {
		return a.RequestId
	}
	return ""
}

func (a *ActorMessage) GetMsgName() string {
	if a != nil {
		return a.MsgName
	}
	return ""
}

func (a *ActorMessage) GetData() []byte {
	if a != nil {
		return a.Data
	}
	return nil
}

func (a *ActorMessage) GetMapCarrier() map[string]string {
	if a != nil {
		return a.MapCarrier
	}
	return nil
}

type RequestDeadLetter struct {
	Err string `protobuf:"bytes,1,opt,name=Err,proto3" json:"Err,omitempty"`
}

func (m *RequestDeadLetter) Reset()         { *m = RequestDeadLetter{} }
func (m *RequestDeadLetter) String() string { return proto.CompactTextString(m) }
func (*RequestDeadLetter) ProtoMessage()    {}
func (*RequestDeadLetter) Descriptor() ([]byte, []int) {
	return fileDescriptor_b5bf8dd4e82efef3, []int{1}
}
func (m *RequestDeadLetter) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *RequestDeadLetter) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_RequestDeadLetter.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *RequestDeadLetter) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RequestDeadLetter.Merge(m, src)
}
func (m *RequestDeadLetter) XXX_Size() int {
	return m.Size()
}
func (m *RequestDeadLetter) XXX_DiscardUnknown() {
	xxx_messageInfo_RequestDeadLetter.DiscardUnknown(m)
}

var xxx_messageInfo_RequestDeadLetter proto.InternalMessageInfo

func (m *RequestDeadLetter) GetErr() string {
	if m != nil {
		return m.Err
	}
	return ""
}

func init() {
	proto.RegisterType((*ActorMessage)(nil), "actor_msg.ActorMessage")
	proto.RegisterMapType((map[string]string)(nil), "actor_msg.ActorMessage.MapCarrierEntry")
	proto.RegisterType((*RequestDeadLetter)(nil), "actor_msg.RequestDeadLetter")
}

func init() { proto.RegisterFile("actor_msg.proto", fileDescriptor_b5bf8dd4e82efef3) }

var fileDescriptor_b5bf8dd4e82efef3 = []byte{
	// 307 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x91, 0xcf, 0x4a, 0xc3, 0x40,
	0x10, 0x87, 0xbb, 0xfd, 0x6b, 0xc7, 0x42, 0x75, 0x11, 0x5c, 0x8a, 0x2c, 0xa5, 0x20, 0xe6, 0x54,
	0xa4, 0x5e, 0x44, 0x10, 0x51, 0x5b, 0xa4, 0x60, 0x15, 0x52, 0xef, 0xb2, 0x26, 0x43, 0x08, 0x5a,
	0x53, 0x27, 0x9b, 0x40, 0xdf, 0xc2, 0xb3, 0x4f, 0xe4, 0xb1, 0x47, 0x8f, 0x92, 0xbc, 0x88, 0x24,
	0x4d, 0x9a, 0x22, 0xf4, 0x36, 0xdf, 0x7c, 0xbb, 0xc3, 0xfc, 0x76, 0xa1, 0xad, 0x2c, 0xed, 0xd1,
	0xf3, 0xcc, 0x77, 0xfa, 0x73, 0xf2, 0xb4, 0xc7, 0x9b, 0xeb, 0x46, 0xef, 0xab, 0x0c, 0xad, 0xeb,
	0x84, 0x26, 0xe8, 0xfb, 0xca, 0x41, 0xde, 0x81, 0x9d, 0xa9, 0x17, 0x90, 0x85, 0x63, 0x5b, 0xb0,
	0x2e, 0x33, 0x9a, 0xe6, 0x9a, 0x13, 0xf7, 0xa4, 0xc8, 0x41, 0x3d, 0xb6, 0x45, 0x79, 0xe5, 0x72,
	0xe6, 0x47, 0xd0, 0x34, 0xf1, 0x23, 0x40, 0x3f, 0x91, 0x95, 0x54, 0x16, 0x0d, 0x2e, 0xa0, 0x31,
	0xf1, 0x9d, 0x07, 0x35, 0x43, 0x51, 0x4d, 0x5d, 0x8e, 0x9c, 0x43, 0x75, 0xa8, 0xb4, 0x12, 0xb5,
	0x2e, 0x33, 0x5a, 0x66, 0x5a, 0xf3, 0x3b, 0x80, 0x89, 0x9a, 0xdf, 0x2a, 0x22, 0x17, 0x49, 0xd4,
	0xbb, 0x15, 0x63, 0x77, 0x70, 0xd2, 0x2f, 0x52, 0x6c, 0x2e, 0xdc, 0x2f, 0x4e, 0x8e, 0xde, 0x35,
	0x2d, 0xcc, 0x8d, 0xab, 0x9d, 0x4b, 0x68, 0xff, 0xd3, 0x7c, 0x0f, 0x2a, 0xaf, 0xb8, 0xc8, 0xa2,
	0x25, 0x25, 0x3f, 0x80, 0x5a, 0xa8, 0xde, 0x02, 0xcc, 0x22, 0xad, 0xe0, 0xa2, 0x7c, 0xce, 0x7a,
	0xc7, 0xb0, 0x9f, 0x45, 0x18, 0xa2, 0xb2, 0xef, 0x51, 0x6b, 0xa4, 0x64, 0xc0, 0x88, 0x28, 0x1f,
	0x30, 0x22, 0x1a, 0x3c, 0x66, 0x4f, 0x38, 0x45, 0x0a, 0x5d, 0x0b, 0xf9, 0x15, 0x34, 0x4c, 0xb4,
	0xd0, 0x0d, 0x91, 0x1f, 0x6e, 0xd9, 0xba, 0xb3, 0x4d, 0x18, 0xec, 0x94, 0xdd, 0x88, 0xef, 0x48,
	0xb2, 0x65, 0x24, 0xd9, 0x6f, 0x24, 0xd9, 0x67, 0x2c, 0x4b, 0xcb, 0x58, 0x96, 0x7e, 0x62, 0x59,
	0x7a, 0xa9, 0xa7, 0x1f, 0x78, 0xf6, 0x17, 0x00, 0x00, 0xff, 0xff, 0xe4, 0x1b, 0x2b, 0x81, 0xd3,
	0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ActorServiceClient is the client API for ActorService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ActorServiceClient interface {
	Receive(ctx context.Context, opts ...grpc.CallOption) (ActorService_ReceiveClient, error)
}

type actorServiceClient struct {
	cc *grpc.ClientConn
}

func NewActorServiceClient(cc *grpc.ClientConn) ActorServiceClient {
	return &actorServiceClient{cc}
}

func (c *actorServiceClient) Receive(ctx context.Context, opts ...grpc.CallOption) (ActorService_ReceiveClient, error) {
	stream, err := c.cc.NewStream(ctx, &_ActorService_serviceDesc.Streams[0], "/actor_msg.ActorService/Receive", opts...)
	if err != nil {
		return nil, err
	}
	x := &actorServiceReceiveClient{stream}
	return x, nil
}

type ActorService_ReceiveClient interface {
	Send(*ActorMessage) error
	Recv() (*ActorMessage, error)
	grpc.ClientStream
}

type actorServiceReceiveClient struct {
	grpc.ClientStream
}

func (x *actorServiceReceiveClient) Send(m *ActorMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *actorServiceReceiveClient) Recv() (*ActorMessage, error) {
	m := new(ActorMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ActorServiceServer is the server API for ActorService service.
type ActorServiceServer interface {
	Receive(ActorService_ReceiveServer) error
}

// UnimplementedActorServiceServer can be embedded to have forward compatible implementations.
type UnimplementedActorServiceServer struct {
}

func (*UnimplementedActorServiceServer) Receive(srv ActorService_ReceiveServer) error {
	return status.Errorf(codes.Unimplemented, "method Receive not implemented")
}

func RegisterActorServiceServer(s *grpc.Server, srv ActorServiceServer) {
	s.RegisterService(&_ActorService_serviceDesc, srv)
}

func _ActorService_Receive_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ActorServiceServer).Receive(&actorServiceReceiveServer{stream})
}

type ActorService_ReceiveServer interface {
	Send(*ActorMessage) error
	Recv() (*ActorMessage, error)
	grpc.ServerStream
}

type actorServiceReceiveServer struct {
	grpc.ServerStream
}

func (x *actorServiceReceiveServer) Send(m *ActorMessage) error {
	return x.ServerStream.SendMsg(m)
}

func (x *actorServiceReceiveServer) Recv() (*ActorMessage, error) {
	m := new(ActorMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _ActorService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "actor_msg.ActorService",
	HandlerType: (*ActorServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Receive",
			Handler:       _ActorService_Receive_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "actor_msg.proto",
}

func (a *ActorMessage) Marshal() (dAtA []byte, err error) {
	size := a.Size()
	dAtA = make([]byte, size)
	n, err := a.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (a *ActorMessage) MarshalTo(dAtA []byte) (int, error) {
	size := a.Size()
	return a.MarshalToSizedBuffer(dAtA[:size])
}

func (a *ActorMessage) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(a.MapCarrier) > 0 {
		for k := range a.MapCarrier {
			v := a.MapCarrier[k]
			baseI := i
			i -= len(v)
			copy(dAtA[i:], v)
			i = encodeVarintActorMsg(dAtA, i, uint64(len(v)))
			i--
			dAtA[i] = 0x12
			i -= len(k)
			copy(dAtA[i:], k)
			i = encodeVarintActorMsg(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = encodeVarintActorMsg(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0x32
		}
	}
	if len(a.Data) > 0 {
		i -= len(a.Data)
		copy(dAtA[i:], a.Data)
		i = encodeVarintActorMsg(dAtA, i, uint64(len(a.Data)))
		i--
		dAtA[i] = 0x2a
	}
	if len(a.MsgName) > 0 {
		i -= len(a.MsgName)
		copy(dAtA[i:], a.MsgName)
		i = encodeVarintActorMsg(dAtA, i, uint64(len(a.MsgName)))
		i--
		dAtA[i] = 0x22
	}
	if len(a.RequestId) > 0 {
		i -= len(a.RequestId)
		copy(dAtA[i:], a.RequestId)
		i = encodeVarintActorMsg(dAtA, i, uint64(len(a.RequestId)))
		i--
		dAtA[i] = 0x1a
	}
	if len(a.TargetId) > 0 {
		i -= len(a.TargetId)
		copy(dAtA[i:], a.TargetId)
		i = encodeVarintActorMsg(dAtA, i, uint64(len(a.TargetId)))
		i--
		dAtA[i] = 0x12
	}
	if len(a.SourceId) > 0 {
		i -= len(a.SourceId)
		copy(dAtA[i:], a.SourceId)
		i = encodeVarintActorMsg(dAtA, i, uint64(len(a.SourceId)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *RequestDeadLetter) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *RequestDeadLetter) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *RequestDeadLetter) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Err) > 0 {
		i -= len(m.Err)
		copy(dAtA[i:], m.Err)
		i = encodeVarintActorMsg(dAtA, i, uint64(len(m.Err)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintActorMsg(dAtA []byte, offset int, v uint64) int {
	offset -= sovActorMsg(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (a *ActorMessage) Size() (n int) {
	if a == nil {
		return 0
	}
	var l int
	_ = l
	l = len(a.SourceId)
	if l > 0 {
		n += 1 + l + sovActorMsg(uint64(l))
	}
	l = len(a.TargetId)
	if l > 0 {
		n += 1 + l + sovActorMsg(uint64(l))
	}
	l = len(a.RequestId)
	if l > 0 {
		n += 1 + l + sovActorMsg(uint64(l))
	}
	l = len(a.MsgName)
	if l > 0 {
		n += 1 + l + sovActorMsg(uint64(l))
	}
	l = len(a.Data)
	if l > 0 {
		n += 1 + l + sovActorMsg(uint64(l))
	}
	if len(a.MapCarrier) > 0 {
		for k, v := range a.MapCarrier {
			_ = k
			_ = v
			mapEntrySize := 1 + len(k) + sovActorMsg(uint64(len(k))) + 1 + len(v) + sovActorMsg(uint64(len(v)))
			n += mapEntrySize + 1 + sovActorMsg(uint64(mapEntrySize))
		}
	}
	return n
}

func (m *RequestDeadLetter) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Err)
	if l > 0 {
		n += 1 + l + sovActorMsg(uint64(l))
	}
	return n
}

func sovActorMsg(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozActorMsg(x uint64) (n int) {
	return sovActorMsg(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (a *ActorMessage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowActorMsg
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ActorMessage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ActorMessage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SourceId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowActorMsg
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthActorMsg
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthActorMsg
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			a.SourceId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TargetId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowActorMsg
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthActorMsg
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthActorMsg
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			a.TargetId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RequestId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowActorMsg
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthActorMsg
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthActorMsg
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			a.RequestId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MsgName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowActorMsg
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthActorMsg
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthActorMsg
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			a.MsgName = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowActorMsg
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthActorMsg
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthActorMsg
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			a.Data = append(a.Data[:0], dAtA[iNdEx:postIndex]...)
			if a.Data == nil {
				a.Data = []byte{}
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MapCarrier", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowActorMsg
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthActorMsg
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthActorMsg
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if a.MapCarrier == nil {
				a.MapCarrier = make(map[string]string)
			}
			var mapkey string
			var mapvalue string
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowActorMsg
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowActorMsg
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthActorMsg
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return ErrInvalidLengthActorMsg
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var stringLenmapvalue uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowActorMsg
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapvalue |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapvalue := int(stringLenmapvalue)
					if intStringLenmapvalue < 0 {
						return ErrInvalidLengthActorMsg
					}
					postStringIndexmapvalue := iNdEx + intStringLenmapvalue
					if postStringIndexmapvalue < 0 {
						return ErrInvalidLengthActorMsg
					}
					if postStringIndexmapvalue > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = string(dAtA[iNdEx:postStringIndexmapvalue])
					iNdEx = postStringIndexmapvalue
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipActorMsg(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if (skippy < 0) || (iNdEx+skippy) < 0 {
						return ErrInvalidLengthActorMsg
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			a.MapCarrier[mapkey] = mapvalue
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipActorMsg(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthActorMsg
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *RequestDeadLetter) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowActorMsg
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: RequestDeadLetter: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: RequestDeadLetter: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Err", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowActorMsg
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthActorMsg
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthActorMsg
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Err = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipActorMsg(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthActorMsg
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipActorMsg(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowActorMsg
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowActorMsg
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowActorMsg
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthActorMsg
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupActorMsg
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthActorMsg
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthActorMsg        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowActorMsg          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupActorMsg = fmt.Errorf("proto: unexpected end of group")
)
