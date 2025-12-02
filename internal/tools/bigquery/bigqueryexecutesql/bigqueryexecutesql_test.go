// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bigqueryexecutesql_test

import (
	"context"
	"net/http"
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/genai-toolbox/internal/server"
	"github.com/googleapis/genai-toolbox/internal/sources"
	bigqueryds "github.com/googleapis/genai-toolbox/internal/sources/bigquery"
	"github.com/googleapis/genai-toolbox/internal/testutils"
	"github.com/googleapis/genai-toolbox/internal/tools"
	"github.com/googleapis/genai-toolbox/internal/tools/bigquery/bigqueryexecutesql"
	"github.com/googleapis/genai-toolbox/internal/util/parameters"
	"google.golang.org/api/option"
	"google.golang.org/api/transport/http/testing as http_testing"

	bq "google.golang.org/api/bigquery/v2"
)

func TestParseFromYamlBigQueryExecuteSql(t *testing.T) {
	ctx, err := testutils.ContextWithNewLogger()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	tcs := []struct {
		desc string
		in   string
		want server.ToolConfigs
	}{
		{
			desc: "basic example",
			in: `
			tools:
				example_tool:
					kind: bigquery-execute-sql
					source: my-instance
					description: some description
			`,
			want: server.ToolConfigs{
				"example_tool": &bigqueryexecutesql.Config{
					Name:         "example_tool",
					Kind:         "bigquery-execute-sql",
					Source:       "my-instance",
					Description:  "some description",
					AuthRequired: []string{},
				},
			},
		},
		{
			desc: "with maxQueryResultRows",
			in: `
			tools:
				example_tool:
					kind: bigquery-execute-sql
					source: my-instance
					description: some description
					maxQueryResultRows: 10
			`,
			want: server.ToolConfigs{
				"example_tool": &bigqueryexecutesql.Config{
					Name:               "example_tool",
					Kind:               "bigquery-execute-sql",
					Source:             "my-instance",
					Description:        "some description",
					AuthRequired:       []string{},
					MaxQueryResultRows: 10,
				},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			got := struct {
				Tools server.ToolConfigs `yaml:"tools"`
			}{}
			// Parse contents
			err := yaml.UnmarshalContext(ctx, testutils.FormatYaml(tc.in), &got)
			if err != nil {
				t.Fatalf("unable to unmarshal: %s", err)
			}
			if diff := cmp.Diff(tc.want, got.Tools, cmp.AllowUnexported(bigqueryexecutesql.Config{})); diff != "" {
				t.Fatalf("incorrect parse: diff %v", diff)
			}
		})
	}
}

type mockBigQuerySource struct {
	sources.Source
	client      *bigquery.Client
	restService *bq.Service
}

func (s *mockBigQuerySource) BigQueryClient() *bigquery.Client                      { return s.client }
func (s *mockBigQuerySource) BigQuerySession() bigqueryds.BigQuerySessionProvider   { return nil }
func (s *mockBigQuerySource) BigQueryWriteMode() string                             { return bigqueryds.WriteModeAllowed }
func (s *mockBigQuerySource) BigQueryRestService() *bq.Service                      { return s.restService }
func (s *mockBigQuerySource) BigQueryClientCreator() bigqueryds.BigqueryClientCreator { return nil }
func (s *mockBigQuerySource) UseClientAuthorization() bool                          { return false }
func (s *mockBigQuerySource) IsDatasetAllowed(projectID, datasetID string) bool     { return true }
func (s *mockBigQuerySource) BigQueryAllowedDatasets() []string                     { return nil }

func TestInvoke(t *testing.T) {
	ctx := context.Background()

	// Create a mock HTTP transport.
	mockTransport, err := http_testing.NewRoundTripper(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       http.NoBody,
		}, nil
	})

	if err != nil {
		t.Fatalf("http_testing.NewRoundTripper: %v", err)
	}

	client, err := bigquery.NewClient(ctx, "test-project", option.WithHTTPClient(&http.Client{Transport: mockTransport}))
	if err != nil {
		t.Fatalf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	restService, err := bq.NewService(ctx, option.WithHTTPClient(&http.Client{Transport: mockTransport}))
	if err != nil {
		t.Fatalf("bq.NewService: %v", err)
	}

	srcs := map[string]sources.Source{
		"my-instance": &mockBigQuerySource{client: client, restService: restService},
	}

	tcs := []struct {
		name            string
		config          tools.ToolConfig
		params          parameters.ParamValues
		expectedNumRows int
		expectedErr     bool
	}{
		{
			name: "no row limit",
			config: &bigqueryexecutesql.Config{
				Name:        "test-tool",
				Source:      "my-instance",
				Description: "test tool",
			},
			params:          parameters.ParamValues{parameters.ParamValue{Name: "sql", Value: "SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3"}, parameters.ParamValue{Name: "dry_run", Value: true}},
			expectedNumRows: 0,
			expectedErr:     false,
		},
		{
			name: "with row limit",
			config: &bigqueryexecutesql.Config{
				Name:               "test-tool",
				Source:             "my-instance",
				Description:        "test tool",
				MaxQueryResultRows: 2,
			},
			params:          parameters.ParamValues{parameters.ParamValue{Name: "sql", Value: "SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3"}, parameters.ParamValue{Name: "dry_run", Value: false}},
			expectedNumRows: 2,
			expectedErr:     false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tool, err := tc.config.Initialize(srcs)
			if err != nil {
				t.Fatalf("Initialize() error = %v", err)
			}

			result, err := tool.Invoke(ctx, tc.params, "")
			if (err != nil) != tc.expectedErr {
				t.Fatalf("Invoke() error = %v, wantErr %v", err, tc.expectedErr)
			}

			if !tc.expectedErr {
				rows, ok := result.([]any)
				if !ok {
					t.Fatalf("Invoke() returned non-rows result: %T", result)
				}
				if len(rows) != tc.expectedNumRows {
					t.Errorf("Invoke() returned %d rows, want %d", len(rows), tc.expectedNumRows)
				}
			}
		})
	}
}
