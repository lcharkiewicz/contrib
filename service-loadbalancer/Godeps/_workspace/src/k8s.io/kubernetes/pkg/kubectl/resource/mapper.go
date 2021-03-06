/*
Copyright 2014 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resource

import (
	"fmt"
	"reflect"

	"k8s.io/kubernetes/pkg/api/meta"
	"k8s.io/kubernetes/pkg/runtime"
)

// DisabledClientForMapping allows callers to avoid allowing remote calls when handling
// resources.
type DisabledClientForMapping struct {
	ClientMapper
}

func (f DisabledClientForMapping) ClientForMapping(mapping *meta.RESTMapping) (RESTClient, error) {
	return nil, nil
}

// Mapper is a convenience struct for holding references to the three interfaces
// needed to create Info for arbitrary objects.
type Mapper struct {
	runtime.ObjectTyper
	meta.RESTMapper
	ClientMapper
	runtime.Decoder
}

// InfoForData creates an Info object for the given data. An error is returned
// if any of the decoding or client lookup steps fail. Name and namespace will be
// set into Info if the mapping's MetadataAccessor can retrieve them.
func (m *Mapper) InfoForData(data []byte, source string) (*Info, error) {
	versions := &runtime.VersionedObjects{}
	_, gvk, err := m.Decode(data, nil, versions)
	if err != nil {
		return nil, fmt.Errorf("unable to decode %q: %v", source, err)
	}
	mapping, err := m.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("unable to recognize %q: %v", source, err)
	}

	client, err := m.ClientForMapping(mapping)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to a server to handle %q: %v", mapping.Resource, err)
	}

	// TODO: decoding the version object is convenient, but questionable. This is used by apply
	// and rolling-update today, but both of those cases should probably be requesting the raw
	// object and performing their own decoding.
	obj, versioned := versions.Last(), versions.First()
	name, _ := mapping.MetadataAccessor.Name(obj)
	namespace, _ := mapping.MetadataAccessor.Namespace(obj)
	resourceVersion, _ := mapping.MetadataAccessor.ResourceVersion(obj)

	return &Info{
		Mapping:         mapping,
		Client:          client,
		Namespace:       namespace,
		Name:            name,
		Source:          source,
		VersionedObject: versioned,
		Object:          obj,
		ResourceVersion: resourceVersion,
	}, nil
}

// InfoForObject creates an Info object for the given Object. An error is returned
// if the object cannot be introspected. Name and namespace will be set into Info
// if the mapping's MetadataAccessor can retrieve them.
func (m *Mapper) InfoForObject(obj runtime.Object) (*Info, error) {
	groupVersionKind, err := m.ObjectKind(obj)
	if err != nil {
		return nil, fmt.Errorf("unable to get type info from the object %q: %v", reflect.TypeOf(obj), err)
	}
	mapping, err := m.RESTMapping(groupVersionKind.GroupKind(), groupVersionKind.Version)
	if err != nil {
		return nil, fmt.Errorf("unable to recognize %v: %v", groupVersionKind, err)
	}
	client, err := m.ClientForMapping(mapping)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to a server to handle %q: %v", mapping.Resource, err)
	}
	name, _ := mapping.MetadataAccessor.Name(obj)
	namespace, _ := mapping.MetadataAccessor.Namespace(obj)
	resourceVersion, _ := mapping.MetadataAccessor.ResourceVersion(obj)
	return &Info{
		Mapping:   mapping,
		Client:    client,
		Namespace: namespace,
		Name:      name,

		Object:          obj,
		ResourceVersion: resourceVersion,
	}, nil
}
