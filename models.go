package main

type TimePeriod struct {
	StartDateTime string `json:"startDateTime,omitempty"`
	EndDateTime   string `json:"endDateTime,omitempty"`
}

type Catalog struct {
	ID              string      `json:"id"`
	Href            string      `json:"href"`
	Name            string      `json:"name"`
	Description     string      `json:"description,omitempty"`
	LifecycleStatus string      `json:"lifecycleStatus,omitempty"`
	LastUpdate      string      `json:"lastUpdate,omitempty"`
	AtType          string      `json:"@type,omitempty"`
	AtBaseType      string      `json:"@baseType,omitempty"`
	ValidFor        *TimePeriod `json:"validFor,omitempty"`
}

type ProductOffering struct {
	ID              string      `json:"id"`
	Href            string      `json:"href"`
	Name            string      `json:"name"`
	Description     string      `json:"description,omitempty"`
	LifecycleStatus string      `json:"lifecycleStatus,omitempty"`
	LastUpdate      string      `json:"lastUpdate,omitempty"`
	AtType          string      `json:"@type,omitempty"`
	AtBaseType      string      `json:"@baseType,omitempty"`
	IsBundle        *bool       `json:"isBundle,omitempty"`
	ValidFor        *TimePeriod `json:"validFor,omitempty"`
}
