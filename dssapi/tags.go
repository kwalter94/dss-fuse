package dssapi

type modificationTag struct {
	LastModifiedOn int64 `json:"lastModifiedOn"`
	LastModifiedBy user  `json:"lastModifiedBy"`
}

type user struct {
	Login string `json:"login"`
}
