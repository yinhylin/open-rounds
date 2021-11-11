// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.17.3
// source: proto/app.proto

package pb

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

type Actions_Event int32

const (
	Actions_NONE       Actions_Event = 0
	Actions_MOVE_UP    Actions_Event = 1
	Actions_MOVE_DOWN  Actions_Event = 2
	Actions_MOVE_LEFT  Actions_Event = 3
	Actions_MOVE_RIGHT Actions_Event = 4
)

// Enum value maps for Actions_Event.
var (
	Actions_Event_name = map[int32]string{
		0: "NONE",
		1: "MOVE_UP",
		2: "MOVE_DOWN",
		3: "MOVE_LEFT",
		4: "MOVE_RIGHT",
	}
	Actions_Event_value = map[string]int32{
		"NONE":       0,
		"MOVE_UP":    1,
		"MOVE_DOWN":  2,
		"MOVE_LEFT":  3,
		"MOVE_RIGHT": 4,
	}
)

func (x Actions_Event) Enum() *Actions_Event {
	p := new(Actions_Event)
	*p = x
	return p
}

func (x Actions_Event) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Actions_Event) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_app_proto_enumTypes[0].Descriptor()
}

func (Actions_Event) Type() protoreflect.EnumType {
	return &file_proto_app_proto_enumTypes[0]
}

func (x Actions_Event) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Actions_Event.Descriptor instead.
func (Actions_Event) EnumDescriptor() ([]byte, []int) {
	return file_proto_app_proto_rawDescGZIP(), []int{3, 0}
}

type Vector struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	X float64 `protobuf:"fixed64,1,opt,name=x,proto3" json:"x,omitempty"`
	Y float64 `protobuf:"fixed64,2,opt,name=y,proto3" json:"y,omitempty"`
}

func (x *Vector) Reset() {
	*x = Vector{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_app_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Vector) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Vector) ProtoMessage() {}

func (x *Vector) ProtoReflect() protoreflect.Message {
	mi := &file_proto_app_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Vector.ProtoReflect.Descriptor instead.
func (*Vector) Descriptor() ([]byte, []int) {
	return file_proto_app_proto_rawDescGZIP(), []int{0}
}

func (x *Vector) GetX() float64 {
	if x != nil {
		return x.X
	}
	return 0
}

func (x *Vector) GetY() float64 {
	if x != nil {
		return x.Y
	}
	return 0
}

type Entity struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Velocity *Vector  `protobuf:"bytes,2,opt,name=velocity,proto3" json:"velocity,omitempty"`
	Position *Vector  `protobuf:"bytes,3,opt,name=position,proto3" json:"position,omitempty"`
	Actions  *Actions `protobuf:"bytes,4,opt,name=actions,proto3" json:"actions,omitempty"`
}

func (x *Entity) Reset() {
	*x = Entity{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_app_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Entity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Entity) ProtoMessage() {}

func (x *Entity) ProtoReflect() protoreflect.Message {
	mi := &file_proto_app_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Entity.ProtoReflect.Descriptor instead.
func (*Entity) Descriptor() ([]byte, []int) {
	return file_proto_app_proto_rawDescGZIP(), []int{1}
}

func (x *Entity) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Entity) GetVelocity() *Vector {
	if x != nil {
		return x.Velocity
	}
	return nil
}

func (x *Entity) GetPosition() *Vector {
	if x != nil {
		return x.Position
	}
	return nil
}

func (x *Entity) GetActions() *Actions {
	if x != nil {
		return x.Actions
	}
	return nil
}

type States struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	States []*Entity `protobuf:"bytes,1,rep,name=states,proto3" json:"states,omitempty"`
}

func (x *States) Reset() {
	*x = States{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_app_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *States) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*States) ProtoMessage() {}

func (x *States) ProtoReflect() protoreflect.Message {
	mi := &file_proto_app_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use States.ProtoReflect.Descriptor instead.
func (*States) Descriptor() ([]byte, []int) {
	return file_proto_app_proto_rawDescGZIP(), []int{2}
}

func (x *States) GetStates() []*Entity {
	if x != nil {
		return x.States
	}
	return nil
}

type Actions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Actions []Actions_Event `protobuf:"varint,1,rep,packed,name=actions,proto3,enum=Actions_Event" json:"actions,omitempty"`
}

func (x *Actions) Reset() {
	*x = Actions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_app_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Actions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Actions) ProtoMessage() {}

func (x *Actions) ProtoReflect() protoreflect.Message {
	mi := &file_proto_app_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Actions.ProtoReflect.Descriptor instead.
func (*Actions) Descriptor() ([]byte, []int) {
	return file_proto_app_proto_rawDescGZIP(), []int{3}
}

func (x *Actions) GetActions() []Actions_Event {
	if x != nil {
		return x.Actions
	}
	return nil
}

type EntityEvents struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Actions *Actions `protobuf:"bytes,2,opt,name=actions,proto3" json:"actions,omitempty"`
}

func (x *EntityEvents) Reset() {
	*x = EntityEvents{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_app_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EntityEvents) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EntityEvents) ProtoMessage() {}

func (x *EntityEvents) ProtoReflect() protoreflect.Message {
	mi := &file_proto_app_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EntityEvents.ProtoReflect.Descriptor instead.
func (*EntityEvents) Descriptor() ([]byte, []int) {
	return file_proto_app_proto_rawDescGZIP(), []int{4}
}

func (x *EntityEvents) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *EntityEvents) GetActions() *Actions {
	if x != nil {
		return x.Actions
	}
	return nil
}

type Connect struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Connect) Reset() {
	*x = Connect{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_app_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Connect) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Connect) ProtoMessage() {}

func (x *Connect) ProtoReflect() protoreflect.Message {
	mi := &file_proto_app_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Connect.ProtoReflect.Descriptor instead.
func (*Connect) Descriptor() ([]byte, []int) {
	return file_proto_app_proto_rawDescGZIP(), []int{5}
}

type AddEntity struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Entity *Entity `protobuf:"bytes,1,opt,name=entity,proto3" json:"entity,omitempty"`
}

func (x *AddEntity) Reset() {
	*x = AddEntity{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_app_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddEntity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddEntity) ProtoMessage() {}

func (x *AddEntity) ProtoReflect() protoreflect.Message {
	mi := &file_proto_app_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddEntity.ProtoReflect.Descriptor instead.
func (*AddEntity) Descriptor() ([]byte, []int) {
	return file_proto_app_proto_rawDescGZIP(), []int{6}
}

func (x *AddEntity) GetEntity() *Entity {
	if x != nil {
		return x.Entity
	}
	return nil
}

type RemoveEntity struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *RemoveEntity) Reset() {
	*x = RemoveEntity{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_app_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoveEntity) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveEntity) ProtoMessage() {}

func (x *RemoveEntity) ProtoReflect() protoreflect.Message {
	mi := &file_proto_app_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveEntity.ProtoReflect.Descriptor instead.
func (*RemoveEntity) Descriptor() ([]byte, []int) {
	return file_proto_app_proto_rawDescGZIP(), []int{7}
}

func (x *RemoveEntity) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type ClientEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id   string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Tick int64  `protobuf:"varint,2,opt,name=tick,proto3" json:"tick,omitempty"`
	// Types that are assignable to Event:
	//	*ClientEvent_Connect
	//	*ClientEvent_Actions
	Event isClientEvent_Event `protobuf_oneof:"event"`
}

func (x *ClientEvent) Reset() {
	*x = ClientEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_app_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClientEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClientEvent) ProtoMessage() {}

func (x *ClientEvent) ProtoReflect() protoreflect.Message {
	mi := &file_proto_app_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClientEvent.ProtoReflect.Descriptor instead.
func (*ClientEvent) Descriptor() ([]byte, []int) {
	return file_proto_app_proto_rawDescGZIP(), []int{8}
}

func (x *ClientEvent) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ClientEvent) GetTick() int64 {
	if x != nil {
		return x.Tick
	}
	return 0
}

func (m *ClientEvent) GetEvent() isClientEvent_Event {
	if m != nil {
		return m.Event
	}
	return nil
}

func (x *ClientEvent) GetConnect() *Connect {
	if x, ok := x.GetEvent().(*ClientEvent_Connect); ok {
		return x.Connect
	}
	return nil
}

func (x *ClientEvent) GetActions() *Actions {
	if x, ok := x.GetEvent().(*ClientEvent_Actions); ok {
		return x.Actions
	}
	return nil
}

type isClientEvent_Event interface {
	isClientEvent_Event()
}

type ClientEvent_Connect struct {
	Connect *Connect `protobuf:"bytes,3,opt,name=connect,proto3,oneof"`
}

type ClientEvent_Actions struct {
	Actions *Actions `protobuf:"bytes,4,opt,name=actions,proto3,oneof"`
}

func (*ClientEvent_Connect) isClientEvent_Event() {}

func (*ClientEvent_Actions) isClientEvent_Event() {}

type ServerEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Tick int64 `protobuf:"varint,1,opt,name=tick,proto3" json:"tick,omitempty"`
	// Types that are assignable to Event:
	//	*ServerEvent_AddEntity
	//	*ServerEvent_RemoveEntity
	//	*ServerEvent_EntityEvents
	//	*ServerEvent_States
	Event isServerEvent_Event `protobuf_oneof:"event"`
}

func (x *ServerEvent) Reset() {
	*x = ServerEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_app_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServerEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServerEvent) ProtoMessage() {}

func (x *ServerEvent) ProtoReflect() protoreflect.Message {
	mi := &file_proto_app_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServerEvent.ProtoReflect.Descriptor instead.
func (*ServerEvent) Descriptor() ([]byte, []int) {
	return file_proto_app_proto_rawDescGZIP(), []int{9}
}

func (x *ServerEvent) GetTick() int64 {
	if x != nil {
		return x.Tick
	}
	return 0
}

func (m *ServerEvent) GetEvent() isServerEvent_Event {
	if m != nil {
		return m.Event
	}
	return nil
}

func (x *ServerEvent) GetAddEntity() *AddEntity {
	if x, ok := x.GetEvent().(*ServerEvent_AddEntity); ok {
		return x.AddEntity
	}
	return nil
}

func (x *ServerEvent) GetRemoveEntity() *RemoveEntity {
	if x, ok := x.GetEvent().(*ServerEvent_RemoveEntity); ok {
		return x.RemoveEntity
	}
	return nil
}

func (x *ServerEvent) GetEntityEvents() *EntityEvents {
	if x, ok := x.GetEvent().(*ServerEvent_EntityEvents); ok {
		return x.EntityEvents
	}
	return nil
}

func (x *ServerEvent) GetStates() *States {
	if x, ok := x.GetEvent().(*ServerEvent_States); ok {
		return x.States
	}
	return nil
}

type isServerEvent_Event interface {
	isServerEvent_Event()
}

type ServerEvent_AddEntity struct {
	AddEntity *AddEntity `protobuf:"bytes,2,opt,name=add_entity,json=addEntity,proto3,oneof"`
}

type ServerEvent_RemoveEntity struct {
	RemoveEntity *RemoveEntity `protobuf:"bytes,3,opt,name=remove_entity,json=removeEntity,proto3,oneof"`
}

type ServerEvent_EntityEvents struct {
	EntityEvents *EntityEvents `protobuf:"bytes,4,opt,name=entity_events,json=entityEvents,proto3,oneof"`
}

type ServerEvent_States struct {
	States *States `protobuf:"bytes,5,opt,name=states,proto3,oneof"`
}

func (*ServerEvent_AddEntity) isServerEvent_Event() {}

func (*ServerEvent_RemoveEntity) isServerEvent_Event() {}

func (*ServerEvent_EntityEvents) isServerEvent_Event() {}

func (*ServerEvent_States) isServerEvent_Event() {}

var File_proto_app_proto protoreflect.FileDescriptor

var file_proto_app_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x70, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0x24, 0x0a, 0x06, 0x56, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x0c, 0x0a, 0x01, 0x78,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x01, 0x52, 0x01, 0x78, 0x12, 0x0c, 0x0a, 0x01, 0x79, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x01, 0x52, 0x01, 0x79, 0x22, 0x86, 0x01, 0x0a, 0x06, 0x45, 0x6e, 0x74, 0x69,
	0x74, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x23, 0x0a, 0x08, 0x76, 0x65, 0x6c, 0x6f, 0x63, 0x69, 0x74, 0x79, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x56, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x52, 0x08, 0x76,
	0x65, 0x6c, 0x6f, 0x63, 0x69, 0x74, 0x79, 0x12, 0x23, 0x0a, 0x08, 0x70, 0x6f, 0x73, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x56, 0x65, 0x63, 0x74,
	0x6f, 0x72, 0x52, 0x08, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x22, 0x0a, 0x07,
	0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e,
	0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x07, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x22, 0x29, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x65, 0x73, 0x12, 0x1f, 0x0a, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x45, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x65, 0x73, 0x22, 0x81, 0x01, 0x0a, 0x07,
	0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x28, 0x0a, 0x07, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x0e, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x2e, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x07, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x22, 0x4c, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x08, 0x0a, 0x04, 0x4e, 0x4f,
	0x4e, 0x45, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x4d, 0x4f, 0x56, 0x45, 0x5f, 0x55, 0x50, 0x10,
	0x01, 0x12, 0x0d, 0x0a, 0x09, 0x4d, 0x4f, 0x56, 0x45, 0x5f, 0x44, 0x4f, 0x57, 0x4e, 0x10, 0x02,
	0x12, 0x0d, 0x0a, 0x09, 0x4d, 0x4f, 0x56, 0x45, 0x5f, 0x4c, 0x45, 0x46, 0x54, 0x10, 0x03, 0x12,
	0x0e, 0x0a, 0x0a, 0x4d, 0x4f, 0x56, 0x45, 0x5f, 0x52, 0x49, 0x47, 0x48, 0x54, 0x10, 0x04, 0x22,
	0x42, 0x0a, 0x0c, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x22, 0x0a, 0x07, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x08, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x07, 0x61, 0x63, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x22, 0x09, 0x0a, 0x07, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x22, 0x2c,
	0x0a, 0x09, 0x41, 0x64, 0x64, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x1f, 0x0a, 0x06, 0x65,
	0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x45, 0x6e,
	0x74, 0x69, 0x74, 0x79, 0x52, 0x06, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x22, 0x1e, 0x0a, 0x0c,
	0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x86, 0x01, 0x0a,
	0x0b, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04,
	0x74, 0x69, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x74, 0x69, 0x63, 0x6b,
	0x12, 0x24, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x08, 0x2e, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x48, 0x00, 0x52, 0x07, 0x63,
	0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x12, 0x24, 0x0a, 0x07, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x73, 0x48, 0x00, 0x52, 0x07, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x42, 0x07, 0x0a, 0x05,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0xe6, 0x01, 0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x63, 0x6b, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x03, 0x52, 0x04, 0x74, 0x69, 0x63, 0x6b, 0x12, 0x2b, 0x0a, 0x0a, 0x61, 0x64, 0x64,
	0x5f, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e,
	0x41, 0x64, 0x64, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x48, 0x00, 0x52, 0x09, 0x61, 0x64, 0x64,
	0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x34, 0x0a, 0x0d, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65,
	0x5f, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e,
	0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x48, 0x00, 0x52, 0x0c,
	0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x34, 0x0a, 0x0d,
	0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x5f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x73, 0x48, 0x00, 0x52, 0x0c, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x73, 0x12, 0x21, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x65, 0x73, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x07, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x73, 0x48, 0x00, 0x52, 0x06, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x73, 0x42, 0x07, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x42, 0x06,
	0x5a, 0x04, 0x2e, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_app_proto_rawDescOnce sync.Once
	file_proto_app_proto_rawDescData = file_proto_app_proto_rawDesc
)

func file_proto_app_proto_rawDescGZIP() []byte {
	file_proto_app_proto_rawDescOnce.Do(func() {
		file_proto_app_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_app_proto_rawDescData)
	})
	return file_proto_app_proto_rawDescData
}

var file_proto_app_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_proto_app_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_proto_app_proto_goTypes = []interface{}{
	(Actions_Event)(0),   // 0: Actions.Event
	(*Vector)(nil),       // 1: Vector
	(*Entity)(nil),       // 2: Entity
	(*States)(nil),       // 3: States
	(*Actions)(nil),      // 4: Actions
	(*EntityEvents)(nil), // 5: EntityEvents
	(*Connect)(nil),      // 6: Connect
	(*AddEntity)(nil),    // 7: AddEntity
	(*RemoveEntity)(nil), // 8: RemoveEntity
	(*ClientEvent)(nil),  // 9: ClientEvent
	(*ServerEvent)(nil),  // 10: ServerEvent
}
var file_proto_app_proto_depIdxs = []int32{
	1,  // 0: Entity.velocity:type_name -> Vector
	1,  // 1: Entity.position:type_name -> Vector
	4,  // 2: Entity.actions:type_name -> Actions
	2,  // 3: States.states:type_name -> Entity
	0,  // 4: Actions.actions:type_name -> Actions.Event
	4,  // 5: EntityEvents.actions:type_name -> Actions
	2,  // 6: AddEntity.entity:type_name -> Entity
	6,  // 7: ClientEvent.connect:type_name -> Connect
	4,  // 8: ClientEvent.actions:type_name -> Actions
	7,  // 9: ServerEvent.add_entity:type_name -> AddEntity
	8,  // 10: ServerEvent.remove_entity:type_name -> RemoveEntity
	5,  // 11: ServerEvent.entity_events:type_name -> EntityEvents
	3,  // 12: ServerEvent.states:type_name -> States
	13, // [13:13] is the sub-list for method output_type
	13, // [13:13] is the sub-list for method input_type
	13, // [13:13] is the sub-list for extension type_name
	13, // [13:13] is the sub-list for extension extendee
	0,  // [0:13] is the sub-list for field type_name
}

func init() { file_proto_app_proto_init() }
func file_proto_app_proto_init() {
	if File_proto_app_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_app_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Vector); i {
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
		file_proto_app_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Entity); i {
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
		file_proto_app_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*States); i {
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
		file_proto_app_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Actions); i {
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
		file_proto_app_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EntityEvents); i {
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
		file_proto_app_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Connect); i {
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
		file_proto_app_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddEntity); i {
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
		file_proto_app_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RemoveEntity); i {
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
		file_proto_app_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClientEvent); i {
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
		file_proto_app_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServerEvent); i {
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
	file_proto_app_proto_msgTypes[8].OneofWrappers = []interface{}{
		(*ClientEvent_Connect)(nil),
		(*ClientEvent_Actions)(nil),
	}
	file_proto_app_proto_msgTypes[9].OneofWrappers = []interface{}{
		(*ServerEvent_AddEntity)(nil),
		(*ServerEvent_RemoveEntity)(nil),
		(*ServerEvent_EntityEvents)(nil),
		(*ServerEvent_States)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_app_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_app_proto_goTypes,
		DependencyIndexes: file_proto_app_proto_depIdxs,
		EnumInfos:         file_proto_app_proto_enumTypes,
		MessageInfos:      file_proto_app_proto_msgTypes,
	}.Build()
	File_proto_app_proto = out.File
	file_proto_app_proto_rawDesc = nil
	file_proto_app_proto_goTypes = nil
	file_proto_app_proto_depIdxs = nil
}
