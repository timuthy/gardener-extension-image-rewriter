// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/ptr"

	"github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/v1alpha1"
	. "github.com/gardener/gardener-extension-image-rewriter/pkg/apis/config/validation"
)

var _ = Describe("Validation", func() {
	var config *v1alpha1.Configuration

	Describe("#ValidateConfiguration", func() {
		BeforeEach(func() {
			config = &v1alpha1.Configuration{}
		})

		Context("Containerd", func() {
			It("should pass with valid configuration", func() {
				config.Containerd = []v1alpha1.ContainerdConfiguration{{
					Upstream: "example.com",
					Server:   "https://example.com",
					Hosts: []v1alpha1.ContainerdHostConfig{{
						URL:      "https://host.example.com",
						Provider: "local",
					}},
				}}

				Expect(ValidateConfiguration(config)).To(BeEmpty())
			})

			It("should validate upstream is set", func() {
				config.Containerd = []v1alpha1.ContainerdConfiguration{{
					Server: "https://example.com",
				}}

				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("containerd[0].upstream"),
				}))))
			})

			It("should validate server URL is set", func() {
				config.Containerd = []v1alpha1.ContainerdConfiguration{{
					Upstream: "example.com",
					Server:   "",
				}}

				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("containerd[0].server"),
				}))))
			})

			It("should validate server URL is valid", func() {
				config.Containerd = []v1alpha1.ContainerdConfiguration{{
					Upstream: "example.com",
					Server:   ":example",
				}}

				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("containerd[0].server"),
				}))))
			})

			It("should validate hosts URL is valid", func() {
				config.Containerd = []v1alpha1.ContainerdConfiguration{{
					Upstream: "example.com",
					Server:   "https://example.com",
					Hosts: []v1alpha1.ContainerdHostConfig{{
						URL:      ":example",
						Provider: "local",
					}},
				}}

				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("containerd[0].hosts[0].url"),
				}))))
			})

			It("should validate hosts provider is set", func() {
				config.Containerd = []v1alpha1.ContainerdConfiguration{{
					Upstream: "example.com",
					Server:   "https://example.com",
					Hosts: []v1alpha1.ContainerdHostConfig{{
						URL:      "https://example.com",
						Provider: "",
					}},
				}}

				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("containerd[0].hosts[0].provider"),
				}))))
			})

			It("should validate hosts cloud profile is not empty", func() {
				config.Containerd = []v1alpha1.ContainerdConfiguration{{
					Upstream: "example.com",
					Server:   "https://example.com",
					Hosts: []v1alpha1.ContainerdHostConfig{{
						URL:           "https://example.com",
						Provider:      "local",
						CloudProfiles: []string{""},
					}},
				}}

				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("containerd[0].hosts[0].cloudProfiles[0]"),
				}))))
			})

			It("should validate hosts region is not empty", func() {
				config.Containerd = []v1alpha1.ContainerdConfiguration{{
					Upstream: "example.com",
					Server:   "https://example.com",
					Hosts: []v1alpha1.ContainerdHostConfig{{
						URL:      "https://example.com",
						Provider: "local",
						Regions:  []string{""},
					}},
				}}

				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("containerd[0].hosts[0].regions[0]"),
				}))))
			})
		})

		Context("Overwrites", func() {
			BeforeEach(func() {
				config = &v1alpha1.Configuration{
					Overwrites: []v1alpha1.ImageOverwrite{{}},
				}
			})

			It("should validate source and target are set", func() {
				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("overwrites[0].source"),
				})), PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("overwrites[0].targets"),
				}))))
			})

			It("should validate only image or prefix is used for source", func() {
				config.Overwrites[0].Source = v1alpha1.Image{
					Image:  ptr.To("foo/bar:latest"),
					Prefix: ptr.To("foo"),
				}

				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeForbidden),
					"Field": Equal("overwrites[0].source"),
				})), PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("overwrites[0].targets"),
				}))))
			})

			It("should validate target has required fields", func() {
				config.Overwrites[0].Source.Image = ptr.To("foo/bar:latest")
				config.Overwrites[0].Targets = []v1alpha1.TargetConfiguration{{}}

				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("overwrites[0].targets[0].image"),
				})), PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("overwrites[0].targets[0].provider"),
				}))))
			})

			It("should validate only prefix is configured", func() {
				config.Overwrites[0].Source.Prefix = ptr.To("foo")
				config.Overwrites[0].Targets = []v1alpha1.TargetConfiguration{{
					Image: v1alpha1.Image{
						Image: ptr.To("foo/bar:latest"),
					},
					Provider: "local",
					Regions:  []string{"local"},
				}}

				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("overwrites[0].targets[0].prefix"),
				})), PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeForbidden),
					"Field": Equal("overwrites[0].targets[0].image"),
				}))))
			})

			It("should validate only image is configured", func() {
				config.Overwrites[0].Source.Image = ptr.To("foo/bar:latest")
				config.Overwrites[0].Targets = []v1alpha1.TargetConfiguration{{
					Image: v1alpha1.Image{
						Prefix: ptr.To("foo"),
					},
					Provider: "local",
					Regions:  []string{"local"},
				}}

				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeRequired),
					"Field": Equal("overwrites[0].targets[0].image"),
				})), PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeForbidden),
					"Field": Equal("overwrites[0].targets[0].prefix"),
				}))))
			})

			It("should validate cloudprofile not empty", func() {
				config.Overwrites[0].Source.Image = ptr.To("foo/bar:latest")
				config.Overwrites[0].Targets = []v1alpha1.TargetConfiguration{{
					Image: v1alpha1.Image{
						Image: ptr.To("foo/bar:latest"),
					},
					Provider:      "local",
					CloudProfiles: []string{""},
				}}

				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("overwrites[0].targets[0].cloudProfiles[0]"),
				}))))
			})

			It("should validate region is not empty", func() {
				config.Overwrites[0].Source.Image = ptr.To("foo/bar:latest")
				config.Overwrites[0].Targets = []v1alpha1.TargetConfiguration{{
					Image: v1alpha1.Image{
						Image: ptr.To("foo/bar:latest"),
					},
					Provider: "local",
					Regions:  []string{""},
				}}

				Expect(ValidateConfiguration(config)).To(ConsistOf(PointTo(MatchFields(IgnoreExtras, Fields{
					"Type":  Equal(field.ErrorTypeInvalid),
					"Field": Equal("overwrites[0].targets[0].regions[0]"),
				}))))
			})
		})
	})
})
