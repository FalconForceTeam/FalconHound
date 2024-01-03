package internal

import (
	"encoding/json"
)

// This file was created based on the following two sources:
// https://github.com/SpecterOps/BloodHound/blob/b3fb3b79334fbb3d4ad8c5f025c89b7f8a92c4da/packages/go/ein/incoming_models.go#L88
// https://github.com/SpecterOps/BloodHound/blob/b3fb3b79334fbb3d4ad8c5f025c89b7f8a92c4da/cmd/api/src/daemons/datapipe/models.go#L28

type DataWrapper struct {
	Metadata Metadata        `json:"meta"`
	Payload  json.RawMessage `json:"data"`
}

type Metadata struct {
	Type    DataType         `json:"type"`
	Methods CollectionMethod `json:"methods"`
	Version int              `json:"version"`
}

type DataType string

const (
	DataTypeSession     DataType = "sessions"
	DataTypeUser        DataType = "users"
	DataTypeGroup       DataType = "groups"
	DataTypeComputer    DataType = "computers"
	DataTypeGPO         DataType = "gpos"
	DataTypeOU          DataType = "ous"
	DataTypeDomain      DataType = "domains"
	DataTypeRemoved     DataType = "deleted"
	DataTypeContainer   DataType = "containers"
	DataTypeLocalGroups DataType = "localgroups"
	DataTypeAzure       DataType = "azure"
)

type CollectionMethod uint64

const (
	CollectionMethodGroup         CollectionMethod = 1
	CollectionMethodLocalAdmin    CollectionMethod = 1 << 1
	CollectionMethodGPOLocalGroup CollectionMethod = 1 << 2
	CollectionMethodSession       CollectionMethod = 1 << 3
	CollectionMethodLoggedOn      CollectionMethod = 1 << 4
	CollectionMethodTrusts        CollectionMethod = 1 << 5
	CollectionMethodACL           CollectionMethod = 1 << 6
	CollectionMethodContainer     CollectionMethod = 1 << 7
	CollectionMethodRDP           CollectionMethod = 1 << 8
	CollectionMethodObjectProps   CollectionMethod = 1 << 9
	CollectionMethodSessionLoop   CollectionMethod = 1 << 10
	CollectionMethodLoggedOnLoop  CollectionMethod = 1 << 11
	CollectionMethodDCOM          CollectionMethod = 1 << 12
	CollectionMethodSPNTargets    CollectionMethod = 1 << 13
	CollectionMethodPSRemote      CollectionMethod = 1 << 14
)

type TypedPrincipal struct {
	ObjectIdentifier string
	ObjectType       string
}

type ACE struct {
	PrincipalSID  string
	PrincipalType string
	RightName     string
	IsInherited   bool
}

type SPNTarget struct {
	ComputerSID string
	Port        int
	Service     string
}

type IngestBase struct {
	ObjectIdentifier string
	Properties       map[string]any
	Aces             []ACE
	IsDeleted        bool
	IsACLProtected   bool
	ContainedBy      TypedPrincipal
}

type GPO IngestBase

type Session struct {
	ComputerSID string
	UserSID     string
	LogonType   int
}

type Group struct {
	IngestBase
	Members []TypedPrincipal
}

type User struct {
	IngestBase
	AllowedToDelegate []TypedPrincipal
	SPNTargets        []SPNTarget
	PrimaryGroupSID   string
	HasSIDHistory     []TypedPrincipal
}

type Container struct {
	IngestBase
	ChildObjects []TypedPrincipal
}

type Trust struct {
	TargetDomainSid     string
	IsTransitive        bool
	TrustDirection      string
	TrustType           string
	SidFilteringEnabled bool
	TargetDomainName    string
}

type GPLink struct {
	Guid       string
	IsEnforced bool
}

type Domain struct {
	IngestBase
	ChildObjects []TypedPrincipal
	Trusts       []Trust
	Links        []GPLink
}

type SessionAPIResult struct {
	APIResult
	Results []Session
}

type ComputerStatus struct {
	Connectable bool
	Error       string
}

type APIResult struct {
	Collected     bool
	FailureReason string
}

type NamedPrincipal struct {
	ObjectIdentifier string
	PrincipalName    string
}

type LocalGroupAPIResult struct {
	APIResult
	Results          []TypedPrincipal
	LocalNames       []NamedPrincipal
	Name             string
	ObjectIdentifier string
}

type UserRightsAssignmentAPIResult struct {
	APIResult
	Results    []TypedPrincipal
	LocalNames []NamedPrincipal
	Privilege  string
}

type Computer struct {
	IngestBase
	PrimaryGroupSID    string
	AllowedToDelegate  []TypedPrincipal
	AllowedToAct       []TypedPrincipal
	DumpSMSAPassword   []TypedPrincipal
	Sessions           SessionAPIResult
	PrivilegedSessions SessionAPIResult
	RegistrySessions   SessionAPIResult
	LocalGroups        []LocalGroupAPIResult
	UserRights         []UserRightsAssignmentAPIResult
	Status             ComputerStatus
	HasSIDHistory      []TypedPrincipal
	Properties         map[string]any
	ObjectIdentifier   string
}

type OU struct {
	IngestBase
	ChildObjects []TypedPrincipal
	Links        []GPLink
}
