// Copyright 2019 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package client

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"

	"github.com/evmos/ethermint/rpc/types"
)

type ServiceRegistry struct {
	Mu       sync.Mutex
	Services map[string]service
}

// service represents a registered object.
type service struct {
	name          string                     // name for service
	callbacks     map[string]*types.Callback // registered handlers
	subscriptions map[string]*types.Callback // available subscriptions/notifications
}

func (r *ServiceRegistry) RegisterName(name string, rcvr interface{}) error {
	rcvrVal := reflect.ValueOf(rcvr)
	if name == "" {
		return fmt.Errorf("no service name for type %s", rcvrVal.Type().String())
	}
	callbacks := suitableCallbacks(rcvrVal)
	if len(callbacks) == 0 {
		return fmt.Errorf("service %T doesn't have any suitable methods/subscriptions to expose", rcvr)
	}

	r.Mu.Lock()
	defer r.Mu.Unlock()
	if r.Services == nil {
		r.Services = make(map[string]service)
	}
	svc, ok := r.Services[name]
	if !ok {
		svc = service{
			name:          name,
			callbacks:     make(map[string]*types.Callback),
			subscriptions: make(map[string]*types.Callback),
		}
		r.Services[name] = svc
	}
	for name := range callbacks {
		cb := callbacks[name]
		if cb.IsSubscribe {
			svc.subscriptions[name] = cb
		} else {
			svc.callbacks[name] = cb
		}
	}
	return nil
}

// callback returns the callback corresponding to the given RPC method name.
func (r *ServiceRegistry) callback(method string) *types.Callback {
	elem := strings.SplitN(method, types.ServiceMethodSeparator, 2)
	if len(elem) != 2 {
		return nil
	}
	r.Mu.Lock()
	defer r.Mu.Unlock()
	return r.Services[elem[0]].callbacks[elem[1]]
}

// subscription returns a subscription callback in the given service.
func (r *ServiceRegistry) subscription(service, name string) *types.Callback {
	r.Mu.Lock()
	defer r.Mu.Unlock()
	return r.Services[service].subscriptions[name]
}

// suitableCallbacks iterates over the methods of the given type. It determines if a method
// satisfies the criteria for a RPC callback or a subscription callback and adds it to the
// collection of callbacks. See server documentation for a summary of these criteria.
func suitableCallbacks(receiver reflect.Value) map[string]*types.Callback {
	typ := receiver.Type()
	callbacks := make(map[string]*types.Callback)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		if method.PkgPath != "" {
			continue // method not exported
		}
		cb := types.NewCallback(receiver, method.Func)
		if cb == nil {
			continue // function invalid
		}
		name := formatName(method.Name)
		callbacks[name] = cb
	}
	return callbacks
}

// formatName converts to first character of name to lowercase.
func formatName(name string) string {
	ret := []rune(name)
	if len(ret) > 0 {
		ret[0] = unicode.ToLower(ret[0])
	}
	return string(ret)
}
