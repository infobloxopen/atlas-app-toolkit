package resourceuuid

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/internal"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resourcepb"
)

func NewUUIDManager(applicationID, resourceType string) *UUIDManager {
	return &UUIDManager{
		applicationName: applicationID,
		resourceType:    resourceType,
	}
}

type UUIDManager struct {
	applicationName string
	resourceType    string
}

func (m UUIDManager) String() string {
	return fmt.Sprintf("uuid manager of %s/%s", m.applicationName, m.resourceType)
}

func (m UUIDManager) Build(id *resourcepb.Identifier) (Identifier, error) {
	var uid uuidIdentifier

	if id == nil || (id.ApplicationName == "" && id.ResourceType == "" && id.ResourceId == "") {
		return &uid, nil
	}

	if id.ApplicationName != "" && id.ApplicationName != m.applicationName {
		return nil, fmt.Errorf("resource: invalid application name %s in %s", id.ApplicationName, m)
	}
	uid.applicationName = id.ApplicationName

	if id.ResourceType != "" && id.ResourceType != m.resourceType {
		return nil, fmt.Errorf("resource: invalid resource type %s in %s", id.ResourceType, m)
	}
	uid.resourceType = id.ResourceType

	if id.ResourceId == "" {
		v, err := uuid.NewUUID()
		if err != nil {
			return nil, err
		}
		uid.resourceID = v.String()
	} else {
		uid.resourceID = id.ResourceId
	}

	return &uid, nil
}

func (m UUIDManager) Resolve(id Identifier) (*resourcepb.Identifier, error) {
	var pbid resourcepb.Identifier

	uid, ok := id.(*uuidIdentifier)
	if !ok {
		return nil, fmt.Errorf("resource: invalid type of Identifier in %s", m)
	}

	if !uid.valid || uid.resourceID == "" {
		return &pbid, nil
	}

	if uid.applicationName != "" && uid.applicationName != m.applicationName {
		return nil, fmt.Errorf("resource: invalid application name %s in %s", uid.applicationName, m)
	} else {
		uid.applicationName = m.applicationName
	}
	pbid.ApplicationName = uid.applicationName

	if uid.resourceType != "" && uid.resourceType != m.resourceType {
		return nil, fmt.Errorf("resource: invalid resource type %s in %s", uid.resourceType, m)
	} else {
		uid.resourceType = m.resourceType
	}
	pbid.ResourceType = uid.resourceType

	v, err := uuid.Parse(uid.resourceID)
	if err != nil {
		return nil, fmt.Errorf("resource: invalid resource id %s in %s", uid.resourceID, m)
	}
	pbid.ResourceId = v.String()

	return &pbid, nil
}

type uuidIdentifier struct {
	applicationName string
	resourceType    string
	resourceID      string
	valid           bool
}

func (i uuidIdentifier) String() string {
	return internal.BuildString(i.applicationName, i.resourceType, i.resourceID)
}

func (i uuidIdentifier) Value() (driver.Value, error) {
	if !i.valid {
		return nil, nil
	}
	return i.resourceID, nil
}

func (i *uuidIdentifier) Scan(v interface{}) error {
	if v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("uuid identifier: invalid type of resource id %T", v)
	}
	i.resourceID = s
	i.valid = true

	return nil
}

func (i uuidIdentifier) ApplicationName() string {
	return i.applicationName
}

func (i uuidIdentifier) ResourceType() string {
	return i.resourceType
}

func (i uuidIdentifier) ResourceID() string {
	return i.resourceID
}

func (i uuidIdentifier) IsExternal() bool {
	return false
}

func (i uuidIdentifier) IsNil() bool {
	return !i.valid
}
