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
	"errors"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type FeeDelegationOptions struct {
	FeePayer sdk.AccAddress
}

const typeDefPrefix = "_"

// WrapTxToTypedData is an ultimate method that wraps Amino-encoded Cosmos Tx JSON data
// into an EIP712-compatible TypedData request.
func WrapTxToTypedData(
	chainID uint64,
	data []byte,
	feeDelegation *FeeDelegationOptions,
) (apitypes.TypedData, error) {
	txData := gjson.ParseBytes(data)

	if !txData.IsObject() {
		return apitypes.TypedData{}, errorsmod.Wrap(errortypes.ErrJSONUnmarshal, "failed to JSON unmarshal data")
	}

	txData, numMessages, err := FlattenPayloadMessages(txData)
	if err != nil {
		return apitypes.TypedData{}, fmt.Errorf("failed to flatten payload JSON messages: %w", err)
	}

	if !txData.IsObject() {
		return apitypes.TypedData{}, errorsmod.Wrap(errortypes.ErrJSONUnmarshal, "failed to flatten JSON data")
	}

	domain := apitypes.TypedDataDomain{
		Name:              "Cosmos Web3",
		Version:           "1.0.0",
		ChainId:           math.NewHexOrDecimal256(int64(chainID)),
		VerifyingContract: "cosmos",
		Salt:              "0",
	}

	payloadTypes, err := extractPayloadTypes(txData, numMessages)
	if err != nil {
		return apitypes.TypedData{}, err
	}

	if feeDelegation != nil {
		// TODO: Consider removing feePayer field as it's not necessary for signature verification

		txWithFee, err := sjson.Set(txData.Raw, "fee.feePayer", feeDelegation.FeePayer.String())
		if err != nil {
			return apitypes.TypedData{}, errorsmod.Wrap(errortypes.ErrInvalidType, "cannot update feePayer from tx data")
		}

		txData = gjson.Parse(txWithFee)
		if !txData.IsObject() {
			return apitypes.TypedData{}, errorsmod.Wrap(errortypes.ErrInvalidType, "could not update feePayer from tx data")
		}

		// also patching payloadTypes to include feePayer
		payloadTypes["Fee"] = []apitypes.Type{
			{Name: "feePayer", Type: "string"},
			{Name: "amount", Type: "Coin[]"},
			{Name: "gas", Type: "string"},
		}
	}

	txDataMap, ok := txData.Value().(map[string]interface{})
	if !ok {
		return apitypes.TypedData{}, errors.New("failed to parse JSON as map")
	}

	typedData := apitypes.TypedData{
		Types:       payloadTypes,
		PrimaryType: "Tx",
		Domain:      domain,
		Message:     txDataMap,
	}

	return typedData, nil
}

func payloadMsgField(i int) string {
	return fmt.Sprintf("msg%d", i)
}

// FlattenPayloadMessages flattens the input payload's messages, representing
// them as key-value pairs of "Message{i}": {Msg}, rather than an array of Msgs.
// We do this to support messages with different schemas.
func FlattenPayloadMessages(payload gjson.Result) (gjson.Result, int, error) {
	var err error
	flattened := payload.Raw

	msgs := payload.Get("msgs")

	if !msgs.Exists() {
		return gjson.Result{}, 0, errors.New("no messages found in payload, unable to parse")
	}

	if !msgs.IsArray() {
		return gjson.Result{}, 0, errors.New("expected type array of messages, cannot parse")
	}

	for i, msg := range msgs.Array() {
		if !msg.IsObject() {
			return gjson.Result{}, 0, fmt.Errorf("msg at index %d is not valid JSON: %v", i, msg)
		}

		msgField := payloadMsgField(i)

		if gjson.Get(flattened, msgField).Exists() {
			return gjson.Result{}, 0, fmt.Errorf("malformed payload received, did not expect to find key with field %v", msgField)
		}

		flattened, err = sjson.SetRaw(flattened, msgField, msg.Raw)
		if err != nil {
			return gjson.Result{}, 0, err
		}
	}

	flattened, err = sjson.Delete(flattened, "msgs")
	if err != nil {
		return gjson.Result{}, 0, err
	}

	return gjson.Parse(flattened), len(msgs.Array()), nil
}

func extractPayloadTypes(payload gjson.Result, numMessages int) (apitypes.Types, error) {
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
		msg := payload.Get(payloadMsgField(i))

		if !msg.IsObject() {
			return nil, fmt.Errorf("ran out of messages at index (%d), expected total of (%d)", i, numMessages)
		}

		msgTypedef, err := walkMsgTypes(rootTypes, msg)

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

// typesAreEqual compares two apitypes.Type arrays and returns a boolean indicating whether they have
// the same naming and type definitions. Assumes both arrays are in the same order.
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

func walkMsgTypes(typeMap apitypes.Types, json gjson.Result) (msgField string, err error) {
	defer doRecover(&err)

	if !json.IsObject() {
		return "", errors.New("expected json object, could not parse")
	}

	rootType := json.Get("type").Str

	if rootType == "" {
		// .Str is empty for arrays and objects
		return "", errors.New("malformed type value, expected type string")
	}

	// Reformat root type name
	tokens := strings.Split(rootType, "/")
	if len(tokens) == 1 {
		rootType = fmt.Sprintf("Type%v", rootType)
	} else {
		rootType = fmt.Sprintf("Type%v", tokens[len(tokens)-1])
	}

	return traverseFields(typeMap, rootType, typeDefPrefix, json)
}

// traverseFields walks all types in the given map, recursively adding sub-maps as new types when necessary, and adds the map's type definition
// to typeMap. It returns the key to the type definition, and an error if it failed.
func traverseFields(
	typeMap apitypes.Types,
	rootType string,
	prefix string,
	json gjson.Result,
) (string, error) {
	mapKeys, err := sortedJSONKeys(json)
	if err != nil {
		return "", fmt.Errorf("unable to parse JSON keys:%w", err)
	}

	newTypes := []apitypes.Type{}

	for _, fieldName := range mapKeys {
		field := json.Get(fieldName)
		if !field.Exists() {
			continue
		}

		var isCollection bool
		if field.IsArray() {
			if len(field.Array()) == 0 {
				// skip empty collections from type mapping
				continue
			}

			field = field.Array()[0]
			isCollection = true
		}

		fieldPrefix := fmt.Sprintf("%s.%s", prefix, fieldName)

		ethType := jsonToEth(field)
		if ethType != "" {
			// Support array types
			if isCollection && !field.IsArray() {
				ethType += "[]"
			}

			newTypes = append(newTypes, apitypes.Type{
				Name: fieldName,
				Type: ethType,
			})

			continue
		}

		if field.IsObject() {
			fieldTypedef, err := traverseFields(typeMap, rootType, fieldPrefix, field)

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

// sortedJSONKeys returns the sorted JSON keys for the input object.
func sortedJSONKeys(json gjson.Result) ([]string, error) {
	if !json.IsObject() {
		return nil, errors.New("expected JSON map to parse")
	}

	jsonMap := json.Map()

	keys := make([]string, len(jsonMap))

	i := 0
	for k := range jsonMap {
		keys[i] = k
		i++
	}

	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(keys[i], keys[j]) > 0
	})

	return keys, nil
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

// jsonToEth supports only basic types and arrays of basic types. Since this converts from a JSON object,
// it only needs to consider types supported by JSON. Returns an empty string for Objects.
// https://github.com/ethereum/EIPs/blob/master/EIPS/eip-712.md
func jsonToEth(json gjson.Result) string {
	switch json.Type {
	case gjson.True, gjson.False:
		return "bool"
	case gjson.Number:
		return "int64"
	case gjson.String:
		// Type gjson.String includes string literals, JSON arrays, and JSON objects
		if json.IsArray() {
			if len(json.Array()) == 0 {
				return ""
			}
			ethName := jsonToEth(json.Array()[0])
			if ethName != "" {
				return ethName + "[]"
			}
		}
		if json.IsObject() {
			return ""
		}
		return "string"
	default:
		return ""
	}
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
