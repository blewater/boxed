// Code generated by protoc-gen-go. DO NOT EDIT.
// source: tasks.proto

package wrpc

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	_struct "github.com/golang/protobuf/ptypes/struct"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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
	RemoteTasks    []string `protobuf:"bytes,3,rep,name=remoteTasks,proto3" json:"remoteTasks,omitempty"`
	ErrorMsg       string   `protobuf:"bytes,4,opt,name=errorMsg,proto3" json:"errorMsg,omitempty"`
	// Custom static configuration
	Datum string   `protobuf:"bytes,5,opt,name=datum,proto3" json:"datum,omitempty"`
	Data  []string `protobuf:"bytes,6,rep,name=data,proto3" json:"data,omitempty"`
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

func (m *ServerMsg) GetRemoteTasks() []string {
	if m != nil {
		return m.RemoteTasks
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
	// user provided workflow key ID
	WorkflowNameKey string `protobuf:"bytes,5,opt,name=workflowNameKey,proto3" json:"workflowNameKey,omitempty"`
	// Custom static configuration
	Datum string   `protobuf:"bytes,6,opt,name=datum,proto3" json:"datum,omitempty"`
	Data  []string `protobuf:"bytes,7,rep,name=data,proto3" json:"data,omitempty"`
	// Custom dynamic json configuration
	Config               *_struct.Struct `protobuf:"bytes,8,opt,name=config,proto3" json:"config,omitempty"`
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

func (m *RemoteMsg) GetWorkflowNameKey() string {
	if m != nil {
		return m.WorkflowNameKey
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
	proto.RegisterType((*ServerMsg)(nil), "grpc.ServerMsg")
	proto.RegisterType((*RemoteMsg)(nil), "grpc.RemoteMsg")
	proto.RegisterType((*RemoteMsg_Tasks)(nil), "grpc.RemoteMsg.Tasks")
}

func init() { proto.RegisterFile("tasks.proto", fileDescriptor_b3834c8ef8464a3f) }

var fileDescriptor_b3834c8ef8464a3f = []byte{
	// 380 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0xdf, 0x6a, 0xa3, 0x40,
	0x14, 0xc6, 0x77, 0x62, 0x34, 0xf1, 0x08, 0x9b, 0x65, 0xd8, 0x65, 0x07, 0x09, 0x45, 0x72, 0x51,
	0x84, 0x82, 0x29, 0x29, 0x7d, 0x82, 0x5c, 0x85, 0xd2, 0x3f, 0x98, 0x42, 0xaf, 0x7a, 0x61, 0xcc,
	0x44, 0x42, 0xa2, 0x23, 0xf3, 0xa7, 0xa1, 0x6f, 0xd0, 0x87, 0xed, 0x43, 0x94, 0x19, 0xad, 0x49,
	0x45, 0x4a, 0xef, 0x3c, 0xdf, 0x7c, 0x9e, 0x73, 0xbe, 0xdf, 0x0c, 0x78, 0x32, 0x11, 0x3b, 0x11,
	0x95, 0x9c, 0x49, 0x86, 0xfb, 0x19, 0x2f, 0x53, 0x7f, 0x9c, 0x31, 0x96, 0xed, 0xe9, 0xd4, 0x68,
	0x2b, 0xb5, 0x99, 0x0a, 0xc9, 0x55, 0x2a, 0x2b, 0xcf, 0xe4, 0x1d, 0x81, 0xbb, 0xa4, 0xfc, 0x85,
	0xf2, 0x5b, 0x91, 0xe1, 0x73, 0xf8, 0xad, 0x1b, 0x2c, 0x8a, 0x07, 0xce, 0x32, 0x4e, 0x85, 0x20,
	0x28, 0x40, 0xa1, 0x1b, 0xb7, 0x54, 0x7c, 0x06, 0xa0, 0x95, 0x7b, 0x25, 0x4b, 0x25, 0x49, 0xcf,
	0x78, 0x4e, 0x14, 0x1c, 0x80, 0xc7, 0x69, 0xce, 0x24, 0x7d, 0xd4, 0xeb, 0x10, 0x2b, 0xb0, 0x42,
	0x37, 0x3e, 0x95, 0xb0, 0x0f, 0x43, 0xca, 0x39, 0xd3, 0x53, 0x49, 0xdf, 0xfc, 0xdf, 0xd4, 0xf8,
	0x2f, 0xd8, 0xeb, 0x44, 0xaa, 0x9c, 0xd8, 0xe6, 0xa0, 0x2a, 0x30, 0x86, 0xfe, 0x3a, 0x91, 0x09,
	0x71, 0x4c, 0x33, 0xf3, 0x8d, 0xa7, 0xe0, 0xa4, 0xac, 0xd8, 0x6c, 0x33, 0x32, 0x08, 0x50, 0xe8,
	0xcd, 0xfe, 0x47, 0x55, 0xd8, 0xe8, 0x33, 0x6c, 0xb4, 0x34, 0x61, 0xe3, 0xda, 0x36, 0x79, 0xb3,
	0xc0, 0x8d, 0xcd, 0x1a, 0x7a, 0xd0, 0x05, 0xd8, 0x86, 0x17, 0x41, 0x81, 0x15, 0x7a, 0xb3, 0x7f,
	0x91, 0x06, 0x16, 0x35, 0xe7, 0x91, 0x59, 0x35, 0xae, 0x3c, 0x1d, 0x6c, 0x7a, 0x9d, 0x6c, 0x6a,
	0x9f, 0x98, 0xb3, 0xbc, 0xdc, 0x53, 0x49, 0xd7, 0xc4, 0x0a, 0x50, 0x38, 0x8c, 0x5b, 0xea, 0xb7,
	0x04, 0x42, 0x18, 0x1d, 0x18, 0xdf, 0x6d, 0xf6, 0xec, 0x70, 0x97, 0xe4, 0xf4, 0x86, 0xbe, 0xd6,
	0x2c, 0xda, 0xf2, 0x91, 0x95, 0xd3, 0xc5, 0x6a, 0xd0, 0xc9, 0x6a, 0xf8, 0x23, 0x56, 0xfe, 0x33,
	0xd8, 0xcd, 0x5d, 0xe9, 0xdd, 0xf5, 0xc8, 0xfa, 0x3d, 0x34, 0xf5, 0x97, 0x14, 0xbd, 0x56, 0x8a,
	0x31, 0xb8, 0x69, 0x0b, 0xc2, 0x51, 0x98, 0x2d, 0xe0, 0x8f, 0x6e, 0x3f, 0x67, 0x79, 0xae, 0x8a,
	0x6d, 0x9a, 0x48, 0xc6, 0xf1, 0x35, 0x78, 0xb1, 0x2a, 0x9e, 0xea, 0x8c, 0x78, 0xd4, 0xba, 0x10,
	0xbf, 0x16, 0x9a, 0x07, 0x3b, 0xf9, 0x15, 0xa2, 0x4b, 0xb4, 0x72, 0x4c, 0x84, 0xab, 0x8f, 0x00,
	0x00, 0x00, 0xff, 0xff, 0xd7, 0xaa, 0x7b, 0xc1, 0xfe, 0x02, 0x00, 0x00,
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
	RunWorkflow(ctx context.Context, opts ...grpc.CallOption) (TaskCommunicator_RunWorkflowClient, error)
}

type taskCommunicatorClient struct {
	cc grpc.ClientConnInterface
}

func NewTaskCommunicatorClient(cc grpc.ClientConnInterface) TaskCommunicatorClient {
	return &taskCommunicatorClient{cc}
}

func (c *taskCommunicatorClient) RunWorkflow(ctx context.Context, opts ...grpc.CallOption) (TaskCommunicator_RunWorkflowClient, error) {
	stream, err := c.cc.NewStream(ctx, &_TaskCommunicator_serviceDesc.Streams[0], "/grpc.TaskCommunicator/RunWorkflow", opts...)
	if err != nil {
		return nil, err
	}
	x := &taskCommunicatorRunWorkflowClient{stream}
	return x, nil
}

type TaskCommunicator_RunWorkflowClient interface {
	Send(*RemoteMsg) error
	Recv() (*ServerMsg, error)
	grpc.ClientStream
}

type taskCommunicatorRunWorkflowClient struct {
	grpc.ClientStream
}

func (x *taskCommunicatorRunWorkflowClient) Send(m *RemoteMsg) error {
	return x.ClientStream.SendMsg(m)
}

func (x *taskCommunicatorRunWorkflowClient) Recv() (*ServerMsg, error) {
	m := new(ServerMsg)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// TaskCommunicatorServer is the server API for TaskCommunicator service.
type TaskCommunicatorServer interface {
	RunWorkflow(TaskCommunicator_RunWorkflowServer) error
}

// UnimplementedTaskCommunicatorServer can be embedded to have forward compatible implementations.
type UnimplementedTaskCommunicatorServer struct {
}

func (*UnimplementedTaskCommunicatorServer) RunWorkflow(srv TaskCommunicator_RunWorkflowServer) error {
	return status.Errorf(codes.Unimplemented, "method RunWorkflow not implemented")
}

func RegisterTaskCommunicatorServer(s *grpc.Server, srv TaskCommunicatorServer) {
	s.RegisterService(&_TaskCommunicator_serviceDesc, srv)
}

func _TaskCommunicator_RunWorkflow_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TaskCommunicatorServer).RunWorkflow(&taskCommunicatorRunWorkflowServer{stream})
}

type TaskCommunicator_RunWorkflowServer interface {
	Send(*ServerMsg) error
	Recv() (*RemoteMsg, error)
	grpc.ServerStream
}

type taskCommunicatorRunWorkflowServer struct {
	grpc.ServerStream
}

func (x *taskCommunicatorRunWorkflowServer) Send(m *ServerMsg) error {
	return x.ServerStream.SendMsg(m)
}

func (x *taskCommunicatorRunWorkflowServer) Recv() (*RemoteMsg, error) {
	m := new(RemoteMsg)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _TaskCommunicator_serviceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.TaskCommunicator",
	HandlerType: (*TaskCommunicatorServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "RunWorkflow",
			Handler:       _TaskCommunicator_RunWorkflow_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "tasks.proto",
}
