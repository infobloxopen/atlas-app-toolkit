package resource_test

import (
	"fmt"

	"github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
	"github.com/infobloxopen/atlas-app-toolkit/gorm/resource/fqstring"
	"github.com/infobloxopen/atlas-app-toolkit/gorm/resource/uuid"
	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

type ExampleGoType struct {
	ID         *resource.Identifier
	ExternalID *resource.Identifier
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
	// register fqstring codec for default resources
	resource.RegisterCodec(fqstring.NewCodec(), nil)
	// register uuid codec for TestProtoMessage resources
	resource.RegisterCodec(uuid.NewCodec("simpleapp", "examples"), &ExampleProtoMessage{})

	// and you want to convert PB type to your application type
	toGoTypeFunc := func(msg *ExampleProtoMessage) (*ExampleGoType, error) {
		var v ExampleGoType

		// arbitrary variables
		v.VarName = msg.VarName

		// convert RPC identifier using UUID Codec for ExampleProtoMessage resource type
		if id, err := resource.Decode(msg, msg.Id); err != nil {
			return nil, err
		} else {
			v.ID = id
		}
		// convert RPC identifier using External Codec for default resource type
		if id, err := resource.Decode(nil, msg.ExternalId); err != nil {
			return nil, err
		} else {
			v.ExternalID = id
		}
		return &v, nil
	}

	// let's create PB message
	pb := &ExampleProtoMessage{
		Id: &resourcepb.Identifier{
			ApplicationName: "simpleapp",
			ResourceType:    "examples",
			ResourceId:      "00000000-0000-0000-0000-000000000000",
		},
		// ExternalId stores data about "external_resource" that belongs to
		// "externalapp" and has id "1"
		ExternalId: &resourcepb.Identifier{
			ApplicationName: "externalapp",
			ResourceType:    "external_resource",
			ResourceId:      "1",
		},
		VarName: "somename",
	}

	val, err := toGoTypeFunc(pb)
	if err != nil {
		fmt.Printf("failed to convert TestProtoMessage to TestGoType: %s\n", err)
		return
	}

	//fmt.Printf("application name of internal id: %s\n", val.ID.ResourceID)
	fmt.Printf("application name of fqstring id: %s\n", val.ExternalID.ResourceID)

	// so now you want to convert it back to PB representation
	toPBMessageFunc := func(v *ExampleGoType) (*ExampleProtoMessage, error) {
		var pb ExampleProtoMessage

		// arbitrary variables
		pb.VarName = v.VarName

		// convert internal id to RPC representation using registered UUID codec
		if id, err := resource.Encode(pb, v.ID); err != nil {
			return nil, err
		} else {
			pb.Id = id
		}

		// convert fqstring id to RPC representation using registered External codec
		if id, err := resource.Encode(nil, v.ExternalID); err != nil {
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
	fmt.Printf("resource type of internal id: %s\n", pb.Id.GetResourceType())
	//fmt.Printf("resource id of internal id: %s\n", pb.Id.GetResourceId())
	fmt.Printf("application name of fqstring id: %s\n", pb.ExternalId.GetApplicationName())
	fmt.Printf("resource type of fqstring id: %s\n", pb.ExternalId.GetResourceType())
	fmt.Printf("resource id of fqstring id: %s\n", pb.ExternalId.GetResourceId())

	// Output:
	//application name of fqstring id: externalapp/external_resource/1
	//application name of internal id: simpleapp
	//resource type of internal id: examples
	//application name of fqstring id: externalapp
	//resource type of fqstring id: external_resource
	//resource id of fqstring id: 1
}
