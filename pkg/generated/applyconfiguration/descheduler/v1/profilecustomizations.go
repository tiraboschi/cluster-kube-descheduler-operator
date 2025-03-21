// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

import (
	deschedulerv1 "github.com/openshift/cluster-kube-descheduler-operator/pkg/apis/descheduler/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProfileCustomizationsApplyConfiguration represents a declarative configuration of the ProfileCustomizations type for use
// with apply.
type ProfileCustomizationsApplyConfiguration struct {
	PodLifetime                      *metav1.Duration                                 `json:"podLifetime,omitempty"`
	ThresholdPriority                *int32                                           `json:"thresholdPriority,omitempty"`
	ThresholdPriorityClassName       *string                                          `json:"thresholdPriorityClassName,omitempty"`
	Namespaces                       *NamespacesApplyConfiguration                    `json:"namespaces,omitempty"`
	DevLowNodeUtilizationThresholds  *deschedulerv1.LowNodeUtilizationThresholdsType  `json:"devLowNodeUtilizationThresholds,omitempty"`
	DevEnableSoftTainter             *bool                                            `json:"devEnableSoftTainter,omitempty"`
	DevKubevirtSchedulable           *bool                                            `json:"devKubevirtSchedulable,omitempty"`
	DevEnableEvictionsInBackground   *bool                                            `json:"devEnableEvictionsInBackground,omitempty"`
	DevHighNodeUtilizationThresholds *deschedulerv1.HighNodeUtilizationThresholdsType `json:"devHighNodeUtilizationThresholds,omitempty"`
	DevActualUtilizationProfile      *deschedulerv1.ActualUtilizationProfile          `json:"devActualUtilizationProfile,omitempty"`
}

// ProfileCustomizationsApplyConfiguration constructs a declarative configuration of the ProfileCustomizations type for use with
// apply.
func ProfileCustomizations() *ProfileCustomizationsApplyConfiguration {
	return &ProfileCustomizationsApplyConfiguration{}
}

// WithPodLifetime sets the PodLifetime field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the PodLifetime field is set to the value of the last call.
func (b *ProfileCustomizationsApplyConfiguration) WithPodLifetime(value metav1.Duration) *ProfileCustomizationsApplyConfiguration {
	b.PodLifetime = &value
	return b
}

// WithThresholdPriority sets the ThresholdPriority field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ThresholdPriority field is set to the value of the last call.
func (b *ProfileCustomizationsApplyConfiguration) WithThresholdPriority(value int32) *ProfileCustomizationsApplyConfiguration {
	b.ThresholdPriority = &value
	return b
}

// WithThresholdPriorityClassName sets the ThresholdPriorityClassName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ThresholdPriorityClassName field is set to the value of the last call.
func (b *ProfileCustomizationsApplyConfiguration) WithThresholdPriorityClassName(value string) *ProfileCustomizationsApplyConfiguration {
	b.ThresholdPriorityClassName = &value
	return b
}

// WithNamespaces sets the Namespaces field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Namespaces field is set to the value of the last call.
func (b *ProfileCustomizationsApplyConfiguration) WithNamespaces(value *NamespacesApplyConfiguration) *ProfileCustomizationsApplyConfiguration {
	b.Namespaces = value
	return b
}

// WithDevLowNodeUtilizationThresholds sets the DevLowNodeUtilizationThresholds field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DevLowNodeUtilizationThresholds field is set to the value of the last call.
func (b *ProfileCustomizationsApplyConfiguration) WithDevLowNodeUtilizationThresholds(value deschedulerv1.LowNodeUtilizationThresholdsType) *ProfileCustomizationsApplyConfiguration {
	b.DevLowNodeUtilizationThresholds = &value
	return b
}

// WithDevEnableSoftTainter sets the DevEnableSoftTainter field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DevEnableSoftTainter field is set to the value of the last call.
func (b *ProfileCustomizationsApplyConfiguration) WithDevEnableSoftTainter(value bool) *ProfileCustomizationsApplyConfiguration {
	b.DevEnableSoftTainter = &value
	return b
}

// WithDevKubevirtSchedulable sets the DevKubevirtSchedulable field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DevKubevirtSchedulable field is set to the value of the last call.
func (b *ProfileCustomizationsApplyConfiguration) WithDevKubevirtSchedulable(value bool) *ProfileCustomizationsApplyConfiguration {
	b.DevKubevirtSchedulable = &value
	return b
}

// WithDevEnableEvictionsInBackground sets the DevEnableEvictionsInBackground field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DevEnableEvictionsInBackground field is set to the value of the last call.
func (b *ProfileCustomizationsApplyConfiguration) WithDevEnableEvictionsInBackground(value bool) *ProfileCustomizationsApplyConfiguration {
	b.DevEnableEvictionsInBackground = &value
	return b
}

// WithDevHighNodeUtilizationThresholds sets the DevHighNodeUtilizationThresholds field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DevHighNodeUtilizationThresholds field is set to the value of the last call.
func (b *ProfileCustomizationsApplyConfiguration) WithDevHighNodeUtilizationThresholds(value deschedulerv1.HighNodeUtilizationThresholdsType) *ProfileCustomizationsApplyConfiguration {
	b.DevHighNodeUtilizationThresholds = &value
	return b
}

// WithDevActualUtilizationProfile sets the DevActualUtilizationProfile field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the DevActualUtilizationProfile field is set to the value of the last call.
func (b *ProfileCustomizationsApplyConfiguration) WithDevActualUtilizationProfile(value deschedulerv1.ActualUtilizationProfile) *ProfileCustomizationsApplyConfiguration {
	b.DevActualUtilizationProfile = &value
	return b
}
