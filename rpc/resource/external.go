package resource

import (
	"database/sql/driver"
	"fmt"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/internal"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resourcepb"
)

func NewExternalManager() ExternalManager {
	return ExternalManager{}
}

type ExternalManager struct{}

func (ExternalManager) Build(id *resourcepb.Identifier) (Identifier, error) {
	var eid externalIdentifier

	if id == nil || (id.ApplicationName == "" && id.ResourceType == "" && id.ResourceId == "") {
		return &eid, nil
	}

	if id.ApplicationName == "" || id.ResourceType == "" || id.ResourceId == "" {
		return nil, fmt.Errorf("resource: external identifier is not fully qualified - %s", id)
	}

	eid.applicationName, eid.resourceType, eid.resourceID = id.ApplicationName, id.ResourceType, id.ResourceId

	return &eid, nil
}

func (ExternalManager) Resolve(id Identifier) (*resourcepb.Identifier, error) {
	var pbid resourcepb.Identifier

	eid, ok := id.(*externalIdentifier)
	if !ok {
		return nil, fmt.Errorf("resource: invalid type of external identifier - %T", id)
	}

	if !eid.valid {
		return &pbid, nil
	}

	if eid.applicationName == "" || eid.resourceType == "" || eid.resourceID == "" {
		return nil, fmt.Errorf("resource: resolved external identifier is not fully qualified - %s", id)
	}
	pbid.ApplicationName, pbid.ResourceType, pbid.ResourceId = eid.applicationName, eid.resourceType, eid.resourceID

	return &pbid, nil
}

type externalIdentifier struct {
	applicationName string
	resourceType    string
	resourceID      string
	valid           bool
}

func (i externalIdentifier) String() string {
	return internal.BuildString(i.applicationName, i.resourceType, i.resourceID)
}

func (i externalIdentifier) Value() (driver.Value, error) {
	if !i.valid {
		return nil, nil
	}
	return i.String(), nil
}

func (i *externalIdentifier) Scan(v interface{}) error {
	if v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("external identifier: invalid type of resource id %T", v)
	}
	i.applicationName, i.resourceType, i.resourceID = internal.ParseString(s)
	i.valid = true

	return nil
}

func (i externalIdentifier) ApplicationName() string {
	return i.applicationName
}

func (i externalIdentifier) ResourceType() string {
	return i.resourceType
}

func (i externalIdentifier) ResourceID() string {
	return i.resourceID
}

func (i externalIdentifier) IsExternal() bool {
	return false
}

func (i externalIdentifier) IsNil() bool {
	return !i.valid
}
