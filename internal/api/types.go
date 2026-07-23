package api

type Me struct {
	Authenticated bool   `json:"authenticated"`
	Email         string `json:"email"`
	UserID        string `json:"userId"`
	IsAdmin       bool   `json:"isAdmin"`
}

type Technology struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Category string `json:"category"`
}

type Project struct {
	ID           int          `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Technologies []Technology `json:"technologies"`
}

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type Suggestion struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type AddTechnologyRequest struct {
	Name     string `json:"name"`
	Version  string `json:"version,omitempty"`
	Category string `json:"category"`
}

type ManifestFile struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}

type UploadManifestsRequest struct {
	Ecosystem string         `json:"ecosystem"`
	GroupName string         `json:"groupName"`
	Files     []ManifestFile `json:"files"`
}

type DependencyGroupInfo struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Ecosystem       string `json:"ecosystem"`
	DependencyCount int    `json:"dependencyCount"`
}

type DependencySummary struct {
	Groups          []DependencyGroupInfo `json:"groups"`
	VulnerableCount int                   `json:"vulnerableCount"`
	TotalCount      int                   `json:"totalCount"`
}

type UploadManifestsResponse struct {
	Summary          DependencySummary `json:"summary"`
	UnsupportedFiles []string          `json:"unsupportedFiles"`
	EmptyFiles       []string          `json:"emptyFiles"`
}

type ProjectStats struct {
	TechnologyCount           int `json:"technologyCount"`
	EndOfLifeCount            int `json:"endOfLifeCount"`
	TechnologyCveCount        int `json:"technologyCveCount"`
	DependencyCount           int `json:"dependencyCount"`
	VulnerableDependencyCount int `json:"vulnerableDependencyCount"`
	DependencyCveCount        int `json:"dependencyCveCount"`
}

type DeviceAuthorization struct {
	DeviceCode              string `json:"deviceCode"`
	UserCode                string `json:"userCode"`
	VerificationURI         string `json:"verificationUri"`
	VerificationURIComplete string `json:"verificationUriComplete"`
	ExpiresInSeconds        int    `json:"expiresInSeconds"`
	IntervalSeconds         int    `json:"intervalSeconds"`
}

type DeviceToken struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
}
