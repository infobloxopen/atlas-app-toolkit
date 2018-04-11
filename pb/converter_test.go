package pb

import (
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"reflect"
	"testing"
)

var testObjectWBR = WBRRsrc{
	OphId:          "abc123",
	Name:           "wbr1",
	SerialNumber:   "idk",
	Location:       "tacoma",
	BgpPassword:    &wrappers.StringValue{"p@$$word"},
	BgpAsn:         &wrappers.UInt32Value{12345},
	AggregateRoute: &wrappers.StringValue{"127.0.0.1/32"},
	Network:        &wrappers.StringValue{"10.0.0.0/24"},
	Description:    "nothing useful",
	Interfaces: []*InterfaceRsrc{
		{
			Name:      "eth0",
			IpAddress: "10.0.0.2",
			Type:      "link-layer",
			// GreKey:
			// LlInterfaceName:
			// Ttl:
		},
		{
			Name:            "tunl_gre0",
			IpAddress:       "10.0.0.3",
			Type:            "tunnel",
			GreKey:          &wrappers.StringValue{"1"},
			LlInterfaceName: &wrappers.StringValue{"eth0"},
			Ttl:             &wrappers.UInt32Value{255},
		},
	},
}
var result *WBR

func BenchmarkServiceObjectToGORM(b *testing.B) {
	var x *WBR
	for n := 0; n < b.N; n++ {
		x = &WBR{}
		Convert(testObjectWBR, x)
	}
	result = x
}

var resultBloat *WBRWithBloat

// A straight iterative approach takes longer when there are case mismatched
// fields later in the struct
func BenchmarkBloatedServiceObjectToGORM(b *testing.B) {
	var x *WBRWithBloat
	for n := 0; n < b.N; n++ {
		x = &WBRWithBloat{}
		Convert(testObjectWBR, x)
	}
	resultBloat = x
}

func BenchmarkWBRRsrcToGORM(b *testing.B) {
	var x *WBR
	for n := 0; n < b.N; n++ {
		x = WBRRsrcToGORM(testObjectWBR)
	}
	result = x
}

var resultI *Interface

func BenchmarkServiceObjectToGORMInterface(b *testing.B) {
	var x *Interface
	for n := 0; n < b.N; n++ {
		x = &Interface{}
		Convert(testObjectWBR.Interfaces[1], x)
	}
	resultI = x
}

func BenchmarkWBRRsrcToGORMInterface(b *testing.B) {
	var x *Interface
	for n := 0; n < b.N; n++ {
		x = InterfaceRsrcToGORM(*testObjectWBR.Interfaces[1])
	}
	resultI = x
}

type mockEnum int32

type demoPB struct {
	Num               uint32
	Str               string
	MaybeNum          *wrappers.UInt32Value
	MaybeStr          *wrappers.StringValue
	TestEnum          mockEnum
	Casecheck         string
	ExtraneousPBField string
}

type demoGORM struct {
	Num                 uint32
	Str                 string
	MaybeNum            *uint32
	MaybeStr            *string
	TestEnum            int32
	CaseCheck           string
	ExtraneousGormField string
}

func pInt(num uint32) *uint32 {
	x := num
	return &x
}
func pString(str string) *string {
	x := str
	return &x
}

func TestConvertServiceGORMObjects(t *testing.T) {
	testCases := map[*demoPB]demoGORM{
		&demoPB{
			Num:      1,
			Str:      "nothing",
			TestEnum: 0,
		}: demoGORM{
			Num:      1,
			Str:      "nothing",
			MaybeNum: nil,
			MaybeStr: nil,
			TestEnum: 0,
		},

		&demoPB{
			Num:      3,
			MaybeNum: &wrappers.UInt32Value{Value: 5},
			MaybeStr: &wrappers.StringValue{Value: "Message"},
			TestEnum: 2,
		}: demoGORM{
			Num:      3,
			Str:      "",
			MaybeNum: pInt(5),
			MaybeStr: pString("Message"),
			TestEnum: 2,
		},

		&demoPB{
			Casecheck: "arbitrary",
		}: demoGORM{
			Num:       0,
			CaseCheck: "arbitrary",
		},
	}
	for input, goal := range testCases {
		to := &demoGORM{}
		err := Convert(input, to)
		if err != nil {
			t.Log(fmt.Sprintf("Conversion from %+v to %+v failed", *input, *to))
			t.Fail()
		}
		if !reflect.DeepEqual(*to, goal) {
			t.Log(fmt.Sprintf("Expected %+v, got %+v", goal, *to))
			t.Fail()
		}
	}
	// Should fail
	err := Convert(&demoPB{}, demoGORM{})
	if err == nil {
		t.Log("Should throw error copying to non-pointer object")
		t.Fail()
	}
}
