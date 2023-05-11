//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright 2020-2023 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	unsafe "unsafe"

	login "go.pinniped.dev/generated/1.27/apis/concierge/login"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*ClusterCredential)(nil), (*login.ClusterCredential)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ClusterCredential_To_login_ClusterCredential(a.(*ClusterCredential), b.(*login.ClusterCredential), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*login.ClusterCredential)(nil), (*ClusterCredential)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_login_ClusterCredential_To_v1alpha1_ClusterCredential(a.(*login.ClusterCredential), b.(*ClusterCredential), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*TokenCredentialRequest)(nil), (*login.TokenCredentialRequest)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_TokenCredentialRequest_To_login_TokenCredentialRequest(a.(*TokenCredentialRequest), b.(*login.TokenCredentialRequest), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*login.TokenCredentialRequest)(nil), (*TokenCredentialRequest)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_login_TokenCredentialRequest_To_v1alpha1_TokenCredentialRequest(a.(*login.TokenCredentialRequest), b.(*TokenCredentialRequest), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*TokenCredentialRequestList)(nil), (*login.TokenCredentialRequestList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_TokenCredentialRequestList_To_login_TokenCredentialRequestList(a.(*TokenCredentialRequestList), b.(*login.TokenCredentialRequestList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*login.TokenCredentialRequestList)(nil), (*TokenCredentialRequestList)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_login_TokenCredentialRequestList_To_v1alpha1_TokenCredentialRequestList(a.(*login.TokenCredentialRequestList), b.(*TokenCredentialRequestList), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*TokenCredentialRequestSpec)(nil), (*login.TokenCredentialRequestSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_TokenCredentialRequestSpec_To_login_TokenCredentialRequestSpec(a.(*TokenCredentialRequestSpec), b.(*login.TokenCredentialRequestSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*login.TokenCredentialRequestSpec)(nil), (*TokenCredentialRequestSpec)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_login_TokenCredentialRequestSpec_To_v1alpha1_TokenCredentialRequestSpec(a.(*login.TokenCredentialRequestSpec), b.(*TokenCredentialRequestSpec), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*TokenCredentialRequestStatus)(nil), (*login.TokenCredentialRequestStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_TokenCredentialRequestStatus_To_login_TokenCredentialRequestStatus(a.(*TokenCredentialRequestStatus), b.(*login.TokenCredentialRequestStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*login.TokenCredentialRequestStatus)(nil), (*TokenCredentialRequestStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_login_TokenCredentialRequestStatus_To_v1alpha1_TokenCredentialRequestStatus(a.(*login.TokenCredentialRequestStatus), b.(*TokenCredentialRequestStatus), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_ClusterCredential_To_login_ClusterCredential(in *ClusterCredential, out *login.ClusterCredential, s conversion.Scope) error {
	out.ExpirationTimestamp = in.ExpirationTimestamp
	out.Token = in.Token
	out.ClientCertificateData = in.ClientCertificateData
	out.ClientKeyData = in.ClientKeyData
	return nil
}

// Convert_v1alpha1_ClusterCredential_To_login_ClusterCredential is an autogenerated conversion function.
func Convert_v1alpha1_ClusterCredential_To_login_ClusterCredential(in *ClusterCredential, out *login.ClusterCredential, s conversion.Scope) error {
	return autoConvert_v1alpha1_ClusterCredential_To_login_ClusterCredential(in, out, s)
}

func autoConvert_login_ClusterCredential_To_v1alpha1_ClusterCredential(in *login.ClusterCredential, out *ClusterCredential, s conversion.Scope) error {
	out.ExpirationTimestamp = in.ExpirationTimestamp
	out.Token = in.Token
	out.ClientCertificateData = in.ClientCertificateData
	out.ClientKeyData = in.ClientKeyData
	return nil
}

// Convert_login_ClusterCredential_To_v1alpha1_ClusterCredential is an autogenerated conversion function.
func Convert_login_ClusterCredential_To_v1alpha1_ClusterCredential(in *login.ClusterCredential, out *ClusterCredential, s conversion.Scope) error {
	return autoConvert_login_ClusterCredential_To_v1alpha1_ClusterCredential(in, out, s)
}

func autoConvert_v1alpha1_TokenCredentialRequest_To_login_TokenCredentialRequest(in *TokenCredentialRequest, out *login.TokenCredentialRequest, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_v1alpha1_TokenCredentialRequestSpec_To_login_TokenCredentialRequestSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_v1alpha1_TokenCredentialRequestStatus_To_login_TokenCredentialRequestStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1alpha1_TokenCredentialRequest_To_login_TokenCredentialRequest is an autogenerated conversion function.
func Convert_v1alpha1_TokenCredentialRequest_To_login_TokenCredentialRequest(in *TokenCredentialRequest, out *login.TokenCredentialRequest, s conversion.Scope) error {
	return autoConvert_v1alpha1_TokenCredentialRequest_To_login_TokenCredentialRequest(in, out, s)
}

func autoConvert_login_TokenCredentialRequest_To_v1alpha1_TokenCredentialRequest(in *login.TokenCredentialRequest, out *TokenCredentialRequest, s conversion.Scope) error {
	out.ObjectMeta = in.ObjectMeta
	if err := Convert_login_TokenCredentialRequestSpec_To_v1alpha1_TokenCredentialRequestSpec(&in.Spec, &out.Spec, s); err != nil {
		return err
	}
	if err := Convert_login_TokenCredentialRequestStatus_To_v1alpha1_TokenCredentialRequestStatus(&in.Status, &out.Status, s); err != nil {
		return err
	}
	return nil
}

// Convert_login_TokenCredentialRequest_To_v1alpha1_TokenCredentialRequest is an autogenerated conversion function.
func Convert_login_TokenCredentialRequest_To_v1alpha1_TokenCredentialRequest(in *login.TokenCredentialRequest, out *TokenCredentialRequest, s conversion.Scope) error {
	return autoConvert_login_TokenCredentialRequest_To_v1alpha1_TokenCredentialRequest(in, out, s)
}

func autoConvert_v1alpha1_TokenCredentialRequestList_To_login_TokenCredentialRequestList(in *TokenCredentialRequestList, out *login.TokenCredentialRequestList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]login.TokenCredentialRequest)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_v1alpha1_TokenCredentialRequestList_To_login_TokenCredentialRequestList is an autogenerated conversion function.
func Convert_v1alpha1_TokenCredentialRequestList_To_login_TokenCredentialRequestList(in *TokenCredentialRequestList, out *login.TokenCredentialRequestList, s conversion.Scope) error {
	return autoConvert_v1alpha1_TokenCredentialRequestList_To_login_TokenCredentialRequestList(in, out, s)
}

func autoConvert_login_TokenCredentialRequestList_To_v1alpha1_TokenCredentialRequestList(in *login.TokenCredentialRequestList, out *TokenCredentialRequestList, s conversion.Scope) error {
	out.ListMeta = in.ListMeta
	out.Items = *(*[]TokenCredentialRequest)(unsafe.Pointer(&in.Items))
	return nil
}

// Convert_login_TokenCredentialRequestList_To_v1alpha1_TokenCredentialRequestList is an autogenerated conversion function.
func Convert_login_TokenCredentialRequestList_To_v1alpha1_TokenCredentialRequestList(in *login.TokenCredentialRequestList, out *TokenCredentialRequestList, s conversion.Scope) error {
	return autoConvert_login_TokenCredentialRequestList_To_v1alpha1_TokenCredentialRequestList(in, out, s)
}

func autoConvert_v1alpha1_TokenCredentialRequestSpec_To_login_TokenCredentialRequestSpec(in *TokenCredentialRequestSpec, out *login.TokenCredentialRequestSpec, s conversion.Scope) error {
	out.Token = in.Token
	out.Authenticator = in.Authenticator
	return nil
}

// Convert_v1alpha1_TokenCredentialRequestSpec_To_login_TokenCredentialRequestSpec is an autogenerated conversion function.
func Convert_v1alpha1_TokenCredentialRequestSpec_To_login_TokenCredentialRequestSpec(in *TokenCredentialRequestSpec, out *login.TokenCredentialRequestSpec, s conversion.Scope) error {
	return autoConvert_v1alpha1_TokenCredentialRequestSpec_To_login_TokenCredentialRequestSpec(in, out, s)
}

func autoConvert_login_TokenCredentialRequestSpec_To_v1alpha1_TokenCredentialRequestSpec(in *login.TokenCredentialRequestSpec, out *TokenCredentialRequestSpec, s conversion.Scope) error {
	out.Token = in.Token
	out.Authenticator = in.Authenticator
	return nil
}

// Convert_login_TokenCredentialRequestSpec_To_v1alpha1_TokenCredentialRequestSpec is an autogenerated conversion function.
func Convert_login_TokenCredentialRequestSpec_To_v1alpha1_TokenCredentialRequestSpec(in *login.TokenCredentialRequestSpec, out *TokenCredentialRequestSpec, s conversion.Scope) error {
	return autoConvert_login_TokenCredentialRequestSpec_To_v1alpha1_TokenCredentialRequestSpec(in, out, s)
}

func autoConvert_v1alpha1_TokenCredentialRequestStatus_To_login_TokenCredentialRequestStatus(in *TokenCredentialRequestStatus, out *login.TokenCredentialRequestStatus, s conversion.Scope) error {
	out.Credential = (*login.ClusterCredential)(unsafe.Pointer(in.Credential))
	out.Message = (*string)(unsafe.Pointer(in.Message))
	return nil
}

// Convert_v1alpha1_TokenCredentialRequestStatus_To_login_TokenCredentialRequestStatus is an autogenerated conversion function.
func Convert_v1alpha1_TokenCredentialRequestStatus_To_login_TokenCredentialRequestStatus(in *TokenCredentialRequestStatus, out *login.TokenCredentialRequestStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_TokenCredentialRequestStatus_To_login_TokenCredentialRequestStatus(in, out, s)
}

func autoConvert_login_TokenCredentialRequestStatus_To_v1alpha1_TokenCredentialRequestStatus(in *login.TokenCredentialRequestStatus, out *TokenCredentialRequestStatus, s conversion.Scope) error {
	out.Credential = (*ClusterCredential)(unsafe.Pointer(in.Credential))
	out.Message = (*string)(unsafe.Pointer(in.Message))
	return nil
}

// Convert_login_TokenCredentialRequestStatus_To_v1alpha1_TokenCredentialRequestStatus is an autogenerated conversion function.
func Convert_login_TokenCredentialRequestStatus_To_v1alpha1_TokenCredentialRequestStatus(in *login.TokenCredentialRequestStatus, out *TokenCredentialRequestStatus, s conversion.Scope) error {
	return autoConvert_login_TokenCredentialRequestStatus_To_v1alpha1_TokenCredentialRequestStatus(in, out, s)
}
