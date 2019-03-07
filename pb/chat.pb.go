// Code generated by protoc-gen-go. DO NOT EDIT.
// source: chat.proto

package pb

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
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

type Message struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Text                 string   `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}
func (*Message) Descriptor() ([]byte, []int) {
	return fileDescriptor_8c585a45e2093e54, []int{0}
}

func (m *Message) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Message.Unmarshal(m, b)
}
func (m *Message) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Message.Marshal(b, m, deterministic)
}
func (m *Message) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Message.Merge(m, src)
}
func (m *Message) XXX_Size() int {
	return xxx_messageInfo_Message.Size(m)
}
func (m *Message) XXX_DiscardUnknown() {
	xxx_messageInfo_Message.DiscardUnknown(m)
}

var xxx_messageInfo_Message proto.InternalMessageInfo

func (m *Message) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Message) GetText() string {
	if m != nil {
		return m.Text
	}
	return ""
}

func init() {
	proto.RegisterType((*Message)(nil), "pb.Message")
}

func init() { proto.RegisterFile("chat.proto", fileDescriptor_8c585a45e2093e54) }

var fileDescriptor_8c585a45e2093e54 = []byte{
	// 174 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x8d, 0xb1, 0x0a, 0xc2, 0x30,
	0x14, 0x45, 0xdb, 0x50, 0x94, 0xbe, 0x82, 0xc3, 0x1b, 0xa4, 0xd4, 0x45, 0x3a, 0xb9, 0x34, 0x15,
	0xf5, 0x17, 0x1c, 0x5d, 0xf4, 0x0b, 0x9a, 0xf6, 0x19, 0x03, 0x6a, 0x42, 0x92, 0x8a, 0xfe, 0xbd,
	0x34, 0x55, 0x70, 0x71, 0xbb, 0x5c, 0x0e, 0xe7, 0x00, 0xb4, 0x97, 0xc6, 0x73, 0x63, 0xb5, 0xd7,
	0xc8, 0x8c, 0x28, 0x16, 0x52, 0x6b, 0x79, 0xa5, 0x3a, 0x3c, 0xa2, 0x3f, 0xd7, 0x74, 0x33, 0xfe,
	0x35, 0x02, 0x65, 0x05, 0xd3, 0x03, 0x39, 0xd7, 0x48, 0xc2, 0x19, 0x30, 0xd5, 0xe5, 0xf1, 0x32,
	0x5e, 0xa5, 0x47, 0xa6, 0x3a, 0x44, 0x48, 0x3c, 0x3d, 0x7d, 0xce, 0xc2, 0x13, 0xf6, 0xc6, 0x42,
	0x36, 0xd8, 0x4f, 0x64, 0x1f, 0xaa, 0x25, 0xac, 0x20, 0x71, 0x74, 0xef, 0x30, 0xe3, 0x46, 0xf0,
	0x8f, 0xa7, 0x98, 0xf3, 0x31, 0xc8, 0xbf, 0x41, 0xbe, 0x1f, 0x82, 0x65, 0x84, 0x3b, 0x48, 0x5d,
	0x2f, 0x5c, 0x6b, 0x95, 0x20, 0xfc, 0x83, 0x15, 0xbf, 0xae, 0x32, 0x5a, 0xc7, 0x62, 0x12, 0x80,
	0xed, 0x3b, 0x00, 0x00, 0xff, 0xff, 0x96, 0x81, 0xa2, 0x57, 0xd8, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// ChatServiceClient is the client API for ChatService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ChatServiceClient interface {
	Send(ctx context.Context, in *Message, opts ...grpc.CallOption) (*empty.Empty, error)
	Subscribe(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (ChatService_SubscribeClient, error)
}

type chatServiceClient struct {
	cc *grpc.ClientConn
}

func NewChatServiceClient(cc *grpc.ClientConn) ChatServiceClient {
	return &chatServiceClient{cc}
}

func (c *chatServiceClient) Send(ctx context.Context, in *Message, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, "/pb.chatService/send", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *chatServiceClient) Subscribe(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (ChatService_SubscribeClient, error) {
	stream, err := c.cc.NewStream(ctx, &_ChatService_serviceDesc.Streams[0], "/pb.chatService/subscribe", opts...)
	if err != nil {
		return nil, err
	}
	x := &chatServiceSubscribeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ChatService_SubscribeClient interface {
	Recv() (*Message, error)
	grpc.ClientStream
}

type chatServiceSubscribeClient struct {
	grpc.ClientStream
}

func (x *chatServiceSubscribeClient) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ChatServiceServer is the server API for ChatService service.
type ChatServiceServer interface {
	Send(context.Context, *Message) (*empty.Empty, error)
	Subscribe(*empty.Empty, ChatService_SubscribeServer) error
}

func RegisterChatServiceServer(s *grpc.Server, srv ChatServiceServer) {
	s.RegisterService(&_ChatService_serviceDesc, srv)
}

func _ChatService_Send_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Message)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServiceServer).Send(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.chatService/Send",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServiceServer).Send(ctx, req.(*Message))
	}
	return interceptor(ctx, in, info, handler)
}

func _ChatService_Subscribe_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(empty.Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ChatServiceServer).Subscribe(m, &chatServiceSubscribeServer{stream})
}

type ChatService_SubscribeServer interface {
	Send(*Message) error
	grpc.ServerStream
}

type chatServiceSubscribeServer struct {
	grpc.ServerStream
}

func (x *chatServiceSubscribeServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

var _ChatService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "pb.chatService",
	HandlerType: (*ChatServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "send",
			Handler:    _ChatService_Send_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "subscribe",
			Handler:       _ChatService_Subscribe_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "chat.proto",
}
