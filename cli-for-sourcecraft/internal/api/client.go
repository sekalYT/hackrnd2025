// internal/api/client.go
package api

import (
	"bytes"
	cliutils "cli-for-sourcecraft/internal/utils"
	"crypto/tls" // For disabling TLS verification
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxRetries = 3
const initialRetryDelay = 1 * time.Second

// Client for the SourceCraft API
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
}

// User struct (from Swagger definition, potentially incomplete)
// Used for Owner field in Repo and RepositoryEmbedded, and Author in PullRequest
type User struct {
	ID   *string `json:"id"`   // Swagger shows string ID for UserEmbedded
	Slug *string `json:"slug"` // Swagger uses 'slug' for UserEmbedded
	// Add other fields if needed from a full User definition if available
}

// NewClient constructor
func NewClient(baseURL string, token string) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 10 * time.Second, Transport: tr},
		Token:      token,
	}
}

// makeRequest private helper for API calls
func (c *Client) makeRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("JSON marshal error: %w", err)
		}
	}
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	fullURL := c.BaseURL + path

	for i := 0; i < maxRetries; i++ {

		// 1. Создание запроса
		// Создаем новый Reader для тела при каждой попытке, так как он будет прочитан
		req, err := http.NewRequest(method, fullURL, bytes.NewBuffer(reqBody))
		if err != nil {
			return nil, fmt.Errorf("request creation error: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		// 2. Выполнение запроса
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			// Ошибка сети/таймаута: повторяем с экспоненциальной задержкой
			if i < maxRetries-1 {
				waitTime := initialRetryDelay * time.Duration(1<<i)
				fmt.Printf("Warning: Request failed (network/timeout). Retrying in %v (Attempt %d/%d)...\n", waitTime, i+1, maxRetries)
				time.Sleep(waitTime)
				continue
			}
			return nil, fmt.Errorf("request execution error after %d retries: %w", maxRetries, err)
		}
		defer resp.Body.Close()

		// 3. Обработка Rate Limit (429)
		if resp.StatusCode == http.StatusTooManyRequests { // 429
			var waitTime time.Duration
			retryAfterStr := resp.Header.Get("Retry-After")

			if retryAfterStr != "" {
				// Пытаемся парсить как секунды
				if seconds, parseErr := time.ParseDuration(retryAfterStr + "s"); parseErr == nil {
					waitTime = seconds
				} else {
					// Если не секунды, используем экспоненциальный backoff
					waitTime = initialRetryDelay * time.Duration(1<<i)
				}
			} else {
				// Если заголовок отсутствует, используем экспоненциальный backoff
				waitTime = initialRetryDelay * time.Duration(1<<i)
			}

			if i < maxRetries-1 {
				// Ждем и повторяем
				fmt.Printf("Warning: Rate limit hit (429). Retrying in %v (Attempt %d/%d)...\n", waitTime, i+1, maxRetries)
				time.Sleep(waitTime)
				continue
			}

			// Если это была последняя попытка
			return nil, fmt.Errorf("API request failed due to rate limiting (429) after %d retries", maxRetries)
		}

		// 4. Чтение тела ответа (происходит только один раз при успешном запросе или не-429 ошибке)
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("API response read error: %w", err)
		}

		// 5. Обработка других ошибок (4xx, 5xx)
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			errMsg := fmt.Sprintf("API returned error: %s (path: %s)", resp.Status, path)
			if resp.StatusCode == http.StatusUnauthorized {
				errMsg = fmt.Sprintf("API error: %s (Invalid or expired token, path: %s)", resp.Status, path)
			} else if resp.StatusCode == http.StatusNotFound {
				errMsg = fmt.Sprintf("API error: %s (Incorrect path: %s)", resp.Status, path)
			} else if resp.StatusCode == http.StatusMethodNotAllowed {
				errMsg = fmt.Sprintf("API error: %s (Wrong HTTP method for path: %s)", resp.Status, path)
			}
			if len(respBody) > 0 {
				snippet := string(respBody)
				if len(snippet) > 500 {
					snippet = snippet[:500] + "..."
				}
				var jsonError map[string]interface{}
				if json.Unmarshal(respBody, &jsonError) == nil {
					if msg, ok := jsonError["message"].(string); ok {
						if reqID, ok := jsonError["request_id"].(string); ok {
							errMsg = fmt.Sprintf("%s. Message: %s (Request ID: %s)", errMsg, msg, reqID)
						} else {
							errMsg = fmt.Sprintf("%s. Message: %s", errMsg, msg)
						}
					}
				} else {
					errMsg = fmt.Sprintf("%s. Response Body: %s", errMsg, snippet)
				}
			}
			return nil, fmt.Errorf(errMsg)
		}

		// 6. Успешный ответ (2xx)
		contentType := resp.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			snippet := string(respBody)
			if len(snippet) > 150 {
				snippet = snippet[:150] + "..."
			}
			return nil, fmt.Errorf(
				"API returned non-JSON response (Content-Type: %s) despite 2xx status. Path: %s. Body start: %s",
				contentType, path, snippet,
			)
		}
		return respBody, nil
	}

	return nil, fmt.Errorf("internal error: retry loop failed to complete")
}

type SetDecisionBody struct {
	ReviewDecision string `json:"review_decision"` // e.g., "approve", "block"
}

// MergeParameters (как в структуре PullRequest, для использования в теле запроса, если API их примет)
type MergeParameters struct {
	Rebase       bool `json:"rebase"`
	Squash       bool `json:"squash"`
	DeleteBranch bool `json:"delete_branch"`
}

type IssueStatus struct {
	ID         *string `json:"id"`
	Slug       *string `json:"slug"`
	Name       *string `json:"name"`
	StatusType *string `json:"status_type"` // "initial", "in_progress", "paused", "completed", "cancelled"
}

// LabelEmbedded (из Swagger #/definitions/LabelEmbedded)
type LabelEmbedded struct {
	ID    *string `json:"id"`
	Slug  *string `json:"slug"`
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

// MilestoneEmbedded (из Swagger #/definitions/MilestoneEmbedded)
type MilestoneEmbedded struct {
	ID   *string `json:"id"`
	Slug *string `json:"slug"`
}

// Issue (из Swagger #/definitions/Issue)
type Issue struct {
	ID          *string            `json:"id"`
	Slug        *string            `json:"slug"`
	Title       *string            `json:"title"`
	Description *string            `json:"description"`
	Status      *IssueStatus       `json:"status"`
	Author      *User              `json:"author"`
	UpdatedBy   *User              `json:"updated_by"`
	CreatedAt   *string            `json:"created_at"`
	UpdatedAt   *string            `json:"updated_at"`
	Assignee    *User              `json:"assignee"`
	Labels      []LabelEmbedded    `json:"labels"`
	Priority    *string            `json:"priority"` // "trivial", "minor", "normal", "critical", "blocker"
	Milestone   *MilestoneEmbedded `json:"milestone"`
	Deadline    *string            `json:"deadline"`
}

// CreateIssueBody (из Swagger #/definitions/CreateIssueBody)
// Используем omitempty, чтобы не отправлять пустые поля
type CreateIssueBody struct {
	Title       string   `json:"title"`                  // Обязательно
	Description string   `json:"description,omitempty"`  // Не обязательно
	StatusSlug  string   `json:"status_slug,omitempty"`  // Не обязательно (default: "open")
	Priority    string   `json:"priority,omitempty"`     // Не обязательно
	AssigneeID  string   `json:"assignee_id,omitempty"`  // Не обязательно
	MilestoneID string   `json:"milestone_id,omitempty"` // Не обязательно
	LabelIDs    []string `json:"label_ids,omitempty"`    // Не обязательно
}

// UpdateIssueBody (из Swagger #/definitions/UpdateIssueBody)
// **ВАЖНО**: Используем *указатели*, чтобы 'omitempty' мог
// отличить 'nil' (не обновлять) от '""' (установить в пустое значение).
type UpdateIssueBody struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	StatusSlug  *string `json:"status_slug,omitempty"`
	Priority    *string `json:"priority,omitempty"`
	AssigneeID  *string `json:"assignee_id,omitempty"`
	MilestoneID *string `json:"milestone_id,omitempty"`
	// Labels и Linked PRs пока опускаем для простоты update
}

type MergeDecisionBody struct {
	ReviewDecision  string          `json:"review_decision"`  // Всегда "approve" для слияния
	MergeParameters MergeParameters `json:"merge_parameters"` // Отправляем параметры
}

type SetDecisionResponse struct {
	CreatedDecision *string `json:"created_decision"`
	PullRequestID   *string `json:"pull_request_id"`
}

// --- Repo Structures based on Swagger ---
type Repo struct {
	ID            *string             `json:"id"`
	Name          *string             `json:"name"`
	DefaultBranch *string             `json:"default_branch"`
	Slug          *string             `json:"slug"`
	IsEmpty       *bool               `json:"is_empty"`
	Description   *string             `json:"description"`
	Visibility    *string             `json:"visibility"` // "public", "internal", "private"
	CloneURL      *CloneURL           `json:"clone_url"`
	LastUpdated   *string             `json:"last_updated"` // Timestamp as string
	Language      *Language           `json:"language"`
	Owner         *User               `json:"owner"`  // Owner of *this* repo
	Parent        *RepositoryEmbedded `json:"parent"` // For sync
}

// RepositoryEmbedded represents a simplified repository, used for the 'parent' field
type RepositoryEmbedded struct {
	ID       *string   `json:"id"`
	Slug     *string   `json:"slug"`      // Slug of the parent repo
	Owner    *User     `json:"owner"`     // Owner of the parent repo
	CloneURL *CloneURL `json:"clone_url"` // Clone URLs of the parent repo
}

type CloneURL struct {
	HTTPS *string `json:"https"`
	SSH   *string `json:"ssh"`
}

type Language struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

// createRepoRequest struct based on Swagger POST /orgs/{org_slug}/repos
type createRepoRequest struct {
	Name        string `json:"name"`                  // Required
	Slug        string `json:"slug"`                  // Required
	Description string `json:"description,omitempty"` // Optional
	Visibility  string `json:"visibility,omitempty"`  // Optional ("public", "internal", "private")
}

type CreatePullRequestBody struct {
	Title        string   `json:"title"`                  // Обязательно
	SourceBranch string   `json:"source_branch"`          // Обязательно
	TargetBranch string   `json:"target_branch"`          // Обязательно
	Description  string   `json:"description,omitempty"`  // Не обязательно
	ReviewerIDs  []string `json:"reviewer_ids,omitempty"` // Не обязательно (пока ID)
	Publish      bool     `json:"publish"`                // true = Опубликовать, false = Черновик
}

// --- Pull Request Structures based on Swagger ---
type PullRequest struct {
	ID           *string `json:"id"`
	Slug         *string `json:"slug"`
	Author       *User   `json:"author"`
	Title        *string `json:"title"`
	SourceBranch *string `json:"source_branch"`
	TargetBranch *string `json:"target_branch"`
	Status       *string `json:"status"`
	UpdatedAt    *string `json:"updated_at"`
	// Добавь сюда другие поля из Swagger, если нужно будет их выводить
}

type Milestone struct {
	ID          *string `json:"id"`
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	StartDate   *string `json:"start_date"` // date-time
	Deadline    *string `json:"deadline"`   // date-time
	Status      *string `json:"status"`     // "open", "closed"
	Author      *User   `json:"author"`
	UpdatedAt   *string `json:"updated_at"` // date-time
}

// CreateMilestoneBody (из Swagger #/definitions/CreateMilestoneBody)
type CreateMilestoneBody struct {
	Name        string `json:"name"` // Обязательно
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
	StartDate   string `json:"start_date,omitempty"` // date-time
	Deadline    string `json:"deadline,omitempty"`   // date-time
}

type RepoRole string

const (
	RepoRoleViewer      RepoRole = "viewer"
	RepoRoleContributor RepoRole = "contributor"
	RepoRoleDeveloper   RepoRole = "developer"
	RepoRoleMaintainer  RepoRole = "maintainer"
	RepoRoleAdmin       RepoRole = "admin"
)

type SubjectType string

const (
	SubjectTypeUser SubjectType = "user"
	// Add other types like 'group' if they appear in Swagger later
)

// Subject (из Swagger #/definitions/Subject)
type Subject struct {
	Type SubjectType `json:"type"`
	ID   string      `json:"id"` // User ID (UUID)
}

// SubjectRole (из Swagger #/definitions/SubjectRole)
type SubjectRole struct {
	Role    RepoRole `json:"role"`
	Subject Subject  `json:"subject"`
}

// ListRepoRolesResponse (Примерная структура, Swagger не дает ее явно, но она нужна для GET /roles)
// Основана на Add/Remove responses
type ListRepoRolesResponse struct {
	SubjectRoles  []SubjectRole `json:"subject_roles"`
	NextPageToken *string       `json:"next_page_token"`
}

// AddRepoRolesBody (из Swagger #/definitions/AddRepoRolesBody)
type AddRepoRolesBody struct {
	SubjectRoles []SubjectRole `json:"subject_roles"`
}

// RemoveRepoRolesBody (из Swagger #/definitions/RemoveRepoRolesBody)
type RemoveRepoRolesBody struct {
	SubjectRoles []SubjectRole `json:"subject_roles"`
}

type RunStatus struct {
	ID        *string `json:"id"`
	Slug      *string `json:"slug"`
	Status    *string `json:"status"` // e.g., "running", "success", "failure"
	CreatedAt *string `json:"created_at"`
	UpdatedAt *string `json:"updated_at"`
	// WorkflowRuns (подразумеваемый массив для вложенных рабочих процессов)
	WorkflowRuns []WorkflowRunEmbedded `json:"workflow_runs"`
}

// WorkflowRunEmbedded represents a single workflow within a run
type WorkflowRunEmbedded struct {
	WorkflowSlug *string `json:"workflow_slug"`
	Status       *string `json:"status"`
	// TaskRuns (для детализации)
	TaskRuns []TaskRunEmbedded `json:"task_runs"`
}

// TaskRunEmbedded represents a single task within a workflow
type TaskRunEmbedded struct {
	TaskSlug *string `json:"task_slug"`
	Status   *string `json:"status"`
	// CubeRuns (для детализации логов)
	CubeRuns []CubeRunEmbedded `json:"cube_runs"`
}

// CubeRunEmbedded represents a single cube execution
type CubeRunEmbedded struct {
	CubeSlug *string `json:"cube_slug"`
	Status   *string `json:"status"`
}

// ListRunsResponse (Структура для ответа GET /cicd/runs)
type ListRunsResponse struct {
	Runs          []RunStatus `json:"runs"`
	NextPageToken *string     `json:"next_page_token"`
}

// --- API Methods ---

// ListRepositories ('src repo list') uses GET /orgs/{org_slug}/repos
func (c *Client) ListRepositories(orgSlug string) ([]Repo, error) {
	path := fmt.Sprintf("/orgs/%s/repos", orgSlug)
	respBody, err := c.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var response struct {
		Repositories  []Repo  `json:"repositories"`
		NextPageToken *string `json:"next_page_token"`
	}
	if err := json.Unmarshal(respBody, &response); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode repo list JSON from GET %s: %w. Response start: %s", path, err, snippet)
	}
	// TODO: Handle pagination using response.NextPageToken
	return response.Repositories, nil
}

type RunCIBody struct {
	Revision         string `json:"revision,omitempty"`
	WorkflowRevision string `json:"workflow_revision,omitempty"`
	WorkflowSlug     string `json:"workflow_slug"`
	// 'input' (WorkflowInput) пока опускаем для простоты
}

// RunCIWorkflowResponse (из Swagger)
type RunCIWorkflowResponse struct {
	FluxID        *string `json:"flux_id"`
	TriggerStatus *string `json:"trigger_status"` // "already_exists", "created", "nothing_to_start"
}

// CreateRepository ('src repo create <name>') uses POST /orgs/{org_slug}/repos
func (c *Client) CreateRepository(orgSlug, name, slug, description, visibility string) (*Repo, error) {
	path := fmt.Sprintf("/orgs/%s/repos", orgSlug)
	reqBody := createRepoRequest{Name: name, Slug: slug, Description: description, Visibility: visibility}
	respBody, err := c.makeRequest(http.MethodPost, path, reqBody)
	if err != nil {
		return nil, err
	}
	var newRepo Repo
	if err := json.Unmarshal(respBody, &newRepo); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode created repo JSON from POST %s: %w. Response start: %s", path, err, snippet)
	}
	return &newRepo, nil
}

// GetRepository uses GET /repos/{org_slug}/{repo_slug}
// Includes the 'Parent' field.
func (c *Client) GetRepository(orgSlug, repoSlug string) (*Repo, error) {
	path := fmt.Sprintf("/repos/%s/%s", orgSlug, repoSlug)
	respBody, err := c.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil, fmt.Errorf("repository '%s/%s' not found or you don't have permission", orgSlug, repoSlug)
		}
		return nil, err
	}
	// Убираем отладочный вывод
	// fmt.Println("--- DEBUG: Raw API Response Body for GetRepository ---")
	// fmt.Println(string(respBody))
	// fmt.Println("--- END DEBUG ---")
	var repo Repo
	if err := json.Unmarshal(respBody, &repo); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode repository JSON from GET %s: %w. Response start: %s", path, err, snippet)
	}
	return &repo, nil
}

// ForkRepository uses POST /repos/{source_org_slug}/{source_repo_slug}/fork
type ForkRepositoryBody struct {
	OrgSlug           string `json:"org_slug"`
	Slug              string `json:"slug,omitempty"`
	DefaultBranchOnly bool   `json:"default_branch_only,omitempty"`
}

func (c *Client) ForkRepository(sourceOrgSlug, sourceRepoSlug, targetOrgSlug, newRepoSlug string, defaultBranchOnly bool) (*Repo, error) {
	path := fmt.Sprintf("/repos/%s/%s/fork", sourceOrgSlug, sourceRepoSlug)
	reqBody := ForkRepositoryBody{OrgSlug: targetOrgSlug, Slug: newRepoSlug, DefaultBranchOnly: defaultBranchOnly}
	respBody, err := c.makeRequest(http.MethodPost, path, reqBody)
	if err != nil {
		return nil, err
	}
	var forkedRepo Repo
	if err := json.Unmarshal(respBody, &forkedRepo); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode forked repo JSON from POST %s: %w. Response start: %s", path, err, snippet)
	}
	return &forkedRepo, nil
}

// ListPullRequests fetches pull requests for a specific repository.
// Uses GET /repos/{org_slug}/{repo_slug}/pulls
func (c *Client) ListPullRequests(orgSlug, repoSlug string) ([]PullRequest, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls", orgSlug, repoSlug)
	// TODO: Add query parameters later for filtering by state (e.g., ?state=open)
	respBody, err := c.makeRequest(http.MethodGet, path, nil) // Method GET
	if err != nil {
		return nil, err
	}

	// Response according to Swagger: ListRepositoryPullRequestsResponse
	var response struct {
		PullRequests  []PullRequest `json:"pull_requests"`
		NextPageToken *string       `json:"next_page_token"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode PR list JSON from GET %s: %w. Response start: %s", path, err, snippet)
	}

	// TODO: Handle pagination using response.NextPageToken

	return response.PullRequests, nil
}

func (c *Client) CreatePullRequest(orgSlug, repoSlug string, body CreatePullRequestBody) (*PullRequest, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls", orgSlug, repoSlug)
	respBody, err := c.makeRequest(http.MethodPost, path, body) // Метод POST
	if err != nil {
		// Обрабатываем возможные ошибки (422 - неверные ветки, 409 - уже существует)
		return nil, err
	}

	// Ответ по Swagger - созданный PullRequest
	var newPR PullRequest
	if err := json.Unmarshal(respBody, &newPR); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode created PR JSON from POST %s: %w. Response start: %s", path, err, snippet)
	}
	return &newPR, nil
}

func (c *Client) GetPullRequest(orgSlug, repoSlug, prSlug string) (*PullRequest, error) {
	//
	// *** ПУТЬ: /repos/{org_slug}/{repo_slug}/pulls/{pull_request_slug} ***
	//
	path := fmt.Sprintf("/repos/%s/%s/pulls/%s", orgSlug, repoSlug, prSlug)
	respBody, err := c.makeRequest(http.MethodGet, path, nil) // Метод GET
	if err != nil {
		// Обрабатываем 404 Not Found
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil, fmt.Errorf("pull request '%s/%s#%s' not found or you don't have permission", orgSlug, repoSlug, prSlug)
		}
		return nil, err
	}

	// Ответ по Swagger - PullRequest объект
	var pr PullRequest
	if err := json.Unmarshal(respBody, &pr); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode PR JSON from GET %s: %w. Response start: %s", path, err, snippet)
	}
	return &pr, nil
}

func (c *Client) MergePullRequest(orgSlug, repoSlug, prSlug string, mergeParams MergeParameters) (*SetDecisionResponse, error) {
	//
	// *** ПУТЬ ИЗМЕНЕН: .../merge -> .../decision ***
	//
	path := fmt.Sprintf("/repos/%s/%s/pulls/%s/decision", orgSlug, repoSlug, prSlug)

	//
	// *** ТЕЛО ЗАПРОСА ИЗМЕНЕНО: MergeDecisionBody -> SetDecisionBody ***
	//
	// Мы жестко задаем "approve", так как это команда "merge".
	// mergeParams (squash, rebase, delete_branch) ИГНОРИРУЮТСЯ.
	//
	reqBody := SetDecisionBody{
		ReviewDecision: "approve",
	}

	// Вызываем makeRequest с POST
	respBody, err := c.makeRequest(http.MethodPost, path, reqBody)
	if err != nil {
		return nil, err
	}

	//
	// *** ОТВЕТ ИЗМЕНЕН: PullRequest -> SetDecisionResponse ***
	//
	var decisionResponse SetDecisionResponse
	if err := json.Unmarshal(respBody, &decisionResponse); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode merge/decision response JSON: %w. Response start: %s", err, snippet)
	}
	return &decisionResponse, nil
}

// ListRepositoryIssues ('src issue list')
// (GET /repos/{org_slug}/{repo_slug}/issues)
func (c *Client) ListRepositoryIssues(orgSlug, repoSlug string) ([]Issue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues", orgSlug, repoSlug)
	// TODO: Добавить query-параметры для фильтрации (state=open, etc.)
	respBody, err := c.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	// Ответ по Swagger: ListRepositoryIssuesResponse
	var response struct {
		Issues        []Issue `json:"issues"`
		NextPageToken *string `json:"next_page_token"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode issue list JSON from GET %s: %w. Response start: %s", path, err, snippet)
	}
	// TODO: Handle pagination
	return response.Issues, nil
}

// CreateIssue ('src issue create')
// (POST /repos/{org_slug}/{repo_slug}/issues)
func (c *Client) CreateIssue(orgSlug, repoSlug string, body CreateIssueBody) (*Issue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues", orgSlug, repoSlug)
	respBody, err := c.makeRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	var newIssue Issue
	if err := json.Unmarshal(respBody, &newIssue); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode created issue JSON from POST %s: %w. Response start: %s", path, err, snippet)
	}
	return &newIssue, nil
}

// GetIssue ('src issue view <id>')
// (GET /repos/{org_slug}/{repo_slug}/issues/{issue_slug})
func (c *Client) GetIssue(orgSlug, repoSlug, issueSlug string) (*Issue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%s", orgSlug, repoSlug, issueSlug)
	respBody, err := c.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil, fmt.Errorf("issue '%s/%s#%s' not found or you don't have permission", orgSlug, repoSlug, issueSlug)
		}
		return nil, err
	}
	var issue Issue
	if err := json.Unmarshal(respBody, &issue); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode issue JSON from GET %s: %w. Response start: %s", path, err, snippet)
	}
	return &issue, nil
}

// UpdateIssue ('src issue update <id>' и 'src issue close <id>')
// (PATCH /repos/{org_slug}/{repo_slug}/issues/{issue_slug})
func (c *Client) UpdateIssue(orgSlug, repoSlug, issueSlug string, body UpdateIssueBody) (*Issue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%s", orgSlug, repoSlug, issueSlug)
	respBody, err := c.makeRequest(http.MethodPatch, path, body)
	if err != nil {
		return nil, err
	}
	var updatedIssue Issue
	if err := json.Unmarshal(respBody, &updatedIssue); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode updated issue JSON from PATCH %s: %w. Response start: %s", path, err, snippet)
	}
	return &updatedIssue, nil
}

func (c *Client) ListMilestonesForRepository(orgSlug, repoSlug string) ([]Milestone, error) {
	path := fmt.Sprintf("/repos/%s/%s/milestones", orgSlug, repoSlug)
	respBody, err := c.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	// Ответ по Swagger: ListMilestonesForRepositoryResponse
	var response struct {
		Items         []Milestone `json:"items"`
		NextPageToken *string     `json:"next_page_token"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode milestone list JSON from GET %s: %w. Response start: %s", path, err, snippet)
	}
	// TODO: Handle pagination
	return response.Items, nil
}

// CreateMilestone ('src milestone create')
// (POST /repos/{org_slug}/{repo_slug}/milestones)
func (c *Client) CreateMilestone(orgSlug, repoSlug string, body CreateMilestoneBody) (*Milestone, error) {
	path := fmt.Sprintf("/repos/%s/%s/milestones", orgSlug, repoSlug)
	respBody, err := c.makeRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	var newMilestone Milestone
	if err := json.Unmarshal(respBody, &newMilestone); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode created milestone JSON from POST %s: %w. Response start: %s", path, err, snippet)
	}
	return &newMilestone, nil
}

// GetMilestone ('src milestone view <id>')
// (GET /repos/{org_slug}/{repo_slug}/milestones/{milestone_slug})
func (c *Client) GetMilestone(orgSlug, repoSlug, milestoneSlug string) (*Milestone, error) {
	path := fmt.Sprintf("/repos/%s/%s/milestones/%s", orgSlug, repoSlug, milestoneSlug)
	respBody, err := c.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil, fmt.Errorf("milestone '%s' in '%s/%s' not found or you don't have permission", milestoneSlug, orgSlug, repoSlug)
		}
		return nil, err
	}
	var milestone Milestone
	if err := json.Unmarshal(respBody, &milestone); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode milestone JSON from GET %s: %w. Response start: %s", path, err, snippet)
	}
	return &milestone, nil
}

func (c *Client) RunWorkflow(orgSlug, repoSlug, workflowName string, body RunCIBody) (*RunCIWorkflowResponse, error) {
	path := fmt.Sprintf("/%s/%s/cicd/runs", orgSlug, repoSlug)
	respBody, err := c.makeRequest(http.MethodPost, path, body)
	if err != nil {
		// API может вернуть 404, если workflow 'workflowName' не найден
		return nil, err
	}

	var runResponse RunCIWorkflowResponse
	if err := json.Unmarshal(respBody, &runResponse); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode workflow run response JSON from POST %s: %w. Response start: %s", path, err, snippet)
	}
	return &runResponse, nil
}

func (c *Client) ListRepoRoles(orgSlug, repoSlug string) ([]SubjectRole, error) {
	path := fmt.Sprintf("/repos/%s/%s/roles", orgSlug, repoSlug)
	// TODO: Add pagination parameters later if needed
	respBody, err := c.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	// Адаптируем под ожидаемый (но неявно заданный) ответ
	var response ListRepoRolesResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode repo roles list JSON from GET %s: %w. Response start: %s", path, err, snippet)
	}
	// TODO: Handle pagination via response.NextPageToken
	return response.SubjectRoles, nil
}

// AddRepoRole ('src access role add <repo> <user_id> <role>')
// (POST /repos/{org_slug}/{repo_slug}/roles)
// Примечание: API принимает массив, но для CLI удобнее добавлять по одному.
func (c *Client) AddRepoRole(orgSlug, repoSlug string, userID string, role RepoRole) error {
	path := fmt.Sprintf("/repos/%s/%s/roles", orgSlug, repoSlug)
	body := AddRepoRolesBody{
		SubjectRoles: []SubjectRole{
			{
				Role: role,
				Subject: Subject{
					Type: SubjectTypeUser, // Пока поддерживаем только пользователей
					ID:   userID,
				},
			},
		},
	}
	_, err := c.makeRequest(http.MethodPost, path, body)
	// Swagger говорит 200 OK, но тело ответа не определено, поэтому игнорируем его.
	if err != nil {
		// API может вернуть 400 Bad Request при неверном role/user_id, 403 Forbidden, 404 Not Found
		return err
	}
	return nil
}

// RemoveRepoRole ('src access role remove <repo> <user_id> <role>')
// (POST /repos/{org_slug}/{repo_slug}/roles/remove)
func (c *Client) RemoveRepoRole(orgSlug, repoSlug string, userID string, role RepoRole) error {
	path := fmt.Sprintf("/repos/%s/%s/roles/remove", orgSlug, repoSlug)
	body := RemoveRepoRolesBody{
		SubjectRoles: []SubjectRole{
			{
				Role: role,
				Subject: Subject{
					Type: SubjectTypeUser,
					ID:   userID,
				},
			},
		},
	}
	_, err := c.makeRequest(http.MethodPost, path, body)
	// Swagger говорит 200 OK, тело ответа не определено.
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ListRuns(orgSlug, repoSlug string) ([]RunStatus, error) {
	path := fmt.Sprintf("/%s/%s/cicd/runs", orgSlug, repoSlug)
	respBody, err := c.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var response ListRunsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		snippet := string(respBody)
		if len(snippet) > 150 {
			snippet = snippet[:150] + "..."
		}
		return nil, fmt.Errorf("failed to decode run list JSON from GET %s: %w. Response start: %s", path, err, snippet)
	}
	return response.Runs, nil
}

// GetRunStatus ('src workflow status <run_slug>') - GET /{org_slug}/{repo_slug}/cicd/runs/{run_slug}
func (c *Client) GetRunStatus(orgSlug, repoSlug, runSlug string) (*RunStatus, error) {
	path := fmt.Sprintf("/%s/%s/cicd/runs/%s", orgSlug, repoSlug, runSlug)
	respBody, err := c.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil, fmt.Errorf("run '%s' in '%s/%s' not found or you don't have permission", runSlug, orgSlug, repoSlug)
		}
		return nil, err
	}
	var status RunStatus
	if err := json.Unmarshal(respBody, &status); err != nil {
		return nil, fmt.Errorf("failed to decode run status JSON from GET %s: %w", path, err)
	}
	return &status, nil
}

// GetLogs - GET /{org_slug}/{repo_slug}/cicd/logs/{run_slug}/{workflow_slug}/{task_slug}/{cube_slug}
func (c *Client) GetLogs(orgSlug, repoSlug, runSlug, workflowSlug, taskSlug, cubeSlug string) (string, error) {
	path := fmt.Sprintf("/%s/%s/cicd/logs/%s/%s/%s/%s", orgSlug, repoSlug, runSlug, workflowSlug, taskSlug, cubeSlug)

	// Используем makeRequest, но ожидаем, что ответ может быть text/plain (логи)
	// makeRequest все еще проверяет на application/json, поэтому нам нужно его изменить.

	// Временное решение:
	// Так как makeRequest жестко проверяет Content-Type на application/json,
	// мы должны либо изменить makeRequest, либо написать отдельный метод.
	// Оставим пока так, предполагая, что API может возвращать JSON-обертку вокруг логов.
	// Если API возвращает чистый текст, makeRequest выдаст ошибку, которую нужно будет отловить.

	respBody, err := c.makeRequest(http.MethodGet, path, nil)
	if err != nil {
		return "", err
	}

	// Предполагаем, что API возвращает { "logs": "..." }
	var logResponse struct {
		Logs *string `json:"logs"`
	}
	if err := json.Unmarshal(respBody, &logResponse); err != nil {
		// Если не JSON, возвращаем сырой ответ, если это текст
		return string(respBody), nil
	}
	return cliutils.DerefString(logResponse.Logs), nil
}

// GetArtifacts - GET /{org_slug}/{repo_slug}/cicd/artifacts/{run_slug}/{workflow_slug}/{task_slug}/{cube_slug}
func (c *Client) GetArtifacts(orgSlug, repoSlug, runSlug, workflowSlug, taskSlug, cubeSlug string) ([]byte, error) {
	path := fmt.Sprintf("/%s/%s/cicd/artifacts/%s/%s/%s/%s", orgSlug, repoSlug, runSlug, workflowSlug, taskSlug, cubeSlug)

	// Для артефактов обычно требуется отдельный HTTP-клиент, который
	// не накладывает жестких ограничений на Content-Type (например, application/octet-stream).

	// Временно используем makeRequest, но без проверки Content-Type
	req, err := http.NewRequest(http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("request creation error: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.Token)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request execution error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body) // Пытаемся прочитать ошибку
		return nil, fmt.Errorf("API returned error: %s (path: %s). Body: %s", resp.Status, path, string(respBody))
	}

	// Читаем сырые данные
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read artifact data: %w", err)
	}
	return data, nil
}
