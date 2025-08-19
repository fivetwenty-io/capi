package capi

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryParams_ToValues(t *testing.T) {
	tests := []struct {
		name     string
		params   *QueryParams
		expected url.Values
	}{
		{
			name:     "empty params",
			params:   NewQueryParams(),
			expected: url.Values{},
		},
		{
			name: "with pagination",
			params: &QueryParams{
				Page:    2,
				PerPage: 50,
			},
			expected: url.Values{
				"page":     []string{"2"},
				"per_page": []string{"50"},
			},
		},
		{
			name: "with ordering",
			params: &QueryParams{
				OrderBy: "-created_at",
			},
			expected: url.Values{
				"order_by": []string{"-created_at"},
			},
		},
		{
			name: "with label selector",
			params: &QueryParams{
				LabelSelector: "environment=production,team=platform",
			},
			expected: url.Values{
				"label_selector": []string{"environment=production,team=platform"},
			},
		},
		{
			name: "with includes",
			params: &QueryParams{
				Include: []string{"space", "organization"},
			},
			expected: url.Values{
				"include": []string{"space,organization"},
			},
		},
		{
			name: "with fields",
			params: &QueryParams{
				Fields: map[string][]string{
					"apps":   {"name", "state"},
					"spaces": {"name"},
				},
			},
			expected: url.Values{
				"fields[apps]":   []string{"name,state"},
				"fields[spaces]": []string{"name"},
			},
		},
		{
			name: "with filters",
			params: &QueryParams{
				Filters: map[string][]string{
					"names":  {"app1", "app2"},
					"states": {"STARTED"},
				},
			},
			expected: url.Values{
				"names":  []string{"app1,app2"},
				"states": []string{"STARTED"},
			},
		},
		{
			name: "with all options",
			params: &QueryParams{
				Page:          3,
				PerPage:       25,
				OrderBy:       "name",
				LabelSelector: "env=prod",
				Include:       []string{"space"},
				Fields: map[string][]string{
					"apps": {"guid", "name"},
				},
				Filters: map[string][]string{
					"states": {"STARTED", "STOPPED"},
				},
			},
			expected: url.Values{
				"page":           []string{"3"},
				"per_page":       []string{"25"},
				"order_by":       []string{"name"},
				"label_selector": []string{"env=prod"},
				"include":        []string{"space"},
				"fields[apps]":   []string{"guid,name"},
				"states":         []string{"STARTED,STOPPED"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.ToValues()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryParams_Builders(t *testing.T) {
	t.Run("chaining methods", func(t *testing.T) {
		params := NewQueryParams().
			WithPage(2).
			WithPerPage(100).
			WithOrderBy("-updated_at").
			WithLabelSelector("team=backend").
			WithInclude("space", "organization").
			WithFields("apps", "guid", "name", "state").
			WithFilter("states", "STARTED").
			WithFilter("names", "app1", "app2")

		values := params.ToValues()

		assert.Equal(t, "2", values.Get("page"))
		assert.Equal(t, "100", values.Get("per_page"))
		assert.Equal(t, "-updated_at", values.Get("order_by"))
		assert.Equal(t, "team=backend", values.Get("label_selector"))
		assert.Equal(t, "space,organization", values.Get("include"))
		assert.Equal(t, "guid,name,state", values.Get("fields[apps]"))
		assert.Equal(t, "STARTED", values.Get("states"))
		assert.Equal(t, "app1,app2", values.Get("names"))
	})

	t.Run("WithInclude appends", func(t *testing.T) {
		params := NewQueryParams().
			WithInclude("space").
			WithInclude("organization", "domain")

		assert.Equal(t, []string{"space", "organization", "domain"}, params.Include)
	})

	t.Run("WithFilter appends", func(t *testing.T) {
		params := NewQueryParams().
			WithFilter("names", "app1").
			WithFilter("names", "app2", "app3")

		assert.Equal(t, []string{"app1", "app2", "app3"}, params.Filters["names"])
	})

	t.Run("WithFields replaces", func(t *testing.T) {
		params := NewQueryParams().
			WithFields("apps", "guid").
			WithFields("apps", "name", "state")

		assert.Equal(t, []string{"name", "state"}, params.Fields["apps"])
	})
}

func TestNewQueryParams(t *testing.T) {
	params := NewQueryParams()

	assert.NotNil(t, params)
	assert.NotNil(t, params.Fields)
	assert.NotNil(t, params.Filters)
	assert.Equal(t, 0, params.Page)
	assert.Equal(t, 0, params.PerPage)
	assert.Equal(t, "", params.OrderBy)
	assert.Equal(t, "", params.LabelSelector)
	assert.Nil(t, params.Include)
}
