package dssapi

import "time"

type Project struct {
	ProjectKey  string          `json:"projectKey"`
	ProjectType string          `json:"projectType"`
	Name        string          `json:"name"`
	CreationTag modificationTag `json:"creationTag"`
	VersionTag  modificationTag `json:"modificationTag"`
}

func (project *Project) DateCreated() time.Time {
	return time.UnixMilli(project.CreationTag.LastModifiedOn)
}

func (project *Project) CreatedBy() string {
	return project.CreationTag.LastModifiedBy.Login
}

func (project *Project) DateModified() time.Time {
	return time.UnixMilli(project.VersionTag.LastModifiedOn)
}

func (project *Project) ModifiedBy() string {
	return project.VersionTag.LastModifiedBy.Login
}
