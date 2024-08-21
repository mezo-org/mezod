// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: mezo/bridge/v1/vote_extension.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// VoteExtension defines the vote extension structure for the Bitcoin bridge.
type VoteExtension struct {
	AssetsLockedEvents []*AssetsLockedEvent `protobuf:"bytes,1,rep,name=assets_locked_events,json=assetsLockedEvents,proto3" json:"assets_locked_events,omitempty"`
}

func (m *VoteExtension) Reset()         { *m = VoteExtension{} }
func (m *VoteExtension) String() string { return proto.CompactTextString(m) }
func (*VoteExtension) ProtoMessage()    {}
func (*VoteExtension) Descriptor() ([]byte, []int) {
	return fileDescriptor_92f77ea72398eb8e, []int{0}
}
func (m *VoteExtension) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *VoteExtension) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_VoteExtension.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *VoteExtension) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VoteExtension.Merge(m, src)
}
func (m *VoteExtension) XXX_Size() int {
	return m.Size()
}
func (m *VoteExtension) XXX_DiscardUnknown() {
	xxx_messageInfo_VoteExtension.DiscardUnknown(m)
}

var xxx_messageInfo_VoteExtension proto.InternalMessageInfo

func (m *VoteExtension) GetAssetsLockedEvents() []*AssetsLockedEvent {
	if m != nil {
		return m.AssetsLockedEvents
	}
	return nil
}

// AssetsLockedEvent represents the event where inbound assets are locked in
// the Bitcoin bridge.
//
// TODO: Consider moving this proto to `proto/mezo/types/v1` package
//
//	with `github.com/mezo-org/mezod/types` as the go_package. This may
//	be useful as the Ethereum sidecar may also need to use this proto
//	as part of its gRPC interface.
type AssetsLockedEvent struct {
	// sequence is the unique identifier of the event.
	Sequence cosmossdk_io_math.Int `protobuf:"bytes,1,opt,name=sequence,proto3,customtype=cosmossdk.io/math.Int" json:"sequence"`
	// recipient is the account address to receive the locked assets on Mezo,
	// in Bech32 format.
	Recipient string `protobuf:"bytes,2,opt,name=recipient,proto3" json:"recipient,omitempty"`
	// amount of assets locked, in 1e18 precision.
	Amount cosmossdk_io_math.Int `protobuf:"bytes,3,opt,name=amount,proto3,customtype=cosmossdk.io/math.Int" json:"amount"`
}

func (m *AssetsLockedEvent) Reset()         { *m = AssetsLockedEvent{} }
func (m *AssetsLockedEvent) String() string { return proto.CompactTextString(m) }
func (*AssetsLockedEvent) ProtoMessage()    {}
func (*AssetsLockedEvent) Descriptor() ([]byte, []int) {
	return fileDescriptor_92f77ea72398eb8e, []int{1}
}
func (m *AssetsLockedEvent) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *AssetsLockedEvent) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_AssetsLockedEvent.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *AssetsLockedEvent) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AssetsLockedEvent.Merge(m, src)
}
func (m *AssetsLockedEvent) XXX_Size() int {
	return m.Size()
}
func (m *AssetsLockedEvent) XXX_DiscardUnknown() {
	xxx_messageInfo_AssetsLockedEvent.DiscardUnknown(m)
}

var xxx_messageInfo_AssetsLockedEvent proto.InternalMessageInfo

func (m *AssetsLockedEvent) GetRecipient() string {
	if m != nil {
		return m.Recipient
	}
	return ""
}

func init() {
	proto.RegisterType((*VoteExtension)(nil), "mezo.bridge.v1.VoteExtension")
	proto.RegisterType((*AssetsLockedEvent)(nil), "mezo.bridge.v1.AssetsLockedEvent")
}

func init() {
	proto.RegisterFile("mezo/bridge/v1/vote_extension.proto", fileDescriptor_92f77ea72398eb8e)
}

var fileDescriptor_92f77ea72398eb8e = []byte{
	// 318 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0xd1, 0xcd, 0x4a, 0x03, 0x31,
	0x10, 0x07, 0xf0, 0x8d, 0x85, 0x62, 0x23, 0x0a, 0x2e, 0x15, 0xd6, 0xa2, 0xdb, 0x5a, 0x2f, 0xbd,
	0x34, 0xa1, 0x8a, 0x07, 0x8f, 0x16, 0x8a, 0x08, 0x9e, 0x2a, 0x78, 0xf0, 0xb2, 0xec, 0xc7, 0xb0,
	0x0d, 0x75, 0x33, 0x75, 0x93, 0x2e, 0xd5, 0xa7, 0xf0, 0x1d, 0x7c, 0x99, 0x1e, 0x7b, 0x14, 0x0f,
	0x45, 0xba, 0x2f, 0x22, 0xbb, 0xb1, 0x8a, 0x7a, 0xf1, 0x36, 0xf9, 0xe7, 0x97, 0x49, 0x98, 0xd0,
	0xe3, 0x04, 0x9e, 0x90, 0x07, 0xa9, 0x88, 0x62, 0xe0, 0x59, 0x8f, 0x67, 0xa8, 0xc1, 0x83, 0x99,
	0x06, 0xa9, 0x04, 0x4a, 0x36, 0x49, 0x51, 0xa3, 0xbd, 0x53, 0x20, 0x66, 0x10, 0xcb, 0x7a, 0x8d,
	0xfd, 0x10, 0x55, 0x82, 0xca, 0x2b, 0x77, 0xb9, 0x59, 0x18, 0xda, 0xa8, 0xc7, 0x18, 0xa3, 0xc9,
	0x8b, 0xca, 0xa4, 0xed, 0x88, 0x6e, 0xdf, 0xa2, 0x86, 0xc1, 0xba, 0xaf, 0x7d, 0x43, 0xeb, 0xbe,
	0x52, 0xa0, 0x95, 0x77, 0x8f, 0xe1, 0x18, 0x22, 0x0f, 0x32, 0x90, 0x5a, 0x39, 0xa4, 0x55, 0xe9,
	0x6c, 0x9d, 0x1c, 0xb1, 0x9f, 0x17, 0xb2, 0x8b, 0xd2, 0x5e, 0x97, 0x74, 0x50, 0xc8, 0xa1, 0xed,
	0xff, 0x8e, 0x54, 0xfb, 0x85, 0xd0, 0xdd, 0x3f, 0xd2, 0x3e, 0xa7, 0x9b, 0x0a, 0x1e, 0xa6, 0x20,
	0x43, 0x70, 0x48, 0x8b, 0x74, 0x6a, 0xfd, 0xc3, 0xf9, 0xb2, 0x69, 0xbd, 0x2d, 0x9b, 0x7b, 0xe6,
	0xe5, 0x2a, 0x1a, 0x33, 0x81, 0x3c, 0xf1, 0xf5, 0x88, 0x5d, 0x49, 0x3d, 0xfc, 0xe2, 0xf6, 0x01,
	0xad, 0xa5, 0x10, 0x8a, 0x89, 0x00, 0xa9, 0x9d, 0x8d, 0xe2, 0xec, 0xf0, 0x3b, 0xb0, 0xcf, 0x68,
	0xd5, 0x4f, 0x70, 0x2a, 0xb5, 0x53, 0xf9, 0x4f, 0xdb, 0x4f, 0xdc, 0xbf, 0x9c, 0xaf, 0x5c, 0xb2,
	0x58, 0xb9, 0xe4, 0x7d, 0xe5, 0x92, 0xe7, 0xdc, 0xb5, 0x16, 0xb9, 0x6b, 0xbd, 0xe6, 0xae, 0x75,
	0xd7, 0x8d, 0x85, 0x1e, 0x4d, 0x03, 0x16, 0x62, 0xc2, 0x8b, 0x01, 0x74, 0x31, 0x8d, 0xcb, 0x22,
	0xe2, 0xb3, 0xf5, 0x17, 0xf9, 0x41, 0x28, 0xb8, 0x7e, 0x9c, 0x80, 0x0a, 0xaa, 0xe5, 0x6c, 0x4f,
	0x3f, 0x02, 0x00, 0x00, 0xff, 0xff, 0x2f, 0x4a, 0x81, 0x5f, 0xc3, 0x01, 0x00, 0x00,
}

func (m *VoteExtension) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *VoteExtension) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *VoteExtension) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.AssetsLockedEvents) > 0 {
		for iNdEx := len(m.AssetsLockedEvents) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.AssetsLockedEvents[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintVoteExtension(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *AssetsLockedEvent) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *AssetsLockedEvent) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *AssetsLockedEvent) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.Amount.Size()
		i -= size
		if _, err := m.Amount.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintVoteExtension(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.Recipient) > 0 {
		i -= len(m.Recipient)
		copy(dAtA[i:], m.Recipient)
		i = encodeVarintVoteExtension(dAtA, i, uint64(len(m.Recipient)))
		i--
		dAtA[i] = 0x12
	}
	{
		size := m.Sequence.Size()
		i -= size
		if _, err := m.Sequence.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintVoteExtension(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintVoteExtension(dAtA []byte, offset int, v uint64) int {
	offset -= sovVoteExtension(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *VoteExtension) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.AssetsLockedEvents) > 0 {
		for _, e := range m.AssetsLockedEvents {
			l = e.Size()
			n += 1 + l + sovVoteExtension(uint64(l))
		}
	}
	return n
}

func (m *AssetsLockedEvent) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Sequence.Size()
	n += 1 + l + sovVoteExtension(uint64(l))
	l = len(m.Recipient)
	if l > 0 {
		n += 1 + l + sovVoteExtension(uint64(l))
	}
	l = m.Amount.Size()
	n += 1 + l + sovVoteExtension(uint64(l))
	return n
}

func sovVoteExtension(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozVoteExtension(x uint64) (n int) {
	return sovVoteExtension(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *VoteExtension) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowVoteExtension
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: VoteExtension: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: VoteExtension: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AssetsLockedEvents", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVoteExtension
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthVoteExtension
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthVoteExtension
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AssetsLockedEvents = append(m.AssetsLockedEvents, &AssetsLockedEvent{})
			if err := m.AssetsLockedEvents[len(m.AssetsLockedEvents)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipVoteExtension(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthVoteExtension
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *AssetsLockedEvent) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowVoteExtension
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: AssetsLockedEvent: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: AssetsLockedEvent: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sequence", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVoteExtension
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthVoteExtension
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthVoteExtension
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Sequence.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Recipient", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVoteExtension
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthVoteExtension
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthVoteExtension
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Recipient = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowVoteExtension
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthVoteExtension
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthVoteExtension
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Amount.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipVoteExtension(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthVoteExtension
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipVoteExtension(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowVoteExtension
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowVoteExtension
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowVoteExtension
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthVoteExtension
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupVoteExtension
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthVoteExtension
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthVoteExtension        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowVoteExtension          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupVoteExtension = fmt.Errorf("proto: unexpected end of group")
)
