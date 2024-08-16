// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE
package eip712

import (
	"encoding/json"
	"errors"
	"fmt"

	msgv1 "cosmossdk.io/api/cosmos/msg/v1"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"

	sdk "github.com/cosmos/cosmos-sdk/types"
	txTypes "github.com/cosmos/cosmos-sdk/types/tx"

	apitypes "github.com/ethereum/go-ethereum/signer/core/apitypes"
	mezo "github.com/mezo-org/mezod/types"
)

type aminoMessage struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// LegacyGetEIP712BytesForMsg returns the EIP-712 object bytes for the given SignDoc bytes by decoding the bytes into
// an EIP-712 object, then converting via LegacyWrapTxToTypedData. See https://eips.ethereum.org/EIPS/eip-712 for more.
func LegacyGetEIP712BytesForMsg(signDocBytes []byte) ([]byte, error) {
	typedData, err := LegacyGetEIP712TypedDataForMsg(signDocBytes)
	if err != nil {
		return nil, err
	}

	_, rawData, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, fmt.Errorf("could not get EIP-712 object bytes: %w", err)
	}

	return []byte(rawData), nil
}

// LegacyGetEIP712TypedDataForMsg returns the EIP-712 TypedData representation for either
// Amino or Protobuf encoded signature doc bytes.
func LegacyGetEIP712TypedDataForMsg(signDocBytes []byte) (apitypes.TypedData, error) {
	// Attempt to decode as both Amino and Protobuf since the message format is unknown.
	// If either decode works, we can move forward with the corresponding typed data.
	typedDataAmino, errAmino := legacyDecodeAminoSignDoc(signDocBytes)
	if errAmino == nil && isValidEIP712Payload(typedDataAmino) {
		return typedDataAmino, nil
	}
	typedDataProtobuf, errProtobuf := legacyDecodeProtobufSignDoc(signDocBytes)
	if errProtobuf == nil && isValidEIP712Payload(typedDataProtobuf) {
		return typedDataProtobuf, nil
	}

	return apitypes.TypedData{}, fmt.Errorf("could not decode sign doc as either Amino or Protobuf.\n amino: %v\n protobuf: %v", errAmino, errProtobuf)
}

// legacyDecodeAminoSignDoc attempts to decode the provided sign doc (bytes) as an Amino payload
// and returns a signable EIP-712 TypedData object.
func legacyDecodeAminoSignDoc(signDocBytes []byte) (apitypes.TypedData, error) {
	// Ensure codecs have been initialized
	if err := validateCodecInit(); err != nil {
		return apitypes.TypedData{}, err
	}

	var aminoDoc legacytx.StdSignDoc
	if err := aminoCodec.UnmarshalJSON(signDocBytes, &aminoDoc); err != nil {
		return apitypes.TypedData{}, err
	}

	var fees legacytx.StdFee
	if err := aminoCodec.UnmarshalJSON(aminoDoc.Fee, &fees); err != nil {
		return apitypes.TypedData{}, err
	}

	// Validate payload messages
	msgs := make([]sdk.Msg, len(aminoDoc.Msgs))
	for i, jsonMsg := range aminoDoc.Msgs {
		var m sdk.Msg
		if err := aminoCodec.UnmarshalJSON(jsonMsg, &m); err != nil {
			return apitypes.TypedData{}, fmt.Errorf("failed to unmarshal sign doc message: %w", err)
		}
		msgs[i] = m
	}

	signer, err := legacyValidatePayloadMessages(msgs)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	if len(signer) == 0 {
		return apitypes.TypedData{}, errors.New("signer not found")
	}

	// Use first message for fee payer and type inference
	msg := msgs[0]

	// By convention, the fee payer is the first address in the list of signers.
	feePayer := sdk.AccAddress(signer)
	feeDelegation := &FeeDelegationOptions{
		FeePayer: feePayer,
	}

	chainID, err := mezo.ParseChainID(aminoDoc.ChainID)
	if err != nil {
		return apitypes.TypedData{}, errors.New("invalid chain ID passed as argument")
	}

	typedData, err := LegacyWrapTxToTypedData(
		protoCodec,
		chainID.Uint64(),
		msg,
		signDocBytes,
		feeDelegation,
	)
	if err != nil {
		return apitypes.TypedData{}, fmt.Errorf("could not convert to EIP712 representation: %w", err)
	}

	return typedData, nil
}

// legacyDecodeProtobufSignDoc attempts to decode the provided sign doc (bytes) as a Protobuf payload
// and returns a signable EIP-712 TypedData object.
func legacyDecodeProtobufSignDoc(signDocBytes []byte) (apitypes.TypedData, error) {
	// Ensure codecs have been initialized
	if err := validateCodecInit(); err != nil {
		return apitypes.TypedData{}, err
	}

	signDoc := &txTypes.SignDoc{}
	if err := signDoc.Unmarshal(signDocBytes); err != nil {
		return apitypes.TypedData{}, err
	}

	authInfo := &txTypes.AuthInfo{}
	if err := authInfo.Unmarshal(signDoc.AuthInfoBytes); err != nil {
		return apitypes.TypedData{}, err
	}

	body := &txTypes.TxBody{}
	if err := body.Unmarshal(signDoc.BodyBytes); err != nil {
		return apitypes.TypedData{}, err
	}

	// Until support for these fields is added, throw an error at their presence
	if body.TimeoutHeight != 0 || len(body.ExtensionOptions) != 0 || len(body.NonCriticalExtensionOptions) != 0 {
		return apitypes.TypedData{}, errors.New("body contains unsupported fields: TimeoutHeight, ExtensionOptions, or NonCriticalExtensionOptions")
	}

	if len(authInfo.SignerInfos) != 1 {
		return apitypes.TypedData{}, fmt.Errorf("invalid number of signer infos provided, expected 1 got %v", len(authInfo.SignerInfos))
	}

	// Validate payload messages
	msgs := make([]sdk.Msg, len(body.Messages))
	for i, protoMsg := range body.Messages {
		var m sdk.Msg
		if err := protoCodec.UnpackAny(protoMsg, &m); err != nil {
			return apitypes.TypedData{}, fmt.Errorf("could not unpack message object with error %w", err)
		}
		msgs[i] = m
	}

	signer, err := legacyValidatePayloadMessages(msgs)
	if err != nil {
		return apitypes.TypedData{}, err
	}
	if len(signer) == 0 {
		return apitypes.TypedData{}, errors.New("signer not found")
	}

	// Use first message for fee payer and type inference
	msg := msgs[0]

	signerInfo := authInfo.SignerInfos[0]

	chainID, err := mezo.ParseChainID(signDoc.ChainId)
	if err != nil {
		return apitypes.TypedData{}, fmt.Errorf("invalid chain ID passed as argument: %w", err)
	}

	stdFee := &legacytx.StdFee{
		Amount: authInfo.Fee.Amount,
		Gas:    authInfo.Fee.GasLimit,
	}

	feePayer := sdk.AccAddress(signer)
	feeDelegation := &FeeDelegationOptions{
		FeePayer: feePayer,
	}

	// WrapTxToTypedData expects the payload as an Amino Sign Doc
	legacytx.RegressionTestingAminoCodec = legacy.Cdc
	signBytes := legacytx.StdSignBytes(
		signDoc.ChainId,
		signDoc.AccountNumber,
		signerInfo.Sequence,
		body.TimeoutHeight,
		*stdFee,
		msgs,
		body.Memo,
	)

	typedData, err := LegacyWrapTxToTypedData(
		protoCodec,
		chainID.Uint64(),
		msg,
		signBytes,
		feeDelegation,
	)
	if err != nil {
		return apitypes.TypedData{}, err
	}

	return typedData, nil
}

// validatePayloadMessages ensures that the transaction messages can be represented in an EIP-712
// encoding by checking that messages exist, are of the same type, and share a single signer.
func legacyValidatePayloadMessages(msgs []sdk.Msg) (string, error) {
	if len(msgs) == 0 {
		return "", errors.New("unable to build EIP-712 payload: transaction does contain any messages")
	}

	var msgType string
	var txSigner string

	for i, msg := range msgs {
		t, err := getMsgType(msg)
		if err != nil {
			return "", err
		}

		msgV2 := protoadapt.MessageV2Of(msg).ProtoReflect()
		msgV2Desc := msgV2.Descriptor()

		signersFields := proto.GetExtension(msgV2Desc.Options(), msgv1.E_Signer).([]string)

		if len(signersFields) != 1 {
			return "", errors.New("unable to build EIP-712 payload: expect exactly 1 signer")
		}

		signerField := signersFields[0]
		signerFieldDesc := msgV2Desc.Fields().ByName(protoreflect.Name(signerField))
		signer := msgV2.Get(signerFieldDesc).String()

		if i == 0 {
			msgType = t
			txSigner = signer
			continue
		}

		if t != msgType {
			return "", errors.New("unable to build EIP-712 payload: different types of messages detected")
		}

		if txSigner != signer {
			return "", errors.New("unable to build EIP-712 payload: multiple signers detected")
		}
	}

	return txSigner, nil
}

// getMsgType returns the message type prefix for the given Cosmos SDK Msg
func getMsgType(msg sdk.Msg) (string, error) {
	jsonBytes, err := aminoCodec.MarshalJSON(msg)
	if err != nil {
		return "", err
	}

	var jsonMsg aminoMessage
	if err := json.Unmarshal(jsonBytes, &jsonMsg); err != nil {
		return "", err
	}

	// Verify Type was successfully filled in
	if jsonMsg.Type == "" {
		return "", errors.New("could not decode message: type is missing")
	}

	return jsonMsg.Type, nil
}
