package rpc

import (
	"go/ast"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"

	go_openrpc_reflect "github.com/etclabscore/go-openrpc-reflect"

	meta_schema "github.com/open-rpc/meta-schema"
)

// DiscoveryService defines a receiver type used for RPC discovery by reflection.
type DiscoveryService struct {
	doc *go_openrpc_reflect.Document
}

func NewDiscoveryService(doc *go_openrpc_reflect.Document) *DiscoveryService {
	return &DiscoveryService{
		doc,
	}
}

// Discover exposes a Discover method to the RPC receiver registration.
func (r *DiscoveryService) Discover() (*meta_schema.OpenrpcDocument, error) {
	return r.doc.Discover()
}

// sharedMetaRegisterer defines common metadata to all possible servers.
// These objects represent server-specific data that cannot be
// reflected.
var sharedMetaRegisterer = &go_openrpc_reflect.MetaT{
	GetInfoFn: func() (info *meta_schema.InfoObject) {
		info = &meta_schema.InfoObject{}
		title := "Ethermint JSON-RPC API"
		info.Title = (*meta_schema.InfoObjectProperties)(&title)

		version := params.VersionWithMeta + "/generated-at:" + time.Now().Format(time.RFC3339)
		info.Version = (*meta_schema.InfoObjectVersion)(&version)
		return info
	},

	GetExternalDocsFn: func() (exDocs *meta_schema.ExternalDocumentationObject) {
		exDocs = &meta_schema.ExternalDocumentationObject{}
		description := "Ethermint Documentation"
		exDocs.Description = (*meta_schema.ExternalDocumentationObjectDescription)(&description)
		url := "https://docs.ethermint.zone/basics/json_rpc.html"
		exDocs.Url = (*meta_schema.ExternalDocumentationObjectUrl)(&url)
		return exDocs
	},
}

// MetaRegistererForURL is a convenience function used to define the Server(s) information
// for a given listener, in this case organized by transport (ws, http, ipc).
// Since we can't get the protocol scheme from the net.Listener itself, we have to define this for each
// transport-specific document.
func MetaRegistererForURL(scheme string) *go_openrpc_reflect.MetaT {
	metaRegisterer := *sharedMetaRegisterer
	metaRegisterer.GetServersFn = func() func(listeners []net.Listener) (*meta_schema.Servers, error) {
		return func(listeners []net.Listener) (*meta_schema.Servers, error) {
			servers := []meta_schema.ServerObject{}

			for _, listener := range listeners {
				url := scheme + listener.Addr().String()
				network := listener.Addr().Network()

				server := meta_schema.ServerObject{
					Url:         (*meta_schema.ServerObjectUrl)(&url),
					Name:        (*meta_schema.ServerObjectName)(&network),
					Description: nil,
					Summary:     nil,
					Variables:   nil,
				}

				servers = append(servers, server)
			}

			return (*meta_schema.Servers)(&servers), nil
		}
	}
	return &metaRegisterer
}

// NewOpenRPCDocument returns a Document configured with application-specific logic.
func NewOpenRPCDocument() *go_openrpc_reflect.Document {
	doc := &go_openrpc_reflect.Document{}

	// Use a provided Ethereum default configuration as a base.
	appReflector := &go_openrpc_reflect.EthereumReflectorT{}

	appReflector.FnGetContentDescriptorRequired = func(r reflect.Value, m reflect.Method, field *ast.Field) (bool, error) {
		// Custom handling for eth_subscribe optional second parameter (depends on channel).
		if m.Name == "Subscribe" && len(field.Names) > 0 && field.Names[0].Name == "subscriptionOptions" {
			return false, nil
		}

		// Otherwise return the default.
		return go_openrpc_reflect.EthereumReflector.GetContentDescriptorRequired(r, m, field)
	}

	appReflector.FnGetMethodExternalDocs = func(r reflect.Value, m reflect.Method, funcDecl *ast.FuncDecl) (*meta_schema.ExternalDocumentationObject, error) {
		standard := go_openrpc_reflect.StandardReflector
		got, err := standard.GetMethodExternalDocs(r, m, funcDecl)
		if err != nil {
			return nil, err
		}

		if got.Url == nil {
			return got, nil
		}

		// Replace links to go-ethereum repo with current ethermint one
		newLink := meta_schema.ExternalDocumentationObjectUrl(strings.Replace(string(*got.Url), "github.com/ethereum/go-ethereum", "github.com/tharsis/ethermint", 1))
		got.Url = &newLink
		return got, nil
	}

	// Finally, register the configured reflector to the document.
	doc.WithReflector(appReflector)
	return doc
}

// RegisterOpenRPCAPIs registers apis to be describe on document
func RegisterOpenRPCAPIs(doc *go_openrpc_reflect.Document, apis []rpc.API) {
	for _, api := range apis {
		doc.RegisterReceiverName(api.Namespace, api.Service)
	}
}
