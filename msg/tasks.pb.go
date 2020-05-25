// Code generated by protoc-gen-go. DO NOT EDIT.
// source: tasks.proto

package msg

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_struct "github.com/golang/protobuf/ptypes/struct"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

// Server Message Task Message
type ServerMsg struct {
	// Fixed Workflow Properties
	TaskInProgress string   `protobuf:"bytes,1,opt,name=taskInProgress,proto3" json:"taskInProgress,omitempty"`
	TaskOutput     string   `protobuf:"bytes,2,opt,name=taskOutput,proto3" json:"taskOutput,omitempty"`
	ClientTasks    []string `protobuf:"bytes,3,rep,name=clientTasks,proto3" json:"clientTasks,omitempty"`
	ErrorMsg       string   `protobuf:"bytes,4,opt,name=errorMsg,proto3" json:"errorMsg,omitempty"`
	Datum          string   `protobuf:"bytes,5,opt,name=datum,proto3" json:"datum,omitempty"`
	// Custom static configuration
	Data []string `protobuf:"bytes,6,rep,name=data,proto3" json:"data,omitempty"`
	// Custom dynamic json configuration
	Config               *_struct.Struct `protobuf:"bytes,7,opt,name=config,proto3" json:"config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *ServerMsg) Reset()         { *m = ServerMsg{} }
func (m *ServerMsg) String() string { return proto.CompactTextString(m) }
func (*ServerMsg) ProtoMessage()    {}
func (*ServerMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_b3834c8ef8464a3f, []int{0}
}

func (m *ServerMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ServerMsg.Unmarshal(m, b)
}
func (m *ServerMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ServerMsg.Marshal(b, m, deterministic)
}
func (m *ServerMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ServerMsg.Merge(m, src)
}
func (m *ServerMsg) XXX_Size() int {
	return xxx_messageInfo_ServerMsg.Size(m)
}
func (m *ServerMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_ServerMsg.DiscardUnknown(m)
}

var xxx_messageInfo_ServerMsg proto.InternalMessageInfo

func (m *ServerMsg) GetTaskInProgress() string {
	if m != nil {
		return m.TaskInProgress
	}
	return ""
}

func (m *ServerMsg) GetTaskOutput() string {
	if m != nil {
		return m.TaskOutput
	}
	return ""
}

func (m *ServerMsg) GetClientTasks() []string {
	if m != nil {
		return m.ClientTasks
	}
	return nil
}

func (m *ServerMsg) GetErrorMsg() string {
	if m != nil {
		return m.ErrorMsg
	}
	return ""
}

func (m *ServerMsg) GetDatum() string {
	if m != nil {
		return m.Datum
	}
	return ""
}

func (m *ServerMsg) GetData() []string {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *ServerMsg) GetConfig() *_struct.Struct {
	if m != nil {
		return m.Config
	}
	return nil
}

// Remote or Client Task Message
type RemoteMsg struct {
	Tasks          []*RemoteMsg_Tasks `protobuf:"bytes,1,rep,name=tasks,proto3" json:"tasks,omitempty"`
	TaskInProgress string             `protobuf:"bytes,2,opt,name=taskInProgress,proto3" json:"taskInProgress,omitempty"`
	TasksCompleted bool               `protobuf:"varint,3,opt,name=tasksCompleted,proto3" json:"tasksCompleted,omitempty"`
	ErrorMsg       string             `protobuf:"bytes,4,opt,name=errorMsg,proto3" json:"errorMsg,omitempty"`
	// Custom static configuration
	Datum string   `protobuf:"bytes,5,opt,name=datum,proto3" json:"datum,omitempty"`
	Data  []string `protobuf:"bytes,6,rep,name=data,proto3" json:"data,omitempty"`
	// Custom dynamic json configuration
	Config               *_struct.Struct `protobuf:"bytes,7,opt,name=config,proto3" json:"config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *RemoteMsg) Reset()         { *m = RemoteMsg{} }
func (m *RemoteMsg) String() string { return proto.CompactTextString(m) }
func (*RemoteMsg) ProtoMessage()    {}
func (*RemoteMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_b3834c8ef8464a3f, []int{1}
}

func (m *RemoteMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemoteMsg.Unmarshal(m, b)
}
func (m *RemoteMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemoteMsg.Marshal(b, m, deterministic)
}
func (m *RemoteMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoteMsg.Merge(m, src)
}
func (m *RemoteMsg) XXX_Size() int {
	return xxx_messageInfo_RemoteMsg.Size(m)
}
func (m *RemoteMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoteMsg.DiscardUnknown(m)
}

var xxx_messageInfo_RemoteMsg proto.InternalMessageInfo

func (m *RemoteMsg) GetTasks() []*RemoteMsg_Tasks {
	if m != nil {
		return m.Tasks
	}
	return nil
}

func (m *RemoteMsg) GetTaskInProgress() string {
	if m != nil {
		return m.TaskInProgress
	}
	return ""
}

func (m *RemoteMsg) GetTasksCompleted() bool {
	if m != nil {
		return m.TasksCompleted
	}
	return false
}

func (m *RemoteMsg) GetErrorMsg() string {
	if m != nil {
		return m.ErrorMsg
	}
	return ""
}

func (m *RemoteMsg) GetDatum() string {
	if m != nil {
		return m.Datum
	}
	return ""
}

func (m *RemoteMsg) GetData() []string {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *RemoteMsg) GetConfig() *_struct.Struct {
	if m != nil {
		return m.Config
	}
	return nil
}

// Fixed Workflow Properties
type RemoteMsg_Tasks struct {
	TaskName             string   `protobuf:"bytes,1,opt,name=taskName,proto3" json:"taskName,omitempty"`
	ErrorMsg             string   `protobuf:"bytes,2,opt,name=errorMsg,proto3" json:"errorMsg,omitempty"`
	Completed            bool     `protobuf:"varint,3,opt,name=completed,proto3" json:"completed,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RemoteMsg_Tasks) Reset()         { *m = RemoteMsg_Tasks{} }
func (m *RemoteMsg_Tasks) String() string { return proto.CompactTextString(m) }
func (*RemoteMsg_Tasks) ProtoMessage()    {}
func (*RemoteMsg_Tasks) Descriptor() ([]byte, []int) {
	return fileDescriptor_b3834c8ef8464a3f, []int{1, 0}
}

func (m *RemoteMsg_Tasks) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RemoteMsg_Tasks.Unmarshal(m, b)
}
func (m *RemoteMsg_Tasks) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RemoteMsg_Tasks.Marshal(b, m, deterministic)
}
func (m *RemoteMsg_Tasks) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RemoteMsg_Tasks.Merge(m, src)
}
func (m *RemoteMsg_Tasks) XXX_Size() int {
	return xxx_messageInfo_RemoteMsg_Tasks.Size(m)
}
func (m *RemoteMsg_Tasks) XXX_DiscardUnknown() {
	xxx_messageInfo_RemoteMsg_Tasks.DiscardUnknown(m)
}

var xxx_messageInfo_RemoteMsg_Tasks proto.InternalMessageInfo

func (m *RemoteMsg_Tasks) GetTaskName() string {
	if m != nil {
		return m.TaskName
	}
	return ""
}

func (m *RemoteMsg_Tasks) GetErrorMsg() string {
	if m != nil {
		return m.ErrorMsg
	}
	return ""
}

func (m *RemoteMsg_Tasks) GetCompleted() bool {
	if m != nil {
		return m.Completed
	}
	return false
}

func init() {
	proto.RegisterType((*ServerMsg)(nil), "msg.ServerMsg")
	proto.RegisterType((*RemoteMsg)(nil), "msg.RemoteMsg")
	proto.RegisterType((*RemoteMsg_Tasks)(nil), "msg.RemoteMsg.Tasks")
}

func init() { proto.RegisterFile("tasks.proto", fileDescriptor_b3834c8ef8464a3f) }

var fileDescriptor_b3834c8ef8464a3f = []byte{
	// 344 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xc4, 0x52, 0x4b, 0x6e, 0xc2, 0x30,
	0x14, 0xac, 0x13, 0x42, 0xc9, 0x8b, 0x84, 0x2a, 0x0b, 0xa9, 0x56, 0x84, 0xaa, 0x88, 0x45, 0x15,
	0x75, 0x11, 0x10, 0x3d, 0x02, 0xdd, 0x74, 0xd1, 0x8f, 0x4c, 0xb7, 0x5d, 0x84, 0x60, 0x2c, 0x54,
	0x1c, 0x23, 0x7f, 0x7a, 0xdb, 0xde, 0xa0, 0x87, 0xa8, 0xec, 0xd0, 0x14, 0x22, 0xf6, 0xdd, 0x65,
	0xc6, 0x63, 0xbf, 0x99, 0x79, 0x81, 0xc4, 0x94, 0xfa, 0x43, 0x17, 0x7b, 0x25, 0x8d, 0xc4, 0xa1,
	0xd0, 0x3c, 0x1d, 0x73, 0x29, 0xf9, 0x8e, 0x4d, 0x3d, 0xb5, 0xb2, 0x9b, 0xa9, 0x36, 0xca, 0x56,
	0xa6, 0x91, 0x4c, 0xbe, 0x11, 0xc4, 0x4b, 0xa6, 0x3e, 0x99, 0x7a, 0xd2, 0x1c, 0xdf, 0xc2, 0xd0,
	0xdd, 0x7f, 0xac, 0x5f, 0x95, 0xe4, 0x8a, 0x69, 0x4d, 0x50, 0x86, 0xf2, 0x98, 0x76, 0x58, 0x7c,
	0x03, 0xe0, 0x98, 0x17, 0x6b, 0xf6, 0xd6, 0x90, 0xc0, 0x6b, 0x8e, 0x18, 0x9c, 0x41, 0x52, 0xed,
	0xb6, 0xac, 0x36, 0x6f, 0xce, 0x0d, 0x09, 0xb3, 0x30, 0x8f, 0xe9, 0x31, 0x85, 0x53, 0x18, 0x30,
	0xa5, 0xa4, 0x9b, 0x4a, 0x7a, 0xfe, 0x7e, 0x8b, 0xf1, 0x08, 0xa2, 0x75, 0x69, 0xac, 0x20, 0x91,
	0x3f, 0x68, 0x00, 0xc6, 0xd0, 0x5b, 0x97, 0xa6, 0x24, 0x7d, 0xff, 0x98, 0xff, 0xc6, 0x53, 0xe8,
	0x57, 0xb2, 0xde, 0x6c, 0x39, 0xb9, 0xcc, 0x50, 0x9e, 0xcc, 0xaf, 0x8b, 0x26, 0x6c, 0xf1, 0x1b,
	0xb6, 0x58, 0xfa, 0xb0, 0xf4, 0x20, 0x9b, 0x7c, 0x05, 0x10, 0x53, 0x26, 0xa4, 0x61, 0x6e, 0xd0,
	0x1d, 0x44, 0xbe, 0x2e, 0x82, 0xb2, 0x30, 0x4f, 0xe6, 0xa3, 0x42, 0x68, 0x5e, 0xb4, 0xc7, 0x85,
	0x77, 0x4a, 0x1b, 0xc9, 0x99, 0x6a, 0x82, 0xb3, 0xd5, 0x1c, 0x74, 0x7a, 0x21, 0xc5, 0x7e, 0xc7,
	0x0c, 0x5b, 0x93, 0x30, 0x43, 0xf9, 0x80, 0x76, 0xd8, 0x7f, 0x2a, 0x20, 0x7d, 0x87, 0xa8, 0x5d,
	0x80, 0x73, 0xf4, 0x5c, 0x0a, 0x76, 0x58, 0x72, 0x8b, 0x4f, 0xbc, 0x05, 0x1d, 0x6f, 0x63, 0x88,
	0xab, 0x4e, 0xb4, 0x3f, 0x62, 0xfe, 0x00, 0x57, 0xee, 0xf9, 0x85, 0x14, 0xc2, 0xd6, 0xdb, 0xaa,
	0x34, 0x52, 0xe1, 0x19, 0x0c, 0xa8, 0xad, 0x9b, 0xa9, 0xc3, 0xd3, 0x8a, 0xd3, 0x06, 0xb7, 0x3f,
	0xe0, 0xe4, 0x22, 0x47, 0x33, 0xb4, 0xea, 0x7b, 0xf7, 0xf7, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff,
	0x98, 0x42, 0x8a, 0xeb, 0xcd, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// TaskCommunicatorClient is the client API for TaskCommunicator service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TaskCommunicatorClient interface {
	RunTasks(ctx context.Context, opts ...grpc.CallOption) (TaskCommunicator_RunTasksClient, error)
}

type taskCommunicatorClient struct {
	cc grpc.ClientConnInterface
}

func NewTaskCommunicatorClient(cc grpc.ClientConnInterface) TaskCommunicatorClient {
	return &taskCommunicatorClient{cc}
}

func (c *taskCommunicatorClient) RunTasks(ctx context.Context, opts ...grpc.CallOption) (TaskCommunicator_RunTasksClient, error) {
	stream, err := c.cc.NewStream(ctx, &_TaskCommunicator_serviceDesc.Streams[0], "/msg.TaskCommunicator/RunTasks", opts...)
	if err != nil {
		return nil, err
	}
	x := &taskCommunicatorRunTasksClient{stream}
	return x, nil
}

type TaskCommunicator_RunTasksClient interface {
	Send(*RemoteMsg) error
	Recv() (*ServerMsg, error)
	grpc.ClientStream
}

type taskCommunicatorRunTasksClient struct {
	grpc.ClientStream
}

func (x *taskCommunicatorRunTasksClient) Send(m *RemoteMsg) error {
	return x.ClientStream.SendMsg(m)
}

func (x *taskCommunicatorRunTasksClient) Recv() (*ServerMsg, error) {
	m := new(ServerMsg)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TaskCommunicatorServer is the server API for TaskCommunicator service.
type TaskCommunicatorServer interface {
	RunTasks(TaskCommunicator_RunTasksServer) error
}

// UnimplementedTaskCommunicatorServer can be embedded to have forward compatible implementations.
type UnimplementedTaskCommunicatorServer struct {
}

func (*UnimplementedTaskCommunicatorServer) RunTasks(srv TaskCommunicator_RunTasksServer) error {
	return status.Errorf(codes.Unimplemented, "method RunTasks not implemented")
}

func RegisterTaskCommunicatorServer(s *grpc.Server, srv TaskCommunicatorServer) {
	s.RegisterService(&_TaskCommunicator_serviceDesc, srv)
}

func _TaskCommunicator_RunTasks_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TaskCommunicatorServer).RunTasks(&taskCommunicatorRunTasksServer{stream})
}

type TaskCommunicator_RunTasksServer interface {
	Send(*ServerMsg) error
	Recv() (*RemoteMsg, error)
	grpc.ServerStream
}

type taskCommunicatorRunTasksServer struct {
	grpc.ServerStream
}

func (x *taskCommunicatorRunTasksServer) Send(m *ServerMsg) error {
	return x.ServerStream.SendMsg(m)
}

func (x *taskCommunicatorRunTasksServer) Recv() (*RemoteMsg, error) {
	m := new(RemoteMsg)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _TaskCommunicator_serviceDesc = grpc.ServiceDesc{
	ServiceName: "msg.TaskCommunicator",
	HandlerType: (*TaskCommunicatorServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "RunTasks",
			Handler:       _TaskCommunicator_RunTasks_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "tasks.proto",
}
