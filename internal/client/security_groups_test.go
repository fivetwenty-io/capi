package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	internalhttp "github.com/fivetwenty-io/capi/v3/internal/http"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityGroupsClient_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/security_groups", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.SecurityGroupCreateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Equal(t, "my-security-group", request.Name)
		assert.NotNil(t, request.GloballyEnabled)
		assert.True(t, request.GloballyEnabled.Running)
		assert.False(t, request.GloballyEnabled.Staging)
		assert.Len(t, request.Rules, 2)

		now := time.Now()
		sg := capi.SecurityGroup{
			Resource: capi.Resource{
				GUID:      "sg-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name:            request.Name,
			GloballyEnabled: *request.GloballyEnabled,
			Rules:           request.Rules,
			Relationships: capi.SecurityGroupRelationships{
				RunningSpaces: capi.ToManyRelationship{
					Data: []capi.RelationshipData{},
				},
				StagingSpaces: capi.ToManyRelationship{
					Data: []capi.RelationshipData{},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(sg)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	securityGroups := NewSecurityGroupsClient(client.httpClient)

	port80 := "80"
	typeICMP := 8
	codeICMP := 0
	descriptionICMP := "Allow ping"

	request := &capi.SecurityGroupCreateRequest{
		Name: "my-security-group",
		GloballyEnabled: &capi.SecurityGroupGloballyEnabled{
			Running: true,
			Staging: false,
		},
		Rules: []capi.SecurityGroupRule{
			{
				Protocol:    "tcp",
				Destination: "10.0.0.0/24",
				Ports:       &port80,
			},
			{
				Protocol:    "icmp",
				Destination: "10.0.0.0/24",
				Type:        &typeICMP,
				Code:        &codeICMP,
				Description: &descriptionICMP,
			},
		},
	}

	sg, err := securityGroups.Create(context.Background(), request)
	require.NoError(t, err)
	assert.NotNil(t, sg)
	assert.Equal(t, "sg-guid", sg.GUID)
	assert.Equal(t, "my-security-group", sg.Name)
	assert.True(t, sg.GloballyEnabled.Running)
	assert.Len(t, sg.Rules, 2)
}

func TestSecurityGroupsClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/security_groups/sg-guid", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		now := time.Now()
		ports := "443,80,8080"
		sg := capi.SecurityGroup{
			Resource: capi.Resource{
				GUID:      "sg-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name: "my-security-group",
			GloballyEnabled: capi.SecurityGroupGloballyEnabled{
				Running: true,
				Staging: false,
			},
			Rules: []capi.SecurityGroupRule{
				{
					Protocol:    "tcp",
					Destination: "10.10.10.0/24",
					Ports:       &ports,
				},
			},
			Relationships: capi.SecurityGroupRelationships{
				RunningSpaces: capi.ToManyRelationship{
					Data: []capi.RelationshipData{
						{GUID: "space-guid-1"},
					},
				},
				StagingSpaces: capi.ToManyRelationship{
					Data: []capi.RelationshipData{},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sg)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	securityGroups := NewSecurityGroupsClient(client.httpClient)

	sg, err := securityGroups.Get(context.Background(), "sg-guid")
	require.NoError(t, err)
	assert.NotNil(t, sg)
	assert.Equal(t, "sg-guid", sg.GUID)
	assert.Equal(t, "my-security-group", sg.Name)
	assert.True(t, sg.GloballyEnabled.Running)
	assert.False(t, sg.GloballyEnabled.Staging)
	assert.Len(t, sg.Rules, 1)
	assert.Equal(t, "tcp", sg.Rules[0].Protocol)
}

func TestSecurityGroupsClient_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/security_groups", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "sg1,sg2", r.URL.Query().Get("names"))
		assert.Equal(t, "true", r.URL.Query().Get("globally_enabled_running"))

		now := time.Now()
		response := capi.ListResponse[capi.SecurityGroup]{
			Pagination: capi.Pagination{
				TotalResults: 2,
				TotalPages:   1,
				First:        capi.Link{Href: "/v3/security_groups?page=1"},
				Last:         capi.Link{Href: "/v3/security_groups?page=1"},
			},
			Resources: []capi.SecurityGroup{
				{
					Resource: capi.Resource{
						GUID:      "sg-guid-1",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name: "sg1",
					GloballyEnabled: capi.SecurityGroupGloballyEnabled{
						Running: true,
						Staging: false,
					},
					Rules: []capi.SecurityGroupRule{},
				},
				{
					Resource: capi.Resource{
						GUID:      "sg-guid-2",
						CreatedAt: now,
						UpdatedAt: now,
					},
					Name: "sg2",
					GloballyEnabled: capi.SecurityGroupGloballyEnabled{
						Running: true,
						Staging: true,
					},
					Rules: []capi.SecurityGroupRule{},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	securityGroups := NewSecurityGroupsClient(client.httpClient)

	params := &capi.QueryParams{
		Filters: map[string][]string{
			"names":                    {"sg1", "sg2"},
			"globally_enabled_running": {"true"},
		},
	}

	list, err := securityGroups.List(context.Background(), params)
	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Equal(t, 2, list.Pagination.TotalResults)
	assert.Len(t, list.Resources, 2)
	assert.Equal(t, "sg1", list.Resources[0].Name)
	assert.Equal(t, "sg2", list.Resources[1].Name)
}

func TestSecurityGroupsClient_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/security_groups/sg-guid", r.URL.Path)
		assert.Equal(t, "PATCH", r.Method)

		var request capi.SecurityGroupUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.NotNil(t, request.Name)
		assert.Equal(t, "updated-sg", *request.Name)
		assert.NotNil(t, request.GloballyEnabled)
		assert.False(t, request.GloballyEnabled.Running)
		assert.True(t, request.GloballyEnabled.Staging)

		now := time.Now()
		sg := capi.SecurityGroup{
			Resource: capi.Resource{
				GUID:      "sg-guid",
				CreatedAt: now,
				UpdatedAt: now,
			},
			Name:            *request.Name,
			GloballyEnabled: *request.GloballyEnabled,
			Rules:           request.Rules,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sg)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	securityGroups := NewSecurityGroupsClient(client.httpClient)

	name := "updated-sg"
	ports := "443"
	request := &capi.SecurityGroupUpdateRequest{
		Name: &name,
		GloballyEnabled: &capi.SecurityGroupGloballyEnabled{
			Running: false,
			Staging: true,
		},
		Rules: []capi.SecurityGroupRule{
			{
				Protocol:    "tcp",
				Destination: "192.168.0.0/16",
				Ports:       &ports,
			},
		},
	}

	sg, err := securityGroups.Update(context.Background(), "sg-guid", request)
	require.NoError(t, err)
	assert.NotNil(t, sg)
	assert.Equal(t, "sg-guid", sg.GUID)
	assert.Equal(t, "updated-sg", sg.Name)
	assert.False(t, sg.GloballyEnabled.Running)
	assert.True(t, sg.GloballyEnabled.Staging)
}

func TestSecurityGroupsClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/security_groups/sg-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		job := capi.Job{
			Resource: capi.Resource{
				GUID: "job-guid",
			},
			Operation: "security_groups.delete",
			State:     "PROCESSING",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(job)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	securityGroups := NewSecurityGroupsClient(client.httpClient)

	job, err := securityGroups.Delete(context.Background(), "sg-guid")
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "job-guid", job.GUID)
	assert.Equal(t, "security_groups.delete", job.Operation)
}

func TestSecurityGroupsClient_BindRunningSpaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/security_groups/sg-guid/relationships/running_spaces", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.SecurityGroupBindRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Len(t, request.Data, 2)
		assert.Equal(t, "space-guid-1", request.Data[0].GUID)
		assert.Equal(t, "space-guid-2", request.Data[1].GUID)

		response := capi.ToManyRelationship(request)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	securityGroups := NewSecurityGroupsClient(client.httpClient)

	relationship, err := securityGroups.BindRunningSpaces(context.Background(), "sg-guid", []string{"space-guid-1", "space-guid-2"})
	require.NoError(t, err)
	assert.NotNil(t, relationship)
	assert.Len(t, relationship.Data, 2)
}

func TestSecurityGroupsClient_UnbindRunningSpace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/security_groups/sg-guid/relationships/running_spaces/space-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	securityGroups := NewSecurityGroupsClient(client.httpClient)

	err := securityGroups.UnbindRunningSpace(context.Background(), "sg-guid", "space-guid")
	require.NoError(t, err)
}

func TestSecurityGroupsClient_BindStagingSpaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/security_groups/sg-guid/relationships/staging_spaces", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var request capi.SecurityGroupBindRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)

		assert.Len(t, request.Data, 1)
		assert.Equal(t, "space-guid-1", request.Data[0].GUID)

		response := capi.ToManyRelationship(request)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	securityGroups := NewSecurityGroupsClient(client.httpClient)

	relationship, err := securityGroups.BindStagingSpaces(context.Background(), "sg-guid", []string{"space-guid-1"})
	require.NoError(t, err)
	assert.NotNil(t, relationship)
	assert.Len(t, relationship.Data, 1)
	assert.Equal(t, "space-guid-1", relationship.Data[0].GUID)
}

func TestSecurityGroupsClient_UnbindStagingSpace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/security_groups/sg-guid/relationships/staging_spaces/space-guid", r.URL.Path)
		assert.Equal(t, "DELETE", r.Method)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &Client{httpClient: internalhttp.NewClient(server.URL, nil)}
	securityGroups := NewSecurityGroupsClient(client.httpClient)

	err := securityGroups.UnbindStagingSpace(context.Background(), "sg-guid", "space-guid")
	require.NoError(t, err)
}
