// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.15.2
// source: atlas/atlasrpc/v1/error_fields.proto

package atlasrpc

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

// FieldsInfo is a default representation of field details that conforms
// REST API Syntax Specification
type FieldInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Fields map[string]*StringListValue `protobuf:"bytes,1,rep,name=fields,proto3" json:"fields,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *FieldInfo) Reset() {
	*x = FieldInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_atlas_atlasrpc_v1_error_fields_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FieldInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FieldInfo) ProtoMessage() {}

func (x *FieldInfo) ProtoReflect() protoreflect.Message {
	mi := &file_atlas_atlasrpc_v1_error_fields_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FieldInfo.ProtoReflect.Descriptor instead.
func (*FieldInfo) Descriptor() ([]byte, []int) {
	return file_atlas_atlasrpc_v1_error_fields_proto_rawDescGZIP(), []int{0}
}

func (x *FieldInfo) GetFields() map[string]*StringListValue {
	if x != nil {
		return x.Fields
	}
	return nil
}

type StringListValue struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Values []string `protobuf:"bytes,1,rep,name=values,proto3" json:"values,omitempty"`
}

func (x *StringListValue) Reset() {
	*x = StringListValue{}
	if protoimpl.UnsafeEnabled {
		mi := &file_atlas_atlasrpc_v1_error_fields_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StringListValue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StringListValue) ProtoMessage() {}

func (x *StringListValue) ProtoReflect() protoreflect.Message {
	mi := &file_atlas_atlasrpc_v1_error_fields_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StringListValue.ProtoReflect.Descriptor instead.
func (*StringListValue) Descriptor() ([]byte, []int) {
	return file_atlas_atlasrpc_v1_error_fields_proto_rawDescGZIP(), []int{1}
}

func (x *StringListValue) GetValues() []string {
	if x != nil {
		return x.Values
	}
	return nil
}

var File_atlas_atlasrpc_v1_error_fields_proto protoreflect.FileDescriptor

var file_atlas_atlasrpc_v1_error_fields_proto_rawDesc = []byte{
	0x0a, 0x24, 0x61, 0x74, 0x6c, 0x61, 0x73, 0x2f, 0x61, 0x74, 0x6c, 0x61, 0x73, 0x72, 0x70, 0x63,
	0x2f, 0x76, 0x31, 0x2f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x5f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x11, 0x61, 0x74, 0x6c, 0x61, 0x73, 0x2e, 0x61, 0x74,
	0x6c, 0x61, 0x73, 0x72, 0x70, 0x63, 0x2e, 0x76, 0x31, 0x22, 0xac, 0x01, 0x0a, 0x09, 0x46, 0x69,
	0x65, 0x6c, 0x64, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x40, 0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x28, 0x2e, 0x61, 0x74, 0x6c, 0x61, 0x73, 0x2e,
	0x61, 0x74, 0x6c, 0x61, 0x73, 0x72, 0x70, 0x63, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x69, 0x65, 0x6c,
	0x64, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x1a, 0x5d, 0x0a, 0x0b, 0x46, 0x69, 0x65,
	0x6c, 0x64, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x38, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x61, 0x74, 0x6c, 0x61,
	0x73, 0x2e, 0x61, 0x74, 0x6c, 0x61, 0x73, 0x72, 0x70, 0x63, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x4c, 0x69, 0x73, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x29, 0x0a, 0x0f, 0x53, 0x74, 0x72, 0x69,
	0x6e, 0x67, 0x4c, 0x69, 0x73, 0x74, 0x56, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x73, 0x42, 0x46, 0x5a, 0x44, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x69, 0x6e, 0x66, 0x6f, 0x62, 0x6c, 0x6f, 0x78, 0x6f, 0x70, 0x65, 0x6e, 0x2f, 0x61,
	0x74, 0x6c, 0x61, 0x73, 0x2d, 0x61, 0x70, 0x70, 0x2d, 0x74, 0x6f, 0x6f, 0x6c, 0x6b, 0x69, 0x74,
	0x2f, 0x76, 0x32, 0x2f, 0x61, 0x74, 0x6c, 0x61, 0x73, 0x2f, 0x61, 0x74, 0x6c, 0x61, 0x73, 0x72,
	0x70, 0x63, 0x3b, 0x61, 0x74, 0x6c, 0x61, 0x73, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_atlas_atlasrpc_v1_error_fields_proto_rawDescOnce sync.Once
	file_atlas_atlasrpc_v1_error_fields_proto_rawDescData = file_atlas_atlasrpc_v1_error_fields_proto_rawDesc
)

func file_atlas_atlasrpc_v1_error_fields_proto_rawDescGZIP() []byte {
	file_atlas_atlasrpc_v1_error_fields_proto_rawDescOnce.Do(func() {
		file_atlas_atlasrpc_v1_error_fields_proto_rawDescData = protoimpl.X.CompressGZIP(file_atlas_atlasrpc_v1_error_fields_proto_rawDescData)
	})
	return file_atlas_atlasrpc_v1_error_fields_proto_rawDescData
}

var file_atlas_atlasrpc_v1_error_fields_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_atlas_atlasrpc_v1_error_fields_proto_goTypes = []interface{}{
	(*FieldInfo)(nil),       // 0: atlas.atlasrpc.v1.FieldInfo
	(*StringListValue)(nil), // 1: atlas.atlasrpc.v1.StringListValue
	nil,                     // 2: atlas.atlasrpc.v1.FieldInfo.FieldsEntry
}
var file_atlas_atlasrpc_v1_error_fields_proto_depIdxs = []int32{
	2, // 0: atlas.atlasrpc.v1.FieldInfo.fields:type_name -> atlas.atlasrpc.v1.FieldInfo.FieldsEntry
	1, // 1: atlas.atlasrpc.v1.FieldInfo.FieldsEntry.value:type_name -> atlas.atlasrpc.v1.StringListValue
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_atlas_atlasrpc_v1_error_fields_proto_init() }
func file_atlas_atlasrpc_v1_error_fields_proto_init() {
	if File_atlas_atlasrpc_v1_error_fields_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_atlas_atlasrpc_v1_error_fields_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FieldInfo); i {
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
		file_atlas_atlasrpc_v1_error_fields_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StringListValue); i {
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
			RawDescriptor: file_atlas_atlasrpc_v1_error_fields_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_atlas_atlasrpc_v1_error_fields_proto_goTypes,
		DependencyIndexes: file_atlas_atlasrpc_v1_error_fields_proto_depIdxs,
		MessageInfos:      file_atlas_atlasrpc_v1_error_fields_proto_msgTypes,
	}.Build()
	File_atlas_atlasrpc_v1_error_fields_proto = out.File
	file_atlas_atlasrpc_v1_error_fields_proto_rawDesc = nil
	file_atlas_atlasrpc_v1_error_fields_proto_goTypes = nil
	file_atlas_atlasrpc_v1_error_fields_proto_depIdxs = nil
}
