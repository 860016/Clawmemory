package services

import (
	"log"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"clawmemory/internal/models"

	"gorm.io/gorm"
)

type SearchService struct {
	db    *gorm.DB
	cache *SearchCache
}

func NewSearchService(db *gorm.DB) *SearchService {
	return &SearchService{
		db:    db,
		cache: NewSearchCache(1*time.Minute, 200),
	}
}

// SetCache allows AppContainer to inject a shared cache instance.
func (s *SearchService) SetCache(c *SearchCache) {
	s.cache = c
}

// InvalidateCache clears cached search results for a user.
func (s *SearchService) InvalidateCache(userID uint) {
	if s.cache != nil {
		s.cache.Invalidate(userID)
	}
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

	// Check cache
	if s.cache != nil {
		if cached, ok := s.cache.Get(userID, query, limit); ok {
			return cached, nil
		}
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

	result := all[:limit]

	// Write to cache
	if s.cache != nil {
		s.cache.Set(userID, query, limit, result)
	}

	return result, nil
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

	var currentToken strings.Builder
	var lastCharType int // 0=unknown, 1=cjk, 2=latin/digit, 3=separator

	for _, r := range text {
		charType := 0
		switch {
		case isCJK(r):
			charType = 1
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			charType = 2
		default:
			charType = 3
		}

		if charType != lastCharType && currentToken.Len() > 0 {
			token := currentToken.String()
			if lastCharType == 1 {
				tokens = append(tokens, cjkBigram(token)...)
			} else if lastCharType == 2 && len(token) > 1 && !isStopWord(token) {
				tokens = append(tokens, token)
			}
			currentToken.Reset()
		}

		if charType == 3 {
			lastCharType = charType
			continue
		}

		currentToken.WriteRune(r)
		lastCharType = charType
	}

	if currentToken.Len() > 0 {
		token := currentToken.String()
		if lastCharType == 1 {
			tokens = append(tokens, cjkBigram(token)...)
		} else if lastCharType == 2 && len(token) > 1 && !isStopWord(token) {
			tokens = append(tokens, token)
		}
	}

	return tokens
}

// isCJK checks if a rune is a CJK ideograph (excludes punctuation ranges)
func isCJK(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		(r >= 0x3400 && r <= 0x4DBF) || // CJK Extension A
		(r >= 0x2E80 && r <= 0x2EFF) || // CJK Radicals Supplement
		(r >= 0xF900 && r <= 0xFAFF) // CJK Compatibility Ideographs
}

// cjkBigram splits CJK text into bigram tokens for search
// "编程语言" → ["编程", "程语", "语言"]
// "语言" → ["语言"] (2 chars: output whole)
// "语" → ["语"] (1 char: output as-is)
func cjkBigram(s string) []string {
	runes := []rune(s)
	if len(runes) <= 2 {
		return []string{s}
	}

	var tokens []string
	for i := 0; i < len(runes)-1; i++ {
		bigram := string(runes[i : i+2])
		// skip bigrams containing stop words
		if !isChineseStopWord(string(runes[i])) && !isChineseStopWord(string(runes[i+1])) {
			tokens = append(tokens, bigram)
		}
	}

	return tokens
}

func isStopWord(token string) bool {
	return englishStopWords[token] || chineseStopWords[token]
}

func isChineseStopWord(char string) bool {
	return chineseStopWords[char]
}

var englishStopWords = map[string]bool{
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

var chineseStopWords = map[string]bool{
	"的": true, "了": true, "在": true, "是": true, "我": true,
	"有": true, "和": true, "就": true, "不": true, "人": true,
	"都": true, "一": true, "上": true, "也": true,
	"很": true, "到": true, "说": true, "要": true, "去": true,
	"你": true, "会": true, "着": true, "看": true,
	"好": true, "这": true, "那": true, "他": true, "她": true,
	"它": true, "们": true, "把": true, "被": true, "从": true,
	"让": true, "给": true, "向": true, "比": true, "对": true,
	"与": true, "或": true, "而": true, "但": true, "却": true,
	"又": true, "还": true, "已": true, "所": true, "其": true,
	"此": true, "之": true, "等": true, "能": true, "得": true,
	"地": true, "吗": true, "吧": true, "呢": true, "啊": true,
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
