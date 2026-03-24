// Copyright 2026 Google LLC
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

package embeddingmodels

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
)

// EmbeddingModelConfigFactory defines the function signature for creating a EmbeddingModelConfig.
type EmbeddingModelConfigFactory func(ctx context.Context, name string, decoder *yaml.Decoder) (EmbeddingModelConfig, error)

var embeddingModelRegistry = make(map[string]EmbeddingModelConfigFactory)

// Register registers a new embeddingModel type with its factory.
// It returns false if the type is already registered.
func Register(embeddingModelType string, factory EmbeddingModelConfigFactory) bool {
	if _, exists := embeddingModelRegistry[embeddingModelType]; exists {
		// EmbeddingModel with this type already exists, do not overwrite.
		return false
	}
	embeddingModelRegistry[embeddingModelType] = factory
	return true
}

// DecodeConfig decodes a embeddingModel configuration using the registered factory for the given type.
func DecodeConfig(ctx context.Context, embeddingModelType string, name string, decoder *yaml.Decoder) (EmbeddingModelConfig, error) {
	factory, found := embeddingModelRegistry[embeddingModelType]
	if !found {
		return nil, fmt.Errorf("unknown embeddingModel type: %q", embeddingModelType)
	}
	embeddingModelConfig, err := factory(ctx, name, decoder)
	if err != nil {
		return nil, fmt.Errorf("unable to parse embeddingModel %q as %q: %w", name, embeddingModelType, err)
	}
	return embeddingModelConfig, err
}

// EmbeddingModelConfig is the interface for configuring embedding models.
type EmbeddingModelConfig interface {
	EmbeddingModelConfigType() string
	Initialize(context.Context) (EmbeddingModel, error)
}

type EmbeddingModel interface {
	EmbeddingModelType() string
	ToConfig() EmbeddingModelConfig
	EmbedParameters(context.Context, []string) ([][]float32, error)
}

type VectorFormatter func(vectorFloats []float32) any

// FormatVectorForPgvector converts a slice of floats into a PostgreSQL vector literal string: '[x, y, z]'
func FormatVectorForPgvector(vectorFloats []float32) any {
	if len(vectorFloats) == 0 {
		return "[]"
	}

	// Pre-allocate the builder.
	var b strings.Builder
	b.Grow(len(vectorFloats) * 10)

	b.WriteByte('[')
	for i, f := range vectorFloats {
		if i > 0 {
			b.WriteString(", ")
		}
		b.Write(strconv.AppendFloat(nil, float64(f), 'g', -1, 32))
	}
	b.WriteByte(']')

	return b.String()
}

var _ VectorFormatter = FormatVectorForPgvector
