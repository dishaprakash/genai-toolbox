// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package bigquerysql

import (
	"testing"

	"cloud.google.com/go/bigquery"
	"github.com/googleapis/genai-toolbox/internal/sources"
	bigqueryds "github.com/googleapis/genai-toolbox/internal/sources/bigquery"
	bigqueryrestapi "google.golang.org/api/bigquery/v2"
)

type fakeSource struct {
	sources.Source
	maxQueryResultRows int
}

func (s fakeSource) GetMaxQueryResultRows() int {
	return s.maxQueryResultRows
}

func (s fakeSource) BigQueryRestService() *bigqueryrestapi.Service {
	return nil
}

func (s fakeSource) BigQueryClientCreator() bigqueryds.BigqueryClientCreator {
	return nil
}

func (s fakeSource) UseClientAuthorization() bool {
	return false
}

func (s fakeSource) BigQueryClient() *bigquery.Client {
	return nil
}

func (s fakeSource) BigQuerySession() bigqueryds.BigQuerySessionProvider {
	return nil
}

func (s fakeSource) BigQueryWriteMode() string {
	return "blocked"
}

func (s fakeSource) SourceConfigKind() string {
	return "bigquery"
}

func TestInitialize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		maxQueryResultRows int
		config             Config
		sources            map[string]sources.Source
		wantErr            bool
	}{
		{
			name:               "Valid config with max query result rows",
			maxQueryResultRows: 100,
			config: Config{
				Name:   "test",
				Source: "bq",
			},
			sources: map[string]sources.Source{
				"bq": fakeSource{maxQueryResultRows: 100},
			},
			wantErr: false,
		},
		{
			name:               "Valid config with default max query result rows",
			maxQueryResultRows: 0,
			config: Config{
				Name:   "test",
				Source: "bq",
			},
			sources: map[string]sources.Source{
				"bq": fakeSource{maxQueryResultRows: 0},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, err := tt.config.Initialize(tt.sources)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Initialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if bqTool, ok := tool.(Tool); ok {
				if bqTool.maxQueryResultRows != tt.maxQueryResultRows {
					t.Errorf("maxQueryResultRows = %v, want %v", bqTool.maxQueryResultRows, tt.maxQueryResultRows)
				}
			} else {
				t.Errorf("tool is not a BigQuery SQL tool")
			}
		})
	}
}
