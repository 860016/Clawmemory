package services

import (
	"log"
	"math"
	"sort"
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

type GraphRAGResult struct {
	MemoryID    uint               `json:"memory_id"`
	Key         string             `json:"key"`
	Value       string             `json:"value"`
	Layer       string             `json:"layer"`
	Importance  float64            `json:"importance"`
	Source      string             `json:"source"`
	Tags        string             `json:"tags"`
	Status      string             `json:"status"`
	Score       float64            `json:"score"`
	ScoreDetail map[string]float64 `json:"score_detail"`
	Paths       []string           `json:"paths,omitempty"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

func (s *SearchService) GraphRAGSearch(userID uint, query string, limit int) ([]GraphRAGResult, error) {
	if limit <= 0 {
		limit = 20
	}

	keywordResults := s.keywordSearch(userID, query, limit*3)
	semanticResults := s.semanticSearch(userID, query, limit*3)
	graphResults := s.graphTraversalSearch(userID, query, limit*2)

	merged := make(map[uint]*GraphRAGResult)

	for _, r := range keywordResults {
		merged[r.MemoryID] = &GraphRAGResult{
			MemoryID:    r.MemoryID,
			Key:         r.Key,
			Value:       r.Value,
			Layer:       r.Layer,
			Importance:  r.Importance,
			Source:      r.Source,
			Tags:        r.Tags,
			Status:      r.Status,
			Score:       r.Score * 0.3,
			ScoreDetail: map[string]float64{"keyword": r.Score * 0.3},
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		}
	}

	for _, r := range semanticResults {
		if existing, ok := merged[r.MemoryID]; ok {
			existing.Score += r.Score * 0.4
			existing.ScoreDetail["semantic"] = r.Score * 0.4
		} else {
			merged[r.MemoryID] = &GraphRAGResult{
				MemoryID:    r.MemoryID,
				Key:         r.Key,
				Value:       r.Value,
				Layer:       r.Layer,
				Importance:  r.Importance,
				Source:      r.Source,
				Tags:        r.Tags,
				Status:      r.Status,
				Score:       r.Score * 0.4,
				ScoreDetail: map[string]float64{"semantic": r.Score * 0.4},
				CreatedAt:   r.CreatedAt,
				UpdatedAt:   r.UpdatedAt,
			}
		}
	}

	for _, r := range graphResults {
		if existing, ok := merged[r.MemoryID]; ok {
			existing.Score += r.Score * 0.3
			existing.ScoreDetail["graph"] = r.Score * 0.3
			if len(r.Paths) > 0 {
				existing.Paths = append(existing.Paths, r.Paths...)
			}
		} else {
			merged[r.MemoryID] = &GraphRAGResult{
				MemoryID:    r.MemoryID,
				Key:         r.Key,
				Value:       r.Value,
				Layer:       r.Layer,
				Importance:  r.Importance,
				Source:      r.Source,
				Tags:        r.Tags,
				Status:      r.Status,
				Score:       r.Score * 0.3,
				ScoreDetail: map[string]float64{"graph": r.Score * 0.3},
				Paths:       r.Paths,
				CreatedAt:   r.CreatedAt,
				UpdatedAt:   r.UpdatedAt,
			}
		}
	}

	for _, r := range merged {
		boost := r.Importance * 0.15
		r.Score += boost
		r.ScoreDetail["importance_boost"] = boost
		r.Score = math.Round(r.Score*1000) / 1000
		for k, v := range r.ScoreDetail {
			r.ScoreDetail[k] = math.Round(v*1000) / 1000
		}
	}

	all := make([]GraphRAGResult, 0, len(merged))
	for _, r := range merged {
		all = append(all, *r)
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Score > all[j].Score
	})

	if limit > len(all) {
		limit = len(all)
	}

	return all[:limit], nil
}

type internalSearchResult struct {
	MemoryID   uint
	Key        string
	Value      string
	Layer      string
	Importance float64
	Source     string
	Tags       string
	Status     string
	Score      float64
	Paths      []string
	CreatedAt  string
	UpdatedAt  string
}

func (s *SearchService) keywordSearch(userID uint, query string, limit int) []internalSearchResult {
	var memories []models.Memory
	escaped := EscapeLikeQuery(query)
	if err := s.db.Where("user_id = ? AND status != ?", userID, "trashed").
		Where("key LIKE ? OR value LIKE ?", "%"+escaped+"%", "%"+escaped+"%").
		Limit(limit).Find(&memories).Error; err != nil {
		log.Printf("[Search] keyword search db error: %v", err)
		return nil
	}

	results := make([]internalSearchResult, 0, len(memories))
	for _, m := range memories {
		keyScore := 0.0
		valScore := 0.0
		qLower := strings.ToLower(query)
		keyLower := strings.ToLower(m.Key)
		valLower := strings.ToLower(m.Value)

		if keyLower == qLower {
			keyScore = 1.0
		} else if strings.HasPrefix(keyLower, qLower) {
			keyScore = 0.8
		} else if strings.Contains(keyLower, qLower) {
			keyScore = 0.6
		}

		count := strings.Count(valLower, qLower)
		if count > 0 {
			valScore = math.Min(float64(count)*0.2, 0.8)
		}

		score := math.Max(keyScore, valScore)
		if score > 0 {
			results = append(results, internalSearchResult{
				MemoryID:   m.ID,
				Key:        m.Key,
				Value:      m.Value,
				Layer:      m.Layer,
				Importance: m.Importance,
				Source:     m.Source,
				Tags:       m.Tags,
				Status:     m.Status,
				Score:      score,
				CreatedAt:  m.CreatedAt.Format("2006-01-02 15:04:05"),
				UpdatedAt:  m.UpdatedAt.Format("2006-01-02 15:04:05"),
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

func (s *SearchService) semanticSearch(userID uint, query string, limit int) []internalSearchResult {
	var memories []models.Memory
	escaped := EscapeLikeQuery(query)
	if err := s.db.Where("user_id = ? AND status != ?", userID, "trashed").
		Where("key LIKE ? OR value LIKE ?", "%"+escaped+"%", "%"+escaped+"%").
		Limit(500).Find(&memories).Error; err != nil {
		log.Printf("[Search] semantic search db error: %v", err)
		return nil
	}

	if len(memories) < 50 {
		memories = nil
		if err := s.db.Where("user_id = ? AND status != ?", userID, "trashed").
			Order("importance DESC, access_count DESC").Limit(500).Find(&memories).Error; err != nil {
			log.Printf("[Search] semantic search fallback db error: %v", err)
			return nil
		}
	}

	if len(memories) == 0 {
		return nil
	}

	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		return nil
	}

	docs := make([][]string, len(memories))
	for i, m := range memories {
		docs[i] = tokenize(m.Key + " " + m.Value)
	}

	idf := computeIDF(docs, queryTokens)

	type scored struct {
		Index int
		Score float64
	}

	var scoredList []scored
	for i, m := range memories {
		tf := computeTF(docs[i], queryTokens)
		score := 0.0
		for _, token := range queryTokens {
			score += tf[token] * idf[token]
		}
		score += float64(m.Importance) * 0.1
		score += float64(min(m.AccessCount, 100)) * 0.01

		if score > 0 {
			scoredList = append(scoredList, scored{Index: i, Score: score})
		}
	}

	sort.Slice(scoredList, func(i, j int) bool {
		return scoredList[i].Score > scoredList[j].Score
	})

	if limit > len(scoredList) {
		limit = len(scoredList)
	}

	results := make([]internalSearchResult, 0, limit)
	for i := 0; i < limit; i++ {
		m := memories[scoredList[i].Index]
		results = append(results, internalSearchResult{
			MemoryID:   m.ID,
			Key:        m.Key,
			Value:      m.Value,
			Layer:      m.Layer,
			Importance: m.Importance,
			Source:     m.Source,
			Tags:       m.Tags,
			Status:     m.Status,
			Score:      scoredList[i].Score,
			CreatedAt:  m.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  m.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return results
}

func (s *SearchService) graphTraversalSearch(userID uint, query string, limit int) []internalSearchResult {
	queryTokens := tokenize(query)
	if len(queryTokens) == 0 {
		return nil
	}

	var entities []models.Entity
	if err := s.db.Where("user_id = ?", userID).Limit(500).Find(&entities).Error; err != nil {
		log.Printf("[Search] graph search entities db error: %v", err)
		return nil
	}

	matchedEntityIDs := make(map[uint]bool)
	for _, e := range entities {
		nameLower := strings.ToLower(e.Name)
		descLower := strings.ToLower(e.Description)
		for _, t := range queryTokens {
			if strings.Contains(nameLower, t) || strings.Contains(descLower, t) {
				matchedEntityIDs[e.ID] = true
				break
			}
		}
	}

	if len(matchedEntityIDs) == 0 {
		return nil
	}

	var relations []models.Relation
	if err := s.db.Where("user_id = ?", userID).Limit(2000).Find(&relations).Error; err != nil {
		log.Printf("[Search] graph search relations db error: %v", err)
	}

	adj := make(map[uint][]struct {
		target  uint
		relType string
		weight  float64
	})
	for _, r := range relations {
		adj[r.SourceID] = append(adj[r.SourceID], struct {
			target  uint
			relType string
			weight  float64
		}{target: r.TargetID, relType: r.RelationType, weight: r.Weight})
		adj[r.TargetID] = append(adj[r.TargetID], struct {
			target  uint
			relType string
			weight  float64
		}{target: r.SourceID, relType: r.RelationType, weight: r.Weight})
	}

	entityName := make(map[uint]string)
	for _, e := range entities {
		entityName[e.ID] = e.Name
	}

	type pathResult struct {
		entityID uint
		score    float64
		paths    []string
	}

	pathResults := make(map[uint]*pathResult)

	for startID := range matchedEntityIDs {
		type bfsEntry struct {
			node    uint
			depth   int
			pathStr string
			score   float64
		}

		visited := make(map[uint]bool)
		queue := []bfsEntry{{
			node:    startID,
			depth:   0,
			pathStr: entityName[startID],
			score:   1.0,
		}}

		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]

			if current.depth >= 3 {
				continue
			}

			if visited[current.node] && current.depth > 0 {
				continue
			}
			visited[current.node] = true

			neighbors := adj[current.node]
			for _, nb := range neighbors {
				if visited[nb.target] {
					continue
				}

				nextScore := current.score * nb.weight * 0.7
				nextPath := current.pathStr + " -> " + entityName[nb.target]

				if existing, ok := pathResults[nb.target]; ok {
					if nextScore > existing.score {
						existing.score = nextScore
						existing.paths = append(existing.paths[:0], nextPath)
					} else {
						existing.paths = append(existing.paths, nextPath)
					}
				} else {
					pathResults[nb.target] = &pathResult{
						entityID: nb.target,
						score:    nextScore,
						paths:    []string{nextPath},
					}
				}

				queue = append(queue, bfsEntry{
					node:    nb.target,
					depth:   current.depth + 1,
					pathStr: nextPath,
					score:   nextScore,
				})
			}
		}
	}

	var allMemories []models.Memory
	if err := s.db.Where("user_id = ? AND status != ?", userID, "trashed").
		Order("importance DESC").Limit(500).Find(&allMemories).Error; err != nil {
		log.Printf("[Search] graph search memories db error: %v", err)
		return nil
	}

	entityToMemories := make(map[uint][]uint)
	if len(entities) > 0 && len(allMemories) > 0 {
		memKeyLower := make([]string, len(allMemories))
		memValLower := make([]string, len(allMemories))
		for i, m := range allMemories {
			memKeyLower[i] = strings.ToLower(m.Key)
			memValLower[i] = strings.ToLower(m.Value)
		}
		for _, e := range entities {
			eNameLower := strings.ToLower(e.Name)
			for i, m := range allMemories {
				if strings.Contains(memKeyLower[i], eNameLower) ||
					strings.Contains(memValLower[i], eNameLower) {
					entityToMemories[e.ID] = append(entityToMemories[e.ID], m.ID)
				}
			}
		}
	}

	memScore := make(map[uint]float64)
	memPaths := make(map[uint][]string)
	for entityID, pr := range pathResults {
		if memIDs, ok := entityToMemories[entityID]; ok {
			for _, mid := range memIDs {
				memScore[mid] += pr.score
				memPaths[mid] = append(memPaths[mid], pr.paths...)
			}
		}
	}

	type memResult struct {
		id    uint
		score float64
		paths []string
	}

	var memResults []memResult
	for mid, score := range memScore {
		memResults = append(memResults, memResult{id: mid, score: score, paths: memPaths[mid]})
	}

	sort.Slice(memResults, func(i, j int) bool {
		return memResults[i].score > memResults[j].score
	})

	if limit > len(memResults) {
		limit = len(memResults)
	}

	memIDSet := make(map[uint]bool)
	for _, mr := range memResults[:limit] {
		memIDSet[mr.id] = true
	}

	var matchedMems []models.Memory
	for _, m := range allMemories {
		if memIDSet[m.ID] {
			matchedMems = append(matchedMems, m)
		}
	}

	results := make([]internalSearchResult, 0, len(matchedMems))
	for _, m := range matchedMems {
		score := 0.0
		if s, ok := memScore[m.ID]; ok {
			score = s
		}
		paths := memPaths[m.ID]
		if len(paths) > 3 {
			paths = paths[:3]
		}

		results = append(results, internalSearchResult{
			MemoryID:   m.ID,
			Key:        m.Key,
			Value:      m.Value,
			Layer:      m.Layer,
			Importance: m.Importance,
			Source:     m.Source,
			Tags:       m.Tags,
			Status:     m.Status,
			Score:      score,
			Paths:      paths,
			CreatedAt:  m.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:  m.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

func (s *SearchService) SemanticSearch(userID uint, query string, limit int) ([]map[string]interface{}, error) {
	var memories []models.Memory
	escaped := EscapeLikeQuery(query)
	_ = s.db.Where("user_id = ? AND status != ?", userID, "trashed").
		Where("key LIKE ? OR value LIKE ?", "%"+escaped+"%", "%"+escaped+"%").
		Limit(500).Find(&memories).Error

	if len(memories) < 50 {
		memories = nil
		_ = s.db.Where("user_id = ? AND status != ?", userID, "trashed").
			Order("importance DESC, access_count DESC").Limit(500).Find(&memories).Error
	}

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

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

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
