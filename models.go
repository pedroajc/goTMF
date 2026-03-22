// models.go
package main

type TimePeriod struct {
	StartDateTime string `json:"startDateTime,omitempty"`
	EndDateTime   string `json:"endDateTime,omitempty"`
}

type Catalog struct {
	ID              string        `json:"id"`
	Href            string        `json:"href"`
	Name            string        `json:"name"`
	Description     string        `json:"description,omitempty"`
	LifecycleStatus string        `json:"lifecycleStatus,omitempty"`
	LastUpdate      string        `json:"lastUpdate,omitempty"`
	AtType          string        `json:"@type,omitempty"`
	AtBaseType      string        `json:"@baseType,omitempty"`
	ValidFor        *TimePeriod   `json:"validFor,omitempty"`
	Category        []CategoryRef `json:"category,omitempty"`
}

type Category struct {
	ID              string               `json:"id"`
	Href            string               `json:"href"`
	Name            string               `json:"name"`
	Description     string               `json:"description,omitempty"`
	LifecycleStatus string               `json:"lifecycleStatus,omitempty"`
	LastUpdate      string               `json:"lastUpdate,omitempty"`
	AtType          string               `json:"@type,omitempty"`
	AtBaseType      string               `json:"@baseType,omitempty"`
	ValidFor        *TimePeriod          `json:"validFor,omitempty"`
	ProductOffering []ProductOfferingRef `json:"productOffering,omitempty"`
}

type ProductOffering struct {
	ID                   string                   `json:"id"`
	Href                 string                   `json:"href"`
	Name                 string                   `json:"name"`
	Description          string                   `json:"description,omitempty"`
	LifecycleStatus      string                   `json:"lifecycleStatus,omitempty"`
	LastUpdate           string                   `json:"lastUpdate,omitempty"`
	AtType               string                   `json:"@type,omitempty"`
	AtBaseType           string                   `json:"@baseType,omitempty"`
	IsBundle             *bool                    `json:"isBundle,omitempty"`
	ValidFor             *TimePeriod              `json:"validFor,omitempty"`
	ProductSpecification *ProductSpecificationRef `json:"productSpecification,omitempty"`
}

type ProductSpecification struct {
	ID              string      `json:"id"`
	Href            string      `json:"href"`
	Name            string      `json:"name"`
	Description     string      `json:"description,omitempty"`
	LifecycleStatus string      `json:"lifecycleStatus,omitempty"`
	LastUpdate      string      `json:"lastUpdate,omitempty"`
	AtType          string      `json:"@type,omitempty"`
	AtBaseType      string      `json:"@baseType,omitempty"`
	ValidFor        *TimePeriod `json:"validFor,omitempty"`
	Brand           string      `json:"brand,omitempty"`
	Version         string      `json:"version,omitempty"`
}

type Error struct {
	Code           string `json:"code"`
	Reason         string `json:"reason"`
	Message        string `json:"message,omitempty"`
	Status         string `json:"status,omitempty"`
	ReferenceError string `json:"referenceError,omitempty"`
	AtType         string `json:"@type,omitempty"`
	AtBaseType     string `json:"@baseType,omitempty"`
}

type CategoryRef struct {
	ID   string `json:"id"`
	Href string `json:"href"`
	Name string `json:"name"`
}

type ProductOfferingRef struct {
	ID   string `json:"id"`
	Href string `json:"href"`
	Name string `json:"name"`
}

type ProductSpecificationRef struct {
	ID      string `json:"id"`
	Href    string `json:"href"`
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}
