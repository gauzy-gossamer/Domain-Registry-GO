//
//gRPC declaration
//
//compile python mapping with:
//pip3 install grpcio-tools
//python3 -m grpc_tools.protoc -I. --python_out=. --pyi_out=. --grpc_python_out=. registry.proto
//
//compile go mapping with:
//apt install protobuf-compiler
//go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
//protoc --go-grpc_out=./ --go_out=./ filename

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.21.12
// source: logger.proto

package logger

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type LogRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ServiceID    uint32 `protobuf:"varint,1,opt,name=ServiceID,proto3" json:"ServiceID,omitempty"`
	RequestType  uint32 `protobuf:"varint,2,opt,name=RequestType,proto3" json:"RequestType,omitempty"`
	SessionID    int64  `protobuf:"varint,3,opt,name=SessionID,proto3" json:"SessionID,omitempty"`
	UserID       uint64 `protobuf:"varint,4,opt,name=UserID,proto3" json:"UserID,omitempty"`
	IsMonitoring bool   `protobuf:"varint,5,opt,name=IsMonitoring,proto3" json:"IsMonitoring,omitempty"`
	SourceIP     string `protobuf:"bytes,6,opt,name=SourceIP,proto3" json:"SourceIP,omitempty"`
	Content      string `protobuf:"bytes,7,opt,name=Content,proto3" json:"Content,omitempty"`
}

func (x *LogRequest) Reset() {
	*x = LogRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logger_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogRequest) ProtoMessage() {}

func (x *LogRequest) ProtoReflect() protoreflect.Message {
	mi := &file_logger_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogRequest.ProtoReflect.Descriptor instead.
func (*LogRequest) Descriptor() ([]byte, []int) {
	return file_logger_proto_rawDescGZIP(), []int{0}
}

func (x *LogRequest) GetServiceID() uint32 {
	if x != nil {
		return x.ServiceID
	}
	return 0
}

func (x *LogRequest) GetRequestType() uint32 {
	if x != nil {
		return x.RequestType
	}
	return 0
}

func (x *LogRequest) GetSessionID() int64 {
	if x != nil {
		return x.SessionID
	}
	return 0
}

func (x *LogRequest) GetUserID() uint64 {
	if x != nil {
		return x.UserID
	}
	return 0
}

func (x *LogRequest) GetIsMonitoring() bool {
	if x != nil {
		return x.IsMonitoring
	}
	return false
}

func (x *LogRequest) GetSourceIP() string {
	if x != nil {
		return x.SourceIP
	}
	return ""
}

func (x *LogRequest) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type EndReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LogID        uint64 `protobuf:"varint,1,opt,name=LogID,proto3" json:"LogID,omitempty"`
	ResponseCode uint32 `protobuf:"varint,2,opt,name=ResponseCode,proto3" json:"ResponseCode,omitempty"`
	Content      string `protobuf:"bytes,3,opt,name=Content,proto3" json:"Content,omitempty"`
}

func (x *EndReq) Reset() {
	*x = EndReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logger_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EndReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EndReq) ProtoMessage() {}

func (x *EndReq) ProtoReflect() protoreflect.Message {
	mi := &file_logger_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EndReq.ProtoReflect.Descriptor instead.
func (*EndReq) Descriptor() ([]byte, []int) {
	return file_logger_proto_rawDescGZIP(), []int{1}
}

func (x *EndReq) GetLogID() uint64 {
	if x != nil {
		return x.LogID
	}
	return 0
}

func (x *EndReq) GetResponseCode() uint32 {
	if x != nil {
		return x.ResponseCode
	}
	return 0
}

func (x *EndReq) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type LogID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LogID      uint64 `protobuf:"varint,1,opt,name=LogID,proto3" json:"LogID,omitempty"`
	ReturnCode int32  `protobuf:"varint,2,opt,name=return_code,json=returnCode,proto3" json:"return_code,omitempty"`
}

func (x *LogID) Reset() {
	*x = LogID{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logger_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogID) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogID) ProtoMessage() {}

func (x *LogID) ProtoReflect() protoreflect.Message {
	mi := &file_logger_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogID.ProtoReflect.Descriptor instead.
func (*LogID) Descriptor() ([]byte, []int) {
	return file_logger_proto_rawDescGZIP(), []int{2}
}

func (x *LogID) GetLogID() uint64 {
	if x != nil {
		return x.LogID
	}
	return 0
}

func (x *LogID) GetReturnCode() int32 {
	if x != nil {
		return x.ReturnCode
	}
	return 0
}

type LogStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ReturnCode int32 `protobuf:"varint,1,opt,name=return_code,json=returnCode,proto3" json:"return_code,omitempty"`
}

func (x *LogStatus) Reset() {
	*x = LogStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_logger_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LogStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogStatus) ProtoMessage() {}

func (x *LogStatus) ProtoReflect() protoreflect.Message {
	mi := &file_logger_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogStatus.ProtoReflect.Descriptor instead.
func (*LogStatus) Descriptor() ([]byte, []int) {
	return file_logger_proto_rawDescGZIP(), []int{3}
}

func (x *LogStatus) GetReturnCode() int32 {
	if x != nil {
		return x.ReturnCode
	}
	return 0
}

var File_logger_proto protoreflect.FileDescriptor

var file_logger_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x6c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xdc,
	0x01, 0x0a, 0x0a, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a,
	0x09, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x09, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x44, 0x12, 0x20, 0x0a, 0x0b, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x54, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x0b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1c, 0x0a,
	0x09, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x09, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x12, 0x16, 0x0a, 0x06, 0x55,
	0x73, 0x65, 0x72, 0x49, 0x44, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x55, 0x73, 0x65,
	0x72, 0x49, 0x44, 0x12, 0x22, 0x0a, 0x0c, 0x49, 0x73, 0x4d, 0x6f, 0x6e, 0x69, 0x74, 0x6f, 0x72,
	0x69, 0x6e, 0x67, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x49, 0x73, 0x4d, 0x6f, 0x6e,
	0x69, 0x74, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x53, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x49, 0x50, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x53, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x49, 0x50, 0x12, 0x18, 0x0a, 0x07, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x07,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x5c, 0x0a,
	0x06, 0x45, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x12, 0x14, 0x0a, 0x05, 0x4c, 0x6f, 0x67, 0x49, 0x44,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x4c, 0x6f, 0x67, 0x49, 0x44, 0x12, 0x22, 0x0a,
	0x0c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x0c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x43, 0x6f, 0x64,
	0x65, 0x12, 0x18, 0x0a, 0x07, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x43, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x3e, 0x0a, 0x05, 0x4c,
	0x6f, 0x67, 0x49, 0x44, 0x12, 0x14, 0x0a, 0x05, 0x4c, 0x6f, 0x67, 0x49, 0x44, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x05, 0x4c, 0x6f, 0x67, 0x49, 0x44, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65,
	0x74, 0x75, 0x72, 0x6e, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x0a, 0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x43, 0x6f, 0x64, 0x65, 0x22, 0x2c, 0x0a, 0x09, 0x4c,
	0x6f, 0x67, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x72, 0x65, 0x74, 0x75,
	0x72, 0x6e, 0x5f, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x72,
	0x65, 0x74, 0x75, 0x72, 0x6e, 0x43, 0x6f, 0x64, 0x65, 0x32, 0x57, 0x0a, 0x09, 0x52, 0x65, 0x67,
	0x4c, 0x6f, 0x67, 0x67, 0x65, 0x72, 0x12, 0x25, 0x0a, 0x0c, 0x53, 0x74, 0x61, 0x72, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0b, 0x2e, 0x4c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x06, 0x2e, 0x4c, 0x6f, 0x67, 0x49, 0x44, 0x22, 0x00, 0x12, 0x23, 0x0a,
	0x0a, 0x45, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x07, 0x2e, 0x45, 0x6e,
	0x64, 0x52, 0x65, 0x71, 0x1a, 0x0a, 0x2e, 0x4c, 0x6f, 0x67, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x22, 0x00, 0x42, 0x07, 0x5a, 0x05, 0x2e, 0x2f, 0x64, 0x69, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_logger_proto_rawDescOnce sync.Once
	file_logger_proto_rawDescData = file_logger_proto_rawDesc
)

func file_logger_proto_rawDescGZIP() []byte {
	file_logger_proto_rawDescOnce.Do(func() {
		file_logger_proto_rawDescData = protoimpl.X.CompressGZIP(file_logger_proto_rawDescData)
	})
	return file_logger_proto_rawDescData
}

var file_logger_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_logger_proto_goTypes = []interface{}{
	(*LogRequest)(nil), // 0: LogRequest
	(*EndReq)(nil),     // 1: EndReq
	(*LogID)(nil),      // 2: LogID
	(*LogStatus)(nil),  // 3: LogStatus
}
var file_logger_proto_depIdxs = []int32{
	0, // 0: RegLogger.StartRequest:input_type -> LogRequest
	1, // 1: RegLogger.EndRequest:input_type -> EndReq
	2, // 2: RegLogger.StartRequest:output_type -> LogID
	3, // 3: RegLogger.EndRequest:output_type -> LogStatus
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_logger_proto_init() }
func file_logger_proto_init() {
	if File_logger_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_logger_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_logger_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EndReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_logger_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogID); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_logger_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LogStatus); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_logger_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_logger_proto_goTypes,
		DependencyIndexes: file_logger_proto_depIdxs,
		MessageInfos:      file_logger_proto_msgTypes,
	}.Build()
	File_logger_proto = out.File
	file_logger_proto_rawDesc = nil
	file_logger_proto_goTypes = nil
	file_logger_proto_depIdxs = nil
}
