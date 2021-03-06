/*
Copyright 2018 The Kubernetes Authors.

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

package shareoptions

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/extensions/trusts"
	"k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/cloud-provider-openstack/pkg/share/manila/shareoptions/validator"
)

// OpenStackOptions contains fields used for authenticating to OpenStack
type OpenStackOptions struct {
	// Common options

	OSAuthURL    string `name:"os-authURL" dependsOn:"os-password|os-trustID"`
	OSRegionName string `name:"os-region"`

	OSCertAuthority string `name:"os-certAuthority" value:"optional"`
	OSTLSInsecure   string `name:"os-TLSInsecure" value:"optional" dependsOn:"os-certAuthority" matches:"^true|false$"`

	// User authentication

	OSPassword string `name:"os-password" value:"optional" dependsOn:"os-domainID|os-domainName,os-projectID|os-projectName,os-userID|os-userName"`
	OSUserID   string `name:"os-userID" value:"optional" dependsOn:"os-password"`
	OSUsername string `name:"os-userName" value:"optional" dependsOn:"os-password"`

	OSDomainID   string `name:"os-domainID" value:"optional" dependsOn:"os-password"`
	OSDomainName string `name:"os-domainName" value:"optional" dependsOn:"os-password"`

	OSProjectID   string `name:"os-projectID" value:"optional" dependsOn:"os-password"`
	OSProjectName string `name:"os-projectName" value:"optional" dependsOn:"os-password"`

	// Trustee authentication

	OSTrustID         string `name:"os-trustID" value:"optional" dependsOn:"os-trusteeID,os-trusteePassword"`
	OSTrusteeID       string `name:"os-trusteeID" value:"optional" dependsOn:"os-trustID"`
	OSTrusteePassword string `name:"os-trusteePassword" value:"optional" dependsOn:"os-trustID"`
}

var (
	osOptionsValidator = validator.New(&OpenStackOptions{})
)

// NewOpenStackOptionsFromSecret reads k8s secrets, validates and populates OpenStackOptions
func NewOpenStackOptionsFromSecret(c clientset.Interface, secretRef *v1.SecretReference) (*OpenStackOptions, error) {
	params, err := readSecrets(c, secretRef)
	if err != nil {
		return nil, err
	}

	return NewOpenStackOptionsFromMap(params)
}

// NewOpenStackOptionsFromMap validates and populates OpenStackOptions
func NewOpenStackOptionsFromMap(params map[string]string) (*OpenStackOptions, error) {
	opts := &OpenStackOptions{}
	return opts, osOptionsValidator.Populate(params, opts)
}

// ToAuthOptions converts OpenStackOptions to gophercloud.AuthOptions
func (o *OpenStackOptions) ToAuthOptions() *gophercloud.AuthOptions {
	authOpts := &gophercloud.AuthOptions{
		IdentityEndpoint: o.OSAuthURL,
		UserID:           o.OSUserID,
		Username:         o.OSUsername,
		Password:         o.OSPassword,
		TenantID:         o.OSProjectID,
		TenantName:       o.OSProjectName,
		DomainID:         o.OSDomainID,
		DomainName:       o.OSDomainName,
	}

	if o.OSTrustID != "" {
		// gophercloud doesn't have dedicated options for trustee credentials
		authOpts.UserID = o.OSTrusteeID
		authOpts.Password = o.OSTrusteePassword
	}

	return authOpts
}

// ToAuthOptionsExt converts OpenStackOptions to trusts.AuthOptsExt
func (o *OpenStackOptions) ToAuthOptionsExt() *trusts.AuthOptsExt {
	return &trusts.AuthOptsExt{
		AuthOptionsBuilder: o.ToAuthOptions(),
		TrustID:            o.OSTrustID,
	}
}
