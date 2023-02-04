package dssapi

import (
	"strings"
	"time"
)

type Recipe struct {
	ProjectKey  string          `json:"projectKey"`
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	VersionTag  modificationTag `json:"versionTag"`
	CreationTag modificationTag `json:"creationTag"`
}

var recipeScriptTypes = []string{"python", "scala", "r"}

func (recipe *Recipe) CreatedBy() string {
	return recipe.CreationTag.LastModifiedBy.Login
}

func (recipe *Recipe) CreatedOn() time.Time {
	return time.UnixMilli(recipe.CreationTag.LastModifiedOn)
}

func (recipe *Recipe) ModifiedBy() string {
	return recipe.VersionTag.LastModifiedBy.Login
}

func (recipe *Recipe) ModifiedOn() time.Time {
	return time.UnixMilli(recipe.VersionTag.LastModifiedOn)
}

func (recipe *Recipe) IsEditable() bool {
	_type := strings.ToLower(recipe.Type)

	for _, recipeType := range recipeScriptTypes {
		if _type == recipeType {
			return true
		}
	}

	return false
}
