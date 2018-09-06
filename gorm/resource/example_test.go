package resource

import (
	"fmt"

	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
	"github.com/infobloxopen/atlas-app-toolkit/util"
)

type ExampleGoType struct {
	ID         int64
	ExternalID string
	VarName    string
}

type ExampleProtoMessage struct {
	Id         *resourcepb.Identifier
	ExternalId *resourcepb.Identifier
	VarName    string
}

func (ExampleProtoMessage) XXX_MessageName() string { return "ExampleProtoMessage" }
func (ExampleProtoMessage) Reset()                  {}
func (ExampleProtoMessage) String() string          { return "ExampleProtoMessage" }
func (ExampleProtoMessage) ProtoMessage()           {}

func Example() {
	RegisterApplication("simpleapp")

	// you want to convert PB type to your application type
	toGoTypeFunc := func(msg *ExampleProtoMessage) (*ExampleGoType, error) {
		var v ExampleGoType

		// arbitrary variables
		v.VarName = msg.VarName

		// convert RPC identifier using UUID Codec for ExampleProtoMessage resource type
		if id, err := DecodeInt64(msg, msg.Id); err != nil {
			return nil, err
		} else {
			v.ID = id
		}
		// convert RPC identifier using External Codec for default resource type
		if id, err := Decode(nil, msg.ExternalId); err != nil {
			return nil, err
		} else {
			v.ExternalID = id.(string)
		}
		return &v, nil
	}

	// let's create PB message
	pb := &ExampleProtoMessage{
		Id: &resourcepb.Identifier{
			ApplicationName: "simpleapp",
			ResourceType:    "example_proto_message",
			ResourceId:      "12",
		},
		// ExternalId stores data about "external_resource" that belongs to
		// "externalapp" and has id "id"
		ExternalId: &resourcepb.Identifier{
			ApplicationName: "externalapp",
			ResourceType:    "external_resource",
			ResourceId:      "id",
		},
		VarName: "somename",
	}

	val, err := toGoTypeFunc(pb)
	if err != nil {
		fmt.Printf("failed to convert TestProtoMessage to TestGoType: %s\n", err)
		return
	}

	fmt.Printf("application name of integer id: %v\n", val.ID)
	fmt.Printf("application name of fqstring id: %v\n", val.ExternalID)

	// so now you want to convert it back to PB representation
	toPBMessageFunc := func(v *ExampleGoType) (*ExampleProtoMessage, error) {
		var pb ExampleProtoMessage

		// arbitrary variables
		pb.VarName = v.VarName

		// convert internal id to RPC representation using registered UUID codec
		if id, err := Encode(pb, v.ID); err != nil {
			return nil, err
		} else {
			pb.Id = id
		}

		// convert fqstring id to RPC representation using registered External codec
		if id, err := Encode(nil, v.ExternalID); err != nil {
			return nil, err
		} else {
			pb.ExternalId = id
		}

		return &pb, nil
	}

	pb, err = toPBMessageFunc(val)
	if err != nil {
		fmt.Println("failed to convert TestGoType to TestProtoMessage")
	}

	fmt.Printf("application name of internal id: %s\n", pb.Id.GetApplicationName())
	fmt.Printf("resource type of internal id: %s\n", util.CamelToSnake(pb.Id.GetResourceType()))
	fmt.Printf("resource id of internal id: %s\n", pb.Id.GetResourceId())
	fmt.Printf("application name of fqstring id: %s\n", pb.ExternalId.GetApplicationName())
	fmt.Printf("resource type of fqstring id: %s\n", pb.ExternalId.GetResourceType())
	fmt.Printf("resource id of fqstring id: %s\n", pb.ExternalId.GetResourceId())

	// Output:
	//application name of integer id: 12
	//application name of fqstring id: externalapp/external_resource/id
	//application name of internal id: simpleapp
	//resource type of internal id: example_proto_message
	//resource id of internal id: 12
	//application name of fqstring id: externalapp
	//resource type of fqstring id: external_resource
	//resource id of fqstring id: id
}
