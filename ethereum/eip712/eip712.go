// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE
package eip712

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// Go representation of a JSON object
type goJSON map[string]interface{}

// WrapTxToTypedData is an ultimate method that wraps Amino-encoded Cosmos Tx JSON data
// into an EIP712-compatible TypedData request.
func WrapTxToTypedData(
	cdc codectypes.AnyUnpacker,
	chainID uint64,
	msg sdk.Msg,
	data []byte,
	feeDelegation *FeeDelegationOptions,
) (apitypes.TypedData, error) {
	txData := make(goJSON)

	if err := json.Unmarshal(data, &txData); err != nil {
		return apitypes.TypedData{}, errorsmod.Wrap(errortypes.ErrJSONUnmarshal, "failed to JSON unmarshal data")
	}

	numMessages, err := flattenPayloadMessages(txData)
	if err != nil {
		return apitypes.TypedData{}, fmt.Errorf("failed to flatten payload JSON messages: %w", err)
	}

	domain := apitypes.TypedDataDomain{
		Name:              "Cosmos Web3",
		Version:           "1.0.0",
		ChainId:           math.NewHexOrDecimal256(int64(chainID)),
		VerifyingContract: "cosmos",
		Salt:              "0",
	}

	payloadTypes, err := extractPayloadTypes(cdc, txData, numMessages)
	if err != nil {
		return apitypes.TypedData{}, err
	}

	if feeDelegation != nil {
		feeInfo, ok := txData["fee"].(map[string]interface{})
		if !ok {
			return apitypes.TypedData{}, errorsmod.Wrap(errortypes.ErrInvalidType, "cannot parse fee from tx data")
		}

		feeInfo["feePayer"] = feeDelegation.FeePayer.String()

		// also patching payloadTypes to include feePayer
		payloadTypes["Fee"] = []apitypes.Type{
			{Name: "feePayer", Type: "string"},
			{Name: "amount", Type: "Coin[]"},
			{Name: "gas", Type: "string"},
		}
	}

	typedData := apitypes.TypedData{
		Types:       payloadTypes,
		PrimaryType: "Tx",
		Domain:      domain,
		Message:     txData,
	}

	return typedData, nil
}

type FeeDelegationOptions struct {
	FeePayer sdk.AccAddress
}

func payloadMsgField(i int) string {
	return fmt.Sprintf("msg%d", i)
}

// flattenPayloadMessages flattens the input payload's messages in-place, representing
// them as key-value pairs of "Message{i}": {Msg}, rather than an array of Msgs.
// We do this to support messages with different schemas, which would be invalid syntax in an
// EIP-712 array.
func flattenPayloadMessages(payload goJSON) (int, error) {
	interfaceMsgs, ok := payload["msgs"]
	if !ok {
		return 0, errors.New("no messages found in payload, unable to parse")
	}

	// Cast from interface{} to []interface{}
	messages, ok := interfaceMsgs.([]interface{})
	if !ok {
		return 0, errors.New("expected type array of messages, cannot parse")
	}

	for i, interfaceMsg := range messages {
		msg, ok := interfaceMsg.(map[string]interface{})
		if !ok {
			return 0, fmt.Errorf("msg at index %d is not valid JSON: %v", i, msg)
		}

		field := payloadMsgField(i)

		if _, hasField := payload[field]; hasField {
			return 0, fmt.Errorf("malformed payload received, did not expect to find key with field %v", field)
		}

		payload[field] = msg
	}

	delete(payload, "msgs")

	return len(messages), nil
}

func extractPayloadTypes(cdc codectypes.AnyUnpacker, payload goJSON, numMessages int) (apitypes.Types, error) {
	rootTypes := apitypes.Types{
		"EIP712Domain": {
			{
				Name: "name",
				Type: "string",
			},
			{
				Name: "version",
				Type: "string",
			},
			{
				Name: "chainId",
				Type: "uint256",
			},
			{
				Name: "verifyingContract",
				Type: "string",
			},
			{
				Name: "salt",
				Type: "string",
			},
		},
		"Tx": {
			{Name: "account_number", Type: "string"},
			{Name: "chain_id", Type: "string"},
			{Name: "fee", Type: "Fee"},
			{Name: "memo", Type: "string"},
			{Name: "sequence", Type: "string"},
			// Note timeout_height was removed because it was not getting filled with the legacyTx
			// {Name: "timeout_height", Type: "string"},
		},
		"Fee": {
			{Name: "amount", Type: "Coin[]"},
			{Name: "gas", Type: "string"},
		},
		"Coin": {
			{Name: "denom", Type: "string"},
			{Name: "amount", Type: "string"},
		},
	}

	for i := 0; i < numMessages; i++ {
		msg, ok := payload[payloadMsgField(i)]

		if !ok {
			return nil, fmt.Errorf("ran out of messages at index (%d), expected total of (%d)", i, numMessages)
		}

		msgTypedef, err := walkMsgTypes(cdc, rootTypes, msg)

		if err != nil {
			return nil, err
		}

		rootTypes["Tx"] = append(rootTypes["Tx"], apitypes.Type{
			Name: payloadMsgField(i),
			Type: msgTypedef,
		})
	}

	return rootTypes, nil
}

const typeDefPrefix = "_"

// addTypesToRoot attempts to add the types to the root at key typeDef and returns the key at which the types are
// present, or an error if they cannot be added. If the typeDef key is a duplicate, we return the key corresponding
// to an identical copy if present (without modifying the structure), otherwise we insert the types at the next
// available typeDef-{n} field. We do this to support identically named payloads with different schemas.
func addTypesToRoot(rootTypes apitypes.Types, typeDef string, types []apitypes.Type) (string, error) {
	var typeDefKey string

	duplicateIndex := 0

	for {
		typeDefKey = fmt.Sprintf("%v%d", typeDef, duplicateIndex)
		duplicateTypes, ok := rootTypes[typeDefKey]

		// Found identical duplicate
		if ok && typesAreEqual(types, duplicateTypes) {
			return typeDefKey, nil
		}

		// Found no element
		if !ok {
			break
		}

		duplicateIndex++

		if duplicateIndex == 1000 {
			return "", errors.New("exceeded maximum number of duplicates for a single type definition")
		}
	}

	// Add new type to root at current duplicate index
	rootTypes[typeDefKey] = types
	return typeDefKey, nil
}

func typesAreEqual(types1 []apitypes.Type, types2 []apitypes.Type) bool {
	if len(types1) != len(types2) {
		return false
	}

	n := len(types1)

	for i := 0; i < n; i++ {
		if types1[i].Name != types2[i].Name || types1[i].Type != types2[i].Type {
			return false
		}
	}

	return true
}

func walkMsgTypes(cdc codectypes.AnyUnpacker, typeMap apitypes.Types, in interface{}) (msgField string, err error) {
	defer doRecover(&err)

	t := reflect.TypeOf(in)
	v := reflect.ValueOf(in)

	for {
		if t.Kind() == reflect.Ptr ||
			t.Kind() == reflect.Interface {
			t = t.Elem()
			v = v.Elem()

			continue
		}

		break
	}

	if t.Kind() != reflect.Map {
		return "", errors.New("expected message format as map, could not parse message")
	}

	rootType := v.MapIndex(reflect.ValueOf("type")).Interface().(string)

	// Reformat root type name
	tokens := strings.Split(rootType, "/")
	if len(tokens) == 1 {
		rootType = fmt.Sprintf("Type%v", rootType)
	} else {
		rootType = fmt.Sprintf("Type%v", tokens[len(tokens)-1])
	}

	return traverseFields(cdc, typeMap, rootType, typeDefPrefix, t, v)
}

type cosmosAnyWrapper struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// traverseFields walks all types in the given map, recursively adding sub-maps as new types when necessary, and adds the map's type definition
// to typeMap. It returns the key to the type definition, and an error if it failed.
func traverseFields(
	cdc codectypes.AnyUnpacker,
	typeMap apitypes.Types,
	rootType string,
	prefix string,
	t reflect.Type,
	v reflect.Value,
) (string, error) {

	mapKeys := v.MapKeys()
	sort.Slice(mapKeys, func(i, j int) bool {
		return strings.Compare(mapKeys[i].String(), mapKeys[j].String()) > 0
	})

	newTypes := []apitypes.Type{}

	for _, key := range mapKeys {
		field := v.MapIndex(key)
		fieldType := field.Type()
		fieldName := key.String()

		fieldType, field = unwrapToElem(fieldType, field)

		var isCollection bool
		if fieldType.Kind() == reflect.Array || fieldType.Kind() == reflect.Slice {
			if field.Len() == 0 {
				// skip empty collections from type mapping
				continue
			}

			fieldType, field = unwrapToElem(fieldType.Elem(), field.Index(0))
			isCollection = true
		}

		fieldPrefix := fmt.Sprintf("%s.%s", prefix, fieldName)

		ethType := typeToEth(fieldType)
		if ethType != "" {
			// Support array types
			if isCollection && fieldType.Kind() != reflect.Slice && fieldType.Kind() != reflect.Array {
				ethType += "[]"
			}

			newTypes = append(newTypes, apitypes.Type{
				Name: fieldName,
				Type: ethType,
			})

			continue
		}

		if fieldType.Kind() == reflect.Map {
			fieldTypedef, err := traverseFields(cdc, typeMap, rootType, fieldPrefix, fieldType, field)

			if err != nil {
				return "", err
			}

			if isCollection {
				fieldTypedef = sanitizeTypedef(fieldTypedef) + "[]"
			} else {
				fieldTypedef = sanitizeTypedef(fieldTypedef)
			}

			newTypes = append(newTypes, apitypes.Type{
				Name: fieldName,
				Type: fieldTypedef,
			})

			continue
		}
	}

	var typeDef string
	if prefix == typeDefPrefix {
		typeDef = rootType
	} else {
		typeDef = sanitizeTypedef(prefix)
	}

	return addTypesToRoot(typeMap, typeDef, newTypes)
}

// unwrapToElem unwraps pointer or interface types to get their underlying values
func unwrapToElem(t reflect.Type, v reflect.Value) (reflect.Type, reflect.Value) {
	fieldType := t
	field := v

	for {
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()

			if field.IsValid() {
				field = field.Elem()
			}

			continue
		}

		if fieldType.Kind() == reflect.Interface {
			fieldType = reflect.TypeOf(field.Interface())

			if field.IsValid() {
				field = field.Elem()
			}

			continue
		}

		if field.Kind() == reflect.Ptr {
			field = field.Elem()
			continue
		}

		break
	}

	return fieldType, field
}

// _.foo_bar.baz -> TypeFooBarBaz
//
// this is needed for Geth's own signing code which doesn't
// tolerate complex type names
func sanitizeTypedef(str string) string {
	buf := new(bytes.Buffer)
	parts := strings.Split(str, ".")
	caser := cases.Title(language.English, cases.NoLower)

	for _, part := range parts {
		if part == "_" {
			buf.WriteString("Type")
			continue
		}

		subparts := strings.Split(part, "_")
		for _, subpart := range subparts {
			buf.WriteString(caser.String(subpart))
		}
	}

	return buf.String()
}

var (
	hashType      = reflect.TypeOf(common.Hash{})
	addressType   = reflect.TypeOf(common.Address{})
	bigIntType    = reflect.TypeOf(big.Int{})
	cosmIntType   = reflect.TypeOf(sdkmath.Int{})
	cosmDecType   = reflect.TypeOf(sdk.Dec{})
	cosmosAnyType = reflect.TypeOf(&codectypes.Any{})
	timeType      = reflect.TypeOf(time.Time{})

	edType = reflect.TypeOf(ed25519.PubKey{})
)

// typeToEth supports only basic types and arrays of basic types.
// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-712.md
func typeToEth(typ reflect.Type) string {
	const str = "string"

	switch typ.Kind() {
	case reflect.String:
		return str
	case reflect.Bool:
		return "bool"
	case reflect.Int:
		return "int64"
	case reflect.Int8:
		return "int8"
	case reflect.Int16:
		return "int16"
	case reflect.Int32:
		return "int32"
	case reflect.Int64:
		return "int64"
	case reflect.Uint:
		return "uint64"
	case reflect.Uint8:
		return "uint8"
	case reflect.Uint16:
		return "uint16"
	case reflect.Uint32:
		return "uint32"
	case reflect.Uint64:
		return "uint64"
	case reflect.Float64:
		// JSON numbers are represented as Float64 by default, see https://pkg.go.dev/encoding/json#Unmarshal
		// Since there is no fixed or floating point in Solidity, we use Int64 instead
		return "int64"
	case reflect.Slice:
		ethName := typeToEth(typ.Elem())
		if len(ethName) > 0 {
			return ethName + "[]"
		}
	case reflect.Array:
		ethName := typeToEth(typ.Elem())
		if len(ethName) > 0 {
			return ethName + "[]"
		}
	case reflect.Ptr:
		if typ.Elem().ConvertibleTo(bigIntType) ||
			typ.Elem().ConvertibleTo(timeType) ||
			typ.Elem().ConvertibleTo(edType) ||
			typ.Elem().ConvertibleTo(cosmDecType) ||
			typ.Elem().ConvertibleTo(cosmIntType) {
			return str
		}
	case reflect.Struct:
		if typ.ConvertibleTo(hashType) ||
			typ.ConvertibleTo(addressType) ||
			typ.ConvertibleTo(bigIntType) ||
			typ.ConvertibleTo(edType) ||
			typ.ConvertibleTo(timeType) ||
			typ.ConvertibleTo(cosmDecType) ||
			typ.ConvertibleTo(cosmIntType) {
			return str
		}
	}

	return ""
}

func doRecover(err *error) {
	if r := recover(); r != nil {
		if e, ok := r.(error); ok {
			e = errorsmod.Wrap(e, "panicked with error")
			*err = e
			return
		}

		*err = fmt.Errorf("%v", r)
	}
}
