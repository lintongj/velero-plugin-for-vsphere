/*
Copyright the Velero contributors.

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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"time"

	v1 "github.com/vmware-tanzu/velero-plugin-for-vsphere/pkg/apis/backupdriver/v1"
	scheme "github.com/vmware-tanzu/velero-plugin-for-vsphere/pkg/generated/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SnapshotsGetter has a method to return a SnapshotInterface.
// A group's client should implement this interface.
type SnapshotsGetter interface {
	Snapshots(namespace string) SnapshotInterface
}

// SnapshotInterface has methods to work with Snapshot resources.
type SnapshotInterface interface {
	Create(*v1.Snapshot) (*v1.Snapshot, error)
	Update(*v1.Snapshot) (*v1.Snapshot, error)
	UpdateStatus(*v1.Snapshot) (*v1.Snapshot, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.Snapshot, error)
	List(opts metav1.ListOptions) (*v1.SnapshotList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Snapshot, err error)
	SnapshotExpansion
}

// snapshots implements SnapshotInterface
type snapshots struct {
	client rest.Interface
	ns     string
}

// newSnapshots returns a Snapshots
func newSnapshots(c *BackupdriverV1Client, namespace string) *snapshots {
	return &snapshots{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the snapshot, and returns the corresponding snapshot object, and an error if there is any.
func (c *snapshots) Get(name string, options metav1.GetOptions) (result *v1.Snapshot, err error) {
	result = &v1.Snapshot{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("snapshots").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Snapshots that match those selectors.
func (c *snapshots) List(opts metav1.ListOptions) (result *v1.SnapshotList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.SnapshotList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("snapshots").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested snapshots.
func (c *snapshots) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("snapshots").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a snapshot and creates it.  Returns the server's representation of the snapshot, and an error, if there is any.
func (c *snapshots) Create(snapshot *v1.Snapshot) (result *v1.Snapshot, err error) {
	result = &v1.Snapshot{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("snapshots").
		Body(snapshot).
		Do().
		Into(result)
	return
}

// Update takes the representation of a snapshot and updates it. Returns the server's representation of the snapshot, and an error, if there is any.
func (c *snapshots) Update(snapshot *v1.Snapshot) (result *v1.Snapshot, err error) {
	result = &v1.Snapshot{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("snapshots").
		Name(snapshot.Name).
		Body(snapshot).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *snapshots) UpdateStatus(snapshot *v1.Snapshot) (result *v1.Snapshot, err error) {
	result = &v1.Snapshot{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("snapshots").
		Name(snapshot.Name).
		SubResource("status").
		Body(snapshot).
		Do().
		Into(result)
	return
}

// Delete takes name of the snapshot and deletes it. Returns an error if one occurs.
func (c *snapshots) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("snapshots").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *snapshots) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("snapshots").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched snapshot.
func (c *snapshots) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Snapshot, err error) {
	result = &v1.Snapshot{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("snapshots").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
