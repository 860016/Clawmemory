package services

import (
	"math"
	"strings"
	"unicode"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type SearchService struct {
	db *gorm.DB
}

func NewSearchService(db *gorm.DB) *SearchService {
	return &SearchService{db: db}
}

func (s *SearchService) SemanticSearch(userID uint, query string, limit int) ([]map[string]interface{}, error) {
	var memories []models.Memory
	s.db.Where("user_id = ? AND status != ?", userID, "trashed").Find(&memories)

	if len(memories) == 0 {
		return []map[string]interface{}{}, nil
	}

	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		return []map[string]interface{}{}, nil
	}

	docs := make([][]string, len(memories))
	for i, m := range memories {
		docs[i] = tokenize(m.Key + " " + m.Value)
	}

	idf := computeIDF(docs, queryTokens)

	type scoredMemory struct {
		Memory models.Memory
		Score  float64
	}
	scored := make([]scoredMemory, 0, len(memories))

	for i, m := range memories {
		tf := computeTF(docs[i], queryTokens)
		score := 0.0
		for _, token := range queryTokens {
			score += tf[token] * idf[token]
		}
		score += float64(m.Importance) * 0.1
		score += float64(min(m.AccessCount, 100)) * 0.01

		if score > 0 {
			scored = append(scored, scoredMemory{Memory: m, Score: score})
		}
	}

	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].Score > scored[i].Score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	if limit > len(scored) {
		limit = len(scored)
	}

	result := make([]map[string]interface{}, 0, limit)
	for i := 0; i < limit; i++ {
		m := scored[i].Memory
		result = append(result, map[string]interface{}{
			"id":         m.ID,
			"key":        m.Key,
			"value":      m.Value,
			"layer":      m.Layer,
			"importance": m.Importance,
			"source":     m.Source,
			"tags":       m.Tags,
			"status":     m.Status,
			"score":      math.Round(scored[i].Score*1000) / 1000,
			"created_at": m.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at": m.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return result, nil
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 1 {
				tokens = append(tokens, current.String())
			}
			current.Reset()
		}
	}
	if current.Len() > 1 {
		tokens = append(tokens, current.String())
	}

	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "can": true, "shall": true, "must": true,
		"of": true, "in": true, "to": true, "for": true, "with": true,
		"on": true, "at": true, "from": true, "by": true, "about": true,
		"as": true, "into": true, "through": true, "during": true, "before": true,
		"after": true, "above": true, "below": true, "between": true, "and": true,
		"or": true, "not": true, "but": true, "if": true, "then": true,
		"de": true, "le": true, "la": true, "les": true, "un": true,
		"une": true, "des": true, "du": true, "et": true, "en": true,
	}

	filtered := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if !stopWords[t] {
			filtered = append(filtered, t)
		}
	}

	return filtered
}

func computeTF(doc []string, queryTokens []string) map[string]float64 {
	tf := make(map[string]float64)
	if len(doc) == 0 {
		return tf
	}

	docFreq := make(map[string]int)
	for _, t := range doc {
		docFreq[t]++
	}

	querySet := make(map[string]bool)
	for _, t := range queryTokens {
		querySet[t] = true
	}

	for token, count := range docFreq {
		if querySet[token] {
			tf[token] = float64(count) / float64(len(doc))
		}
	}

	return tf
}

func computeIDF(docs [][]string, queryTokens []string) map[string]float64 {
	idf := make(map[string]float64)
	n := float64(len(docs))

	docContains := make(map[string]int)
	for _, doc := range docs {
		seen := make(map[string]bool)
		for _, t := range doc {
			seen[t] = true
		}
		for t := range seen {
			docContains[t]++
		}
	}

	for _, token := range queryTokens {
		if count, ok := docContains[token]; ok && count > 0 {
			idf[token] = math.Log(n / float64(count))
		} else {
			idf[token] = 0
		}
	}

	return idf
}
