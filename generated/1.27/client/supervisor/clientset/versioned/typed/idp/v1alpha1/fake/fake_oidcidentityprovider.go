// Copyright 2020-2023 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "go.pinniped.dev/generated/1.27/apis/supervisor/idp/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeOIDCIdentityProviders implements OIDCIdentityProviderInterface
type FakeOIDCIdentityProviders struct {
	Fake *FakeIDPV1alpha1
	ns   string
}

var oidcidentityprovidersResource = v1alpha1.SchemeGroupVersion.WithResource("oidcidentityproviders")

var oidcidentityprovidersKind = v1alpha1.SchemeGroupVersion.WithKind("OIDCIdentityProvider")

// Get takes name of the oIDCIdentityProvider, and returns the corresponding oIDCIdentityProvider object, and an error if there is any.
func (c *FakeOIDCIdentityProviders) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.OIDCIdentityProvider, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(oidcidentityprovidersResource, c.ns, name), &v1alpha1.OIDCIdentityProvider{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OIDCIdentityProvider), err
}

// List takes label and field selectors, and returns the list of OIDCIdentityProviders that match those selectors.
func (c *FakeOIDCIdentityProviders) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.OIDCIdentityProviderList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(oidcidentityprovidersResource, oidcidentityprovidersKind, c.ns, opts), &v1alpha1.OIDCIdentityProviderList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.OIDCIdentityProviderList{ListMeta: obj.(*v1alpha1.OIDCIdentityProviderList).ListMeta}
	for _, item := range obj.(*v1alpha1.OIDCIdentityProviderList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested oIDCIdentityProviders.
func (c *FakeOIDCIdentityProviders) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(oidcidentityprovidersResource, c.ns, opts))

}

// Create takes the representation of a oIDCIdentityProvider and creates it.  Returns the server's representation of the oIDCIdentityProvider, and an error, if there is any.
func (c *FakeOIDCIdentityProviders) Create(ctx context.Context, oIDCIdentityProvider *v1alpha1.OIDCIdentityProvider, opts v1.CreateOptions) (result *v1alpha1.OIDCIdentityProvider, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(oidcidentityprovidersResource, c.ns, oIDCIdentityProvider), &v1alpha1.OIDCIdentityProvider{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OIDCIdentityProvider), err
}

// Update takes the representation of a oIDCIdentityProvider and updates it. Returns the server's representation of the oIDCIdentityProvider, and an error, if there is any.
func (c *FakeOIDCIdentityProviders) Update(ctx context.Context, oIDCIdentityProvider *v1alpha1.OIDCIdentityProvider, opts v1.UpdateOptions) (result *v1alpha1.OIDCIdentityProvider, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(oidcidentityprovidersResource, c.ns, oIDCIdentityProvider), &v1alpha1.OIDCIdentityProvider{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OIDCIdentityProvider), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeOIDCIdentityProviders) UpdateStatus(ctx context.Context, oIDCIdentityProvider *v1alpha1.OIDCIdentityProvider, opts v1.UpdateOptions) (*v1alpha1.OIDCIdentityProvider, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(oidcidentityprovidersResource, "status", c.ns, oIDCIdentityProvider), &v1alpha1.OIDCIdentityProvider{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OIDCIdentityProvider), err
}

// Delete takes name of the oIDCIdentityProvider and deletes it. Returns an error if one occurs.
func (c *FakeOIDCIdentityProviders) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(oidcidentityprovidersResource, c.ns, name, opts), &v1alpha1.OIDCIdentityProvider{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeOIDCIdentityProviders) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(oidcidentityprovidersResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.OIDCIdentityProviderList{})
	return err
}

// Patch applies the patch and returns the patched oIDCIdentityProvider.
func (c *FakeOIDCIdentityProviders) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.OIDCIdentityProvider, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(oidcidentityprovidersResource, c.ns, name, pt, data, subresources...), &v1alpha1.OIDCIdentityProvider{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.OIDCIdentityProvider), err
}
