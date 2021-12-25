// SPDX-License-Identifier: MIT

package server

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/dynamicpb"
)

var (
	pf protoreflect.FileDescriptor

	pb = &descriptorpb.FileDescriptorProto{
		Syntax:  proto.String("proto3"),
		Name:    proto.String("result.proto"),
		Package: proto.String("server"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: proto.String("Field"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("name"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_Type(protoreflect.StringKind).Enum(),
					},
					{
						Name:   proto.String("message"),
						Number: proto.Int32(2),
						Label:  descriptorpb.FieldDescriptorProto_Label(protoreflect.Repeated).Enum(),
						Type:   descriptorpb.FieldDescriptorProto_Type(protoreflect.StringKind).Enum(),
					},
				},
				NestedType: []*descriptorpb.DescriptorProto{
					{
						Name: proto.String("Field"),
						Field: []*descriptorpb.FieldDescriptorProto{
							{
								Name:   proto.String("name"),
								Number: proto.Int32(1),
								Type:   descriptorpb.FieldDescriptorProto_Type(protoreflect.StringKind).Enum(),
							},
							{
								Name:   proto.String("message"),
								Number: proto.Int32(2),
								Label:  descriptorpb.FieldDescriptorProto_Label(protoreflect.Repeated).Enum(),
								Type:   descriptorpb.FieldDescriptorProto_Type(protoreflect.StringKind).Enum(),
							},
						},
					},
				},
			},
			{
				Name: proto.String("Result"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:   proto.String("message"),
						Number: proto.Int32(1),
						Type:   descriptorpb.FieldDescriptorProto_Type(protoreflect.StringKind).Enum(),
					},
					{
						Name:   proto.String("code"),
						Number: proto.Int32(2),
						Type:   descriptorpb.FieldDescriptorProto_Type(protoreflect.StringKind).Enum(),
					},
					{
						Name:     proto.String("fields"),
						Number:   proto.Int32(3),
						Label:    descriptorpb.FieldDescriptorProto_Label(protoreflect.Repeated).Enum(),
						Type:     descriptorpb.FieldDescriptorProto_Type(protoreflect.MessageKind).Enum(),
						TypeName: proto.String(".server.Field"),
					},
				},
			},
		},
	}
)

func init() {
	var err error
	pf, err = protodesc.NewFile(pb, nil)
	if err != nil {
		panic(err)
	}
}

// ProtoReflect 实现了 proto.Message 接口
func (rslt *defaultResult) ProtoReflect() protoreflect.Message {
	m := pf.Messages().ByName("Result")
	msg := dynamicpb.NewMessage(m)
	msg.Set(m.Fields().ByName("code"), protoreflect.ValueOfString(rslt.Code))
	msg.Set(m.Fields().ByName("message"), protoreflect.ValueOfString(rslt.Message))

	if len(rslt.Fields) > 0 {
		fields := m.Fields().ByName("fields")

		fieldsList := msg.NewField(fields).List()
		for _, f := range rslt.Fields {
			field := pf.Messages().ByName("Field")
			fieldMessage := dynamicpb.NewMessage(field)
			fieldMessage.Set(field.Fields().ByName("name"), protoreflect.ValueOfString(f.Name))

			msgList := fieldMessage.NewField(field.Fields().ByName("message")).List()
			for _, item := range f.Message {
				msgList.Append(protoreflect.ValueOfString(item))
			}
			fieldMessage.Set(field.Fields().ByName("message"), protoreflect.ValueOf(msgList))

			fieldsList.Append(protoreflect.ValueOf(fieldMessage))
		}

		msg.Set(fields, protoreflect.ValueOf(fieldsList))
	}

	return msg
}
