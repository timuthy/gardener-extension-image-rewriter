// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"net/url"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
)

// ValidateConfiguration validates the passed configuration object.
func ValidateConfiguration(config *v1alpha1.Configuration) field.ErrorList {
	allErrs := field.ErrorList{}

	allErrs = append(allErrs, validateContainerd(config.Containerd, field.NewPath("containerd"))...)
	allErrs = append(allErrs, validateOverwrites(config.Overwrites, field.NewPath("overwrites"))...)

	return allErrs
}

func validateContainerd(config []v1alpha1.ContainerdConfiguration, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for i, containerd := range config {
		fldContainerd := fldPath.Index(i)

		if containerd.Upstream == "" {
			allErrs = append(allErrs, field.Required(fldContainerd.Child("upstream"), "upstream must be specified"))
		}

		allErrs = append(allErrs, validateURL(containerd.Server, fldContainerd.Child("server"))...)

		for j, host := range containerd.Hosts {
			fldHost := fldContainerd.Child("hosts").Index(j)

			allErrs = append(allErrs, validateURL(host.URL, fldHost.Child("url"))...)

			if host.Provider == "" {
				allErrs = append(allErrs, field.Required(fldHost.Child("provider"), "provider must be specified"))
			}

			for k, cloudProfile := range host.CloudProfiles {
				if cloudProfile == "" {
					allErrs = append(allErrs, field.Invalid(fldHost.Child("cloudProfiles").Index(k), cloudProfile, "cloudProfile must not be empty"))
				}
			}

			for k, region := range host.Regions {
				if region == "" {
					allErrs = append(allErrs, field.Invalid(fldHost.Child("regions").Index(k), region, "region must not be empty"))
				}
			}
		}
	}

	return allErrs
}

func validateOverwrites(overwrites []v1alpha1.ImageOverwrite, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	for i, overwrite := range overwrites {
		fldOverwrites := fldPath.Index(i)

		if overwrite.Source.Image == nil && overwrite.Source.Prefix == nil {
			allErrs = append(allErrs, field.Required(fldOverwrites.Child("source"), "either 'image' or 'prefix' must be set"))
		}
		if overwrite.Source.Image != nil && overwrite.Source.Prefix != nil {
			allErrs = append(allErrs, field.Forbidden(fldOverwrites.Child("source"), "only one of 'image' or 'prefix' can be set"))
		}
		if len(overwrite.Targets) == 0 {
			allErrs = append(allErrs, field.Required(fldOverwrites.Child("targets"), "at least one target must be specified"))
		}
		for j, target := range overwrite.Targets {
			fldTarget := fldOverwrites.Child("targets").Index(j)

			switch {
			case target.Image.Image == nil && target.Prefix == nil:
				allErrs = append(allErrs, field.Required(fldTarget.Child("image"), "either 'image' or 'prefix' must be set"))
			case target.Prefix != nil && target.Image.Image != nil:
				allErrs = append(allErrs, field.Forbidden(fldTarget.Child("image"), "only one of 'image' or 'prefix' can be set"))
			case overwrite.Source.Prefix != nil:
				if target.Prefix == nil {
					allErrs = append(allErrs, field.Required(fldTarget.Child("prefix"), "target must use 'prefix' when source 'prefix' is set"))
				}
				if target.Image.Image != nil {
					allErrs = append(allErrs, field.Forbidden(fldTarget.Child("image"), "target 'image' must not be set when source 'prefix' is set"))
				}
			case overwrite.Source.Image != nil:
				if target.Image.Image == nil {
					allErrs = append(allErrs, field.Required(fldTarget.Child("image"), "target 'image' must be set when source 'image' is set"))
				}
				if target.Prefix != nil {
					allErrs = append(allErrs, field.Forbidden(fldTarget.Child("prefix"), "target must not set 'prefix' when source 'image' is set"))
				}
			}

			if target.Provider == "" {
				allErrs = append(allErrs, field.Required(fldTarget.Child("provider"), "provider must be specified"))
			}

			for k, cloudProfile := range target.CloudProfiles {
				if cloudProfile == "" {
					allErrs = append(allErrs, field.Invalid(fldTarget.Child("cloudProfiles").Index(k), cloudProfile, "cloudProfile must not be empty"))
				}
			}
			for k, region := range target.Regions {
				if region == "" {
					allErrs = append(allErrs, field.Invalid(fldTarget.Child("regions").Index(k), region, "region must not be empty"))
				}
			}
		}
	}

	return allErrs
}

func validateURL(urlString string, fldPath *field.Path) field.ErrorList {
	allErrs := field.ErrorList{}

	if urlString == "" {
		allErrs = append(allErrs, field.Required(fldPath, "server must be specified"))
	} else {
		url, err := url.Parse(urlString)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath, url, "must be a valid URL"))
		} else if url.Scheme == "" {
			allErrs = append(allErrs, field.Invalid(fldPath, url, "must have a scheme"))
		}
	}

	return allErrs
}
