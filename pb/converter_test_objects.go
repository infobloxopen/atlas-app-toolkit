package pb

import google_protobuf1 "github.com/golang/protobuf/ptypes/wrappers"
import "time"

// InterfaceRsrc pulled from a protobuf generated file
type InterfaceRsrc struct {
	Name            string                        `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
	IpAddress       string                        `protobuf:"bytes,2,opt,name=ip_address,json=ipAddress" json:"ip_address,omitempty"`
	Type            string                        `protobuf:"bytes,3,opt,name=type" json:"type,omitempty"`
	GreKey          *google_protobuf1.StringValue `protobuf:"bytes,4,opt,name=gre_key,json=greKey" json:"gre_key,omitempty"`
	LlInterfaceName *google_protobuf1.StringValue `protobuf:"bytes,5,opt,name=ll_interface_name,json=llInterfaceName" json:"ll_interface_name,omitempty"`
	Ttl             *google_protobuf1.UInt32Value `protobuf:"bytes,6,opt,name=ttl" json:"ttl,omitempty"`
}

// WBRRsrc pulled from a protobuf generated file
type WBRRsrc struct {
	Id             uint32                        `protobuf:"varint,11,opt,name=id" json:"id,omitempty"`
	OphId          string                        `protobuf:"bytes,1,opt,name=oph_id,json=ophId" json:"oph_id,omitempty"`
	Name           string                        `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	SerialNumber   string                        `protobuf:"bytes,3,opt,name=serial_number,json=serialNumber" json:"serial_number,omitempty"`
	Location       string                        `protobuf:"bytes,4,opt,name=location" json:"location,omitempty"`
	BgpPassword    *google_protobuf1.StringValue `protobuf:"bytes,5,opt,name=bgp_password,json=bgpPassword" json:"bgp_password,omitempty"`
	BgpAsn         *google_protobuf1.UInt32Value `protobuf:"bytes,6,opt,name=bgp_asn,json=bgpAsn" json:"bgp_asn,omitempty"`
	AggregateRoute *google_protobuf1.StringValue `protobuf:"bytes,7,opt,name=aggregate_route,json=aggregateRoute" json:"aggregate_route,omitempty"`
	Network        *google_protobuf1.StringValue `protobuf:"bytes,8,opt,name=network" json:"network,omitempty"`
	Description    string                        `protobuf:"bytes,9,opt,name=description" json:"description,omitempty"`
	Interfaces     []*InterfaceRsrc              `protobuf:"bytes,10,rep,name=interfaces" json:"interfaces,omitempty"`
}

// Interface is the gorm compatible version of InterfaceRsrc from above
type Interface struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time

	WBRID           uint
	Name            string
	IPAddress       string
	Type            string
	GreKey          *string
	LlInterfaceName *string
	TTL             *uint32
}

// WBR is the gorm compatible version of WBRRsrc from above
type WBR struct {
	ID        uint32
	CreatedAt time.Time
	UpdatedAt time.Time

	OphID          string
	Name           string
	SerialNumber   string
	Location       string
	BgpPassword    *string
	BgpAsn         *uint32
	AggregateRoute *string
	Network        *string
	Description    string
	Interfaces     []Interface
}

// WBRWithBloat also has a number of unused fields before the last case mismatch
type WBRWithBloat struct {
	ID        uint32
	CreatedAt time.Time
	UpdatedAt time.Time

	Name           string
	SerialNumber   string
	Location       string
	BgpPassword    *string
	BgpAsn         *uint32
	AggregateRoute *string
	Network        *string
	Description    string
	Interfaces     []Interface

	Bloat1  string
	Bloat2  string
	Bloat3  string
	Bloat4  string
	Bloat5  string
	Bloat6  string
	Bloat7  string
	Bloat8  string
	Bloat9  string
	Bloat10 string
	Bloat11 string
	Bloat12 string

	OphID string
}

func unwrapInt(num *google_protobuf1.UInt32Value) *uint32 {
	if num != nil {
		value := num.Value
		return &value
	}
	return nil
}

func unwrapString(str *google_protobuf1.StringValue) *string {
	if str != nil {
		value := str.Value
		return &value
	}
	return nil
}

// InterfaceRsrcToGORM ...
func InterfaceRsrcToGORM(iface InterfaceRsrc) *Interface {
	ifaceGORM := Interface{
		// WBRRsrcID       int
		// FK Field handled by GORM
		Name:            iface.Name,
		IPAddress:       iface.IpAddress,
		Type:            iface.Type,
		GreKey:          unwrapString(iface.GreKey),
		LlInterfaceName: unwrapString(iface.LlInterfaceName),
		TTL:             unwrapInt(iface.Ttl),
	}
	// Other validation that could be necessary goes in here

	return &ifaceGORM
}

// WBRRsrcToGORM ...
func WBRRsrcToGORM(wbr WBRRsrc) *WBR {
	wbrGORM := WBR{

		OphID:          wbr.OphId,
		Name:           wbr.Name,
		SerialNumber:   wbr.SerialNumber,
		Location:       wbr.Location,
		BgpPassword:    unwrapString(wbr.BgpPassword),
		BgpAsn:         unwrapInt(wbr.BgpAsn),
		AggregateRoute: unwrapString(wbr.AggregateRoute),
		Network:        unwrapString(wbr.Network),
		Description:    wbr.Description,
		// Interfaces     []*InterfaceRsrc
	}
	for _, iface := range wbr.Interfaces {
		wbrGORM.Interfaces = append(wbrGORM.Interfaces, *InterfaceRsrcToGORM(*iface))
	}
	return &wbrGORM
}
