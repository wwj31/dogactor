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

func (m *ActorMessage) Reset()         { *m = ActorMessage{} }
func (m *ActorMessage) String() string { return proto.CompactTextString(m) }
func (*ActorMessage) ProtoMessage()    {}
func (*ActorMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_b5bf8dd4e82efef3, []int{0}
}
func (m *ActorMessage) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ActorMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ActorMessage.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ActorMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ActorMessage.Merge(m, src)
}
func (m *ActorMessage) XXX_Size() int {
	return m.Size()
}
func (m *ActorMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_ActorMessage.DiscardUnknown(m)
}

var xxx_messageInfo_ActorMessage proto.InternalMessageInfo

func (m *ActorMessage) GetSourceId() string {
	if m != nil {
		return m.SourceId
	}
	return ""
}

func (m *ActorMessage) GetTargetId() string {
	if m != nil {
		return m.TargetId
	}
	return ""
}

func (m *ActorMessage) GetRequestId() string {
	if m != nil {
		return m.RequestId
	}
	return ""
}

func (m *ActorMessage) GetMsgName() string {
	if m != nil {
		return m.MsgName
	}
	return ""
}

func (m *ActorMessage) GetData() []byte {
	if m != nil {
		return m.Data
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
	proto.RegisterType((*RequestDeadLetter)(nil), "actor_msg.RequestDeadLetter")
}

func init() { proto.RegisterFile("actor_msg.proto", fileDescriptor_b5bf8dd4e82efef3) }

var fileDescriptor_b5bf8dd4e82efef3 = []byte{
	// 253 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4f, 0x4c, 0x2e, 0xc9,
	0x2f, 0x8a, 0xcf, 0x2d, 0x4e, 0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x84, 0x0b, 0x28,
	0x4d, 0x62, 0xe4, 0xe2, 0x71, 0x04, 0xf1, 0x7c, 0x53, 0x8b, 0x8b, 0x13, 0xd3, 0x53, 0x85, 0xa4,
	0xb8, 0x38, 0x82, 0xf3, 0x4b, 0x8b, 0x92, 0x53, 0x3d, 0x53, 0x24, 0x18, 0x15, 0x18, 0x35, 0x38,
	0x83, 0xe0, 0x7c, 0x90, 0x5c, 0x48, 0x62, 0x51, 0x7a, 0x6a, 0x89, 0x67, 0x8a, 0x04, 0x13, 0x44,
	0x0e, 0xc6, 0x17, 0x92, 0xe1, 0xe2, 0x0c, 0x4a, 0x2d, 0x2c, 0x4d, 0x2d, 0x06, 0x49, 0x32, 0x83,
	0x25, 0x11, 0x02, 0x42, 0x12, 0x5c, 0xec, 0xbe, 0xc5, 0xe9, 0x7e, 0x89, 0xb9, 0xa9, 0x12, 0x2c,
	0x60, 0x39, 0x18, 0x57, 0x48, 0x88, 0x8b, 0xc5, 0x25, 0xb1, 0x24, 0x51, 0x82, 0x55, 0x81, 0x51,
	0x83, 0x27, 0x08, 0xcc, 0x56, 0x52, 0xe5, 0x12, 0x84, 0x6a, 0x75, 0x49, 0x4d, 0x4c, 0xf1, 0x49,
	0x2d, 0x29, 0x49, 0x2d, 0x12, 0x12, 0xe0, 0x62, 0x76, 0x2d, 0x2a, 0x82, 0xba, 0x09, 0xc4, 0x34,
	0xea, 0x86, 0xb9, 0x3d, 0x38, 0xb5, 0xa8, 0x2c, 0x33, 0x39, 0x55, 0xc8, 0x9e, 0x8b, 0x3d, 0x28,
	0x35, 0x39, 0x35, 0xb3, 0x2c, 0x55, 0x48, 0x5c, 0x0f, 0xe1, 0x69, 0x64, 0xff, 0x49, 0xe1, 0x92,
	0xd0, 0x60, 0x34, 0x60, 0x14, 0xb2, 0xe2, 0x62, 0x09, 0x49, 0x2d, 0x2e, 0x21, 0x47, 0xb7, 0x93,
	0xc4, 0x89, 0x47, 0x72, 0x8c, 0x17, 0x1e, 0xc9, 0x31, 0x3e, 0x78, 0x24, 0xc7, 0x38, 0xe1, 0xb1,
	0x1c, 0xc3, 0x85, 0xc7, 0x72, 0x0c, 0x37, 0x1e, 0xcb, 0x31, 0x24, 0xb1, 0x81, 0x43, 0xdd, 0x18,
	0x10, 0x00, 0x00, 0xff, 0xff, 0xbf, 0x88, 0xfd, 0xdf, 0x88, 0x01, 0x00, 0x00,
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
	Test(ctx context.Context, opts ...grpc.CallOption) (ActorService_TestClient, error)
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

func (c *actorServiceClient) Test(ctx context.Context, opts ...grpc.CallOption) (ActorService_TestClient, error) {
	stream, err := c.cc.NewStream(ctx, &_ActorService_serviceDesc.Streams[1], "/actor_msg.ActorService/Test", opts...)
	if err != nil {
		return nil, err
	}
	x := &actorServiceTestClient{stream}
	return x, nil
}

type ActorService_TestClient interface {
	Send(*ActorMessage) error
	CloseAndRecv() (*ActorMessage, error)
	grpc.ClientStream
}

type actorServiceTestClient struct {
	grpc.ClientStream
}

func (x *actorServiceTestClient) Send(m *ActorMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *actorServiceTestClient) CloseAndRecv() (*ActorMessage, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(ActorMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ActorServiceServer is the server API for ActorService service.
type ActorServiceServer interface {
	Receive(ActorService_ReceiveServer) error
	Test(ActorService_TestServer) error
}

// UnimplementedActorServiceServer can be embedded to have forward compatible implementations.
type UnimplementedActorServiceServer struct {
}

func (*UnimplementedActorServiceServer) Receive(srv ActorService_ReceiveServer) error {
	return status.Errorf(codes.Unimplemented, "method Receive not implemented")
}
func (*UnimplementedActorServiceServer) Test(srv ActorService_TestServer) error {
	return status.Errorf(codes.Unimplemented, "method Test not implemented")
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

func _ActorService_Test_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ActorServiceServer).Test(&actorServiceTestServer{stream})
}

type ActorService_TestServer interface {
	SendAndClose(*ActorMessage) error
	Recv() (*ActorMessage, error)
	grpc.ServerStream
}

type actorServiceTestServer struct {
	grpc.ServerStream
}

func (x *actorServiceTestServer) SendAndClose(m *ActorMessage) error {
	return x.ServerStream.SendMsg(m)
}

func (x *actorServiceTestServer) Recv() (*ActorMessage, error) {
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
		{
			StreamName:    "Test",
			Handler:       _ActorService_Test_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "actor_msg.proto",
}

func (m *ActorMessage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ActorMessage) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ActorMessage) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Data) > 0 {
		i -= len(m.Data)
		copy(dAtA[i:], m.Data)
		i = encodeVarintActorMsg(dAtA, i, uint64(len(m.Data)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.MsgName) > 0 {
		i -= len(m.MsgName)
		copy(dAtA[i:], m.MsgName)
		i = encodeVarintActorMsg(dAtA, i, uint64(len(m.MsgName)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.RequestId) > 0 {
		i -= len(m.RequestId)
		copy(dAtA[i:], m.RequestId)
		i = encodeVarintActorMsg(dAtA, i, uint64(len(m.RequestId)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.TargetId) > 0 {
		i -= len(m.TargetId)
		copy(dAtA[i:], m.TargetId)
		i = encodeVarintActorMsg(dAtA, i, uint64(len(m.TargetId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.SourceId) > 0 {
		i -= len(m.SourceId)
		copy(dAtA[i:], m.SourceId)
		i = encodeVarintActorMsg(dAtA, i, uint64(len(m.SourceId)))
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
func (m *ActorMessage) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.SourceId)
	if l > 0 {
		n += 1 + l + sovActorMsg(uint64(l))
	}
	l = len(m.TargetId)
	if l > 0 {
		n += 1 + l + sovActorMsg(uint64(l))
	}
	l = len(m.RequestId)
	if l > 0 {
		n += 1 + l + sovActorMsg(uint64(l))
	}
	l = len(m.MsgName)
	if l > 0 {
		n += 1 + l + sovActorMsg(uint64(l))
	}
	l = len(m.Data)
	if l > 0 {
		n += 1 + l + sovActorMsg(uint64(l))
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
func (m *ActorMessage) Unmarshal(dAtA []byte) error {
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
			m.SourceId = string(dAtA[iNdEx:postIndex])
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
			m.TargetId = string(dAtA[iNdEx:postIndex])
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
			m.RequestId = string(dAtA[iNdEx:postIndex])
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
			m.MsgName = string(dAtA[iNdEx:postIndex])
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
			m.Data = append(m.Data[:0], dAtA[iNdEx:postIndex]...)
			if m.Data == nil {
				m.Data = []byte{}
			}
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
