// Copyright 2020-2023 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "go.pinniped.dev/generated/1.27/apis/supervisor/idp/v1alpha1"
	scheme "go.pinniped.dev/generated/1.27/client/supervisor/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// LDAPIdentityProvidersGetter has a method to return a LDAPIdentityProviderInterface.
// A group's client should implement this interface.
type LDAPIdentityProvidersGetter interface {
	LDAPIdentityProviders(namespace string) LDAPIdentityProviderInterface
}

// LDAPIdentityProviderInterface has methods to work with LDAPIdentityProvider resources.
type LDAPIdentityProviderInterface interface {
	Create(ctx context.Context, lDAPIdentityProvider *v1alpha1.LDAPIdentityProvider, opts v1.CreateOptions) (*v1alpha1.LDAPIdentityProvider, error)
	Update(ctx context.Context, lDAPIdentityProvider *v1alpha1.LDAPIdentityProvider, opts v1.UpdateOptions) (*v1alpha1.LDAPIdentityProvider, error)
	UpdateStatus(ctx context.Context, lDAPIdentityProvider *v1alpha1.LDAPIdentityProvider, opts v1.UpdateOptions) (*v1alpha1.LDAPIdentityProvider, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.LDAPIdentityProvider, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.LDAPIdentityProviderList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.LDAPIdentityProvider, err error)
	LDAPIdentityProviderExpansion
}

// lDAPIdentityProviders implements LDAPIdentityProviderInterface
type lDAPIdentityProviders struct {
	client rest.Interface
	ns     string
}

// newLDAPIdentityProviders returns a LDAPIdentityProviders
func newLDAPIdentityProviders(c *IDPV1alpha1Client, namespace string) *lDAPIdentityProviders {
	return &lDAPIdentityProviders{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the lDAPIdentityProvider, and returns the corresponding lDAPIdentityProvider object, and an error if there is any.
func (c *lDAPIdentityProviders) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.LDAPIdentityProvider, err error) {
	result = &v1alpha1.LDAPIdentityProvider{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("ldapidentityproviders").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of LDAPIdentityProviders that match those selectors.
func (c *lDAPIdentityProviders) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.LDAPIdentityProviderList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.LDAPIdentityProviderList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("ldapidentityproviders").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested lDAPIdentityProviders.
func (c *lDAPIdentityProviders) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("ldapidentityproviders").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a lDAPIdentityProvider and creates it.  Returns the server's representation of the lDAPIdentityProvider, and an error, if there is any.
func (c *lDAPIdentityProviders) Create(ctx context.Context, lDAPIdentityProvider *v1alpha1.LDAPIdentityProvider, opts v1.CreateOptions) (result *v1alpha1.LDAPIdentityProvider, err error) {
	result = &v1alpha1.LDAPIdentityProvider{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("ldapidentityproviders").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(lDAPIdentityProvider).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a lDAPIdentityProvider and updates it. Returns the server's representation of the lDAPIdentityProvider, and an error, if there is any.
func (c *lDAPIdentityProviders) Update(ctx context.Context, lDAPIdentityProvider *v1alpha1.LDAPIdentityProvider, opts v1.UpdateOptions) (result *v1alpha1.LDAPIdentityProvider, err error) {
	result = &v1alpha1.LDAPIdentityProvider{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("ldapidentityproviders").
		Name(lDAPIdentityProvider.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(lDAPIdentityProvider).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *lDAPIdentityProviders) UpdateStatus(ctx context.Context, lDAPIdentityProvider *v1alpha1.LDAPIdentityProvider, opts v1.UpdateOptions) (result *v1alpha1.LDAPIdentityProvider, err error) {
	result = &v1alpha1.LDAPIdentityProvider{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("ldapidentityproviders").
		Name(lDAPIdentityProvider.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(lDAPIdentityProvider).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the lDAPIdentityProvider and deletes it. Returns an error if one occurs.
func (c *lDAPIdentityProviders) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("ldapidentityproviders").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *lDAPIdentityProviders) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("ldapidentityproviders").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched lDAPIdentityProvider.
func (c *lDAPIdentityProviders) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.LDAPIdentityProvider, err error) {
	result = &v1alpha1.LDAPIdentityProvider{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("ldapidentityproviders").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
