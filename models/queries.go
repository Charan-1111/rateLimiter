package models

type Queries struct {
	Fetch Fetch `json:"fetch"`
}

type Fetch struct {
	FetchPolicies string `json:"fetchPolicies"`
}