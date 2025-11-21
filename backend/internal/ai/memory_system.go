// backend/internal/ai/memory_system.go
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
)

// MemoryType represents different types of memory
type MemoryType string

const (
	MemoryTypeShortTerm MemoryType = "short_term"
	MemoryTypeLongTerm  MemoryType = "long_term"
	MemoryTypeWorking   MemoryType = "working"
	MemoryTypeEpisodic  MemoryType = "episodic"
	MemoryTypeSemantic  MemoryType = "semantic"
	MemoryTypeDeclarative MemoryType = "declarative"
	MemoryTypeProcedural MemoryType = "procedural"
)

// MemoryPriority represents the priority/importance of a memory
type MemoryPriority int

const (
	PriorityLow    MemoryPriority = 1
	PriorityMedium MemoryPriority = 5
	PriorityHigh   MemoryPriority = 9
	PriorityCritical MemoryPriority = 10
)

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
	ID          string              `json:"id"`
	AgentID     string              `json:"agent_id"`
	Type        MemoryType          `json:"type"`
	Content     string              `json:"content"`
	Embedding   []float32           `json:"embedding,omitempty"`
	Timestamp   time.Time           `json:"timestamp"`
	Importance  float64             `json:"importance"` // 0.0-1.0
	Category    string              `json:"category"`   // Used for semantic organization
	Source      string              `json:"source"`     // Source of the memory (user, system, etc.)
	Keywords    []string            `json:"keywords"`
	Metadata    map[string]interface{} `json:"metadata"`
	Expiry      *time.Time          `json:"expiry,omitempty"` // For temporary memories
	AccessCount int                 `json:"access_count"`
	LastAccessed time.Time           `json:"last_accessed"`
	Version     int                 `json:"version"`
}

// MemoryConfig represents configuration for the memory system
type MemoryConfig struct {
	EnableShortTermMemory  bool    `json:"enable_short_term_memory"`
	EnableLongTermMemory   bool    `json:"enable_long_term_memory"`
	ShortTermMemoryLimit   int     `json:"short_term_memory_limit"`
	LongTermMemoryLimit    int     `json:"long_term_memory_limit"`
	MemoryCompression      bool    `json:"memory_compression"`
	CompressionThreshold  float64  `json:"compression_threshold"` // Importance threshold for compression
	AutoSummarization      bool    `json:"auto_summarization"`
	SummarizationInterval  time.Duration `json:"summarization_interval"`
	VectorDimension       int     `json:"vector_dimension"`
	EmbeddingModel        string  `json:"embedding_model"`
	MaxContextWindow      int     `json:"max_context_window"`
	MinMemoryRetention    time.Duration `json:"min_memory_retention"`
	MaxMemoryRetention    time.Duration `json:"max_memory_retention"`
	EnableFuzzySearch     bool    `json:"enable_fuzzy_search"`
	EnableSemanticSearch  bool    `json:"enable_semantic_search"`
	SearchResultLimit     int     `json:"search_result_limit"`
	SearchThreshold       float64  `json:"search_threshold"` // Minimum relevance score
	EnableMemoryPruning   bool    `json:"enable_memory_pruning"`
	PruningInterval       time.Duration `json:"pruning_interval"`
	MinPruningImportance  float64  `json:"min_pruning_importance"` // Min importance to keep during pruning
	EnableMemoryArchival  bool    `json:"enable_memory_archival"`
	ArchivalThreshold     time.Duration `json:"archival_threshold"` // Age threshold for archival
}

// MemoryStorage interface for storing memories
type MemoryStorage interface {
	Save(ctx context.Context, entry *MemoryEntry) error
	Get(ctx context.Context, id string) (*MemoryEntry, error)
	Search(ctx context.Context, agentID string, query string, limit int) ([]*MemoryEntry, error)
	SearchBySimilarity(ctx context.Context, agentID string, queryEmbedding []float32, limit int) ([]*MemoryEntry, error)
	GetByCategory(ctx context.Context, agentID, category string) ([]*MemoryEntry, error)
	GetRecent(ctx context.Context, agentID string, since time.Duration) ([]*MemoryEntry, error)
	GetByType(ctx context.Context, agentID string, memoryType MemoryType) ([]*MemoryEntry, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, agentID string, limit, offset int) ([]*MemoryEntry, error)
	Update(ctx context.Context, entry *MemoryEntry) error
	Close() error
}

// InMemoryStorage implements in-memory storage for memories (for development)
type InMemoryStorage struct {
	memories map[string]*MemoryEntry
	mutex    sync.RWMutex
}

// NewInMemoryStorage creates a new in-memory storage
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		memories: make(map[string]*MemoryEntry),
	}
}

func (ims *InMemoryStorage) Save(ctx context.Context, entry *MemoryEntry) error {
	ims.mutex.Lock()
	defer ims.mutex.Unlock()
	
	ims.memories[entry.ID] = entry
	return nil
}

func (ims *InMemoryStorage) Get(ctx context.Context, id string) (*MemoryEntry, error) {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()
	
	entry, exists := ims.memories[id]
	if !exists {
		return nil, fmt.Errorf("memory entry %s not found", id)
	}
	
	return entry, nil
}

func (ims *InMemoryStorage) Search(ctx context.Context, agentID string, query string, limit int) (*[]MemoryEntry, error) {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()
	
	var results []*MemoryEntry
	
	query = strings.ToLower(query)
	for _, entry := range ims.memories {
		if entry.AgentID == agentID {
			// Simple fuzzy matching - in real implementation, use proper NLP/similarity
			if strings.Contains(strings.ToLower(entry.Content), query) {
				results = append(results, entry)
				
				if limit > 0 && len(results) >= limit {
					break
				}
			}
		}
	}
	
	// Sort by importance/relevance
	sort.Slice(results, func(i, j int) bool {
		return results[i].Importance > results[j].Importance
	})
	
	return results, nil
}

func (ims *InMemoryStorage) SearchBySimilarity(ctx context.Context, agentID string, queryEmbedding []float32, limit int) ([]*MemoryEntry, error) {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()
	
	var results []*MemoryEntry
	
	for _, entry := range ims.memories {
		if entry.AgentID == agentID && entry.Embedding != nil {
			// Calculate cosine similarity
			similarity := calculateCosineSimilarity(queryEmbedding, entry.Embedding)
			
			// Only include if similarity is above threshold (0.7)
			if similarity > 0.7 {
				modifiedEntry := *entry
				modifiedEntry.Metadata["similarity"] = similarity
				results = append(results, &modifiedEntry)
				
				if limit > 0 && len(results) >= limit {
					break
				}
			}
		}
	}
	
	// Sort by similarity
	sort.Slice(results, func(i, j int) bool {
		sim1 := results[i].Metadata["similarity"].(float64)
		sim2 := results[j].Metadata["similarity"].(float64)
		return sim1 > sim2
	})
	
	return results, nil
}

func (ims *InMemoryStorage) GetByCategory(ctx context.Context, agentID, category string) ([]*MemoryEntry, error) {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()
	
	var results []*MemoryEntry
	
	for _, entry := range ims.memories {
		if entry.AgentID == agentID && entry.Category == category {
			results = append(results, entry)
		}
	}
	
	return results, nil
}

func (ims *InMemoryStorage) GetRecent(ctx context.Context, agentID string, since time.Duration) ([]*MemoryEntry, error) {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()
	
	threshold := time.Now().Add(-since)
	var results []*MemoryEntry
	
	for _, entry := range ims.memories {
		if entry.AgentID == agentID && entry.Timestamp.After(threshold) {
			results = append(results, entry)
		}
	}
	
	return results, nil
}

func (ims *InMemoryStorage) GetByType(ctx context.Context, agentID string, memoryType MemoryType) ([]*MemoryEntry, error) {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()
	
	var results []*MemoryEntry
	
	for _, entry := range ims.memories {
		if entry.AgentID == agentID && entry.Type == memoryType {
			results = append(results, entry)
		}
	}
	
	return results, nil
}

func (ims *InMemoryStorage) Delete(ctx context.Context, id string) error {
	ims.mutex.Lock()
	defer ims.mutex.Unlock()
	
	delete(ims.memories, id)
	return nil
}

func (ims *InMemoryStorage) List(ctx context.Context, agentID string, limit, offset int) ([]*MemoryEntry, error) {
	ims.mutex.RLock()
	defer ims.mutex.RUnlock()
	
	var allMemories []*MemoryEntry
	for _, entry := range ims.memories {
		if entry.AgentID == agentID {
			allMemories = append(allMemories, entry)
		}
	}
	
	// Sort by timestamp (newest first)
	sort.Slice(allMemories, func(i, j int) bool {
		return allMemories[i].Timestamp.After(allMemories[j].Timestamp)
	})
	
	// Apply pagination
	start := offset
	if start >= len(allMemories) {
		return []*MemoryEntry{}, nil
	}
	
	end := start + limit
	if end > len(allMemories) {
		end = len(allMemories)
	}
	
	return allMemories[start:end], nil
}

func (ims *InMemoryStorage) Update(ctx context.Context, entry *MemoryEntry) error {
	ims.mutex.Lock()
	defer ims.mutex.Unlock()
	
	if _, exists := ims.memories[entry.ID]; exists {
		ims.memories[entry.ID] = entry
		return nil
	}
	
	return fmt.Errorf("memory entry %s not found", entry.ID)
}

func (ims *InMemoryStorage) Close() error {
	// Nothing to close for in-memory storage
	return nil
}

// AIMemoryManager manages the AI agent's memory system
type AIMemoryManager struct {
	config        *MemoryConfig
	storage       MemoryStorage
	embedder      embeddings.Embedder
	agentID       string
	shortTerm     []*MemoryEntry // Recent memories for immediate context
	longTerm      map[string]*MemoryEntry // All long-term memories
	mutex         sync.RWMutex
	contextWindow []string // Recent context for immediate use
	searchCache   map[string][]*MemoryEntry
	cacheMutex    sync.RWMutex
}

// NewAIMemoryManager creates a new AI memory manager
func NewAIMemoryManager(config *MemoryConfig, storage MemoryStorage, embedder embeddings.Embedder) *AIMemoryManager {
	if config.ShortTermMemoryLimit == 0 {
		config.ShortTermMemoryLimit = 100 // Default short term limit
	}
	if config.LongTermMemoryLimit == 0 {
		config.LongTermMemoryLimit = 10000 // Default long term limit
	}
	if config.MaxContextWindow == 0 {
		config.MaxContextWindow = 50 // Default context window
	}
	if config.SearchResultLimit == 0 {
		config.SearchResultLimit = 10 // Default search limit
	}
	if config.VectorDimension == 0 {
		config.VectorDimension = 1536 // Default dimensions for OpenAI embeddings
	}
	if config.EmbeddingModel == "" {
		config.EmbeddingModel = "text-embedding-ada-002" // Default embedding model
	}

	manager := &AIMemoryManager{
		config:      config,
		storage:     storage,
		embedder:    embedder,
		shortTerm:   make([]*MemoryEntry, 0),
		longTerm:    make(map[string]*MemoryEntry),
		contextWindow: make([]string, 0),
		searchCache: make(map[string][]*MemoryEntry),
	}

	// Start periodic memory management tasks if enabled
	if config.EnableSummarization {
		go manager.startSummarizationProcess()
	}

	if config.EnableMemoryPruning {
		go manager.startPruningProcess()
	}

	if config.EnableMemoryArchival {
		go manager.startArchivalProcess()
	}

	return manager
}

// AddMemory adds a new memory entry
func (amm *AIMemoryManager) AddMemory(ctx context.Context, memoryType MemoryType, content string, importance float64, category string, source string) error {
	entry := &MemoryEntry{
		ID:          uuid.New().String(),
		Type:        memoryType,
		Content:     content,
		Timestamp:   time.Now(),
		Importance:  importance,
		Category:    category,
		Source:      source,
		Keywords:    extractKeywords(content),
		Metadata:    make(map[string]interface{}),
		AccessCount: 0,
		LastAccessed: time.Now(),
		Version:     1,
	}

	// Generate embedding if we're using semantic search
	if amm.config.EnableSemanticSearch && amm.embedder != nil {
		embedding, err := amm.embedder.EmbedDocuments(ctx, []string{content})
		if err != nil {
			// Log error but don't fail the memory addition
			fmt.Printf("Warning: failed to generate embedding for memory: %v\n", err)
		} else if len(embedding) > 0 {
			entry.Embedding = embedding[0]
		}
	}

	// Save to storage
	if err := amm.storage.Save(ctx, entry); err != nil {
		return fmt.Errorf("failed to save memory entry: %w", err)
	}

	// Add to appropriate memory stores
	amm.mutex.Lock()
	defer amm.mutex.Unlock()

	if memoryType == MemoryTypeShortTerm {
		amm.shortTerm = append(amm.shortTerm, entry)
		// Trim if we exceed the limit
		if len(amm.shortTerm) > amm.config.ShortTermMemoryLimit {
			amm.shortTerm = amm.shortTerm[len(amm.shortTerm)-amm.config.ShortTermMemoryLimit:]
		}
	} else {
		amm.longTerm[entry.ID] = entry
	}

	// Update context window
	amm.contextWindow = append(amm.contextWindow, content)
	if len(amm.contextWindow) > amm.config.MaxContextWindow {
		amm.contextWindow = amm.contextWindow[len(amm.contextWindow)-amm.config.MaxContextWindow:]
	}

	return nil
}

// GetMemory retrieves a specific memory entry
func (amm *AIMemoryManager) GetMemory(ctx context.Context, id string) (*MemoryEntry, error) {
	return amm.storage.Get(ctx, id)
}

// SearchMemories searches for memories relevant to a query
func (amm *AIMemoryManager) SearchMemories(ctx context.Context, query string, limit int) ([]*MemoryEntry, error) {
	if limit == 0 {
		limit = amm.config.SearchResultLimit
	}

	// Check cache first
	cacheKey := fmt.Sprintf("%s_%d", query, limit)
	amm.cacheMutex.RLock()
	if cached, exists := amm.searchCache[cacheKey]; exists {
		amm.cacheMutex.RUnlock()
		return cached, nil
	}
	amm.cacheMutex.RUnlock()

	var results []*MemoryEntry
	var err error

	if amm.config.EnableSemanticSearch && amm.embedder != nil {
		// Generate embedding for the query
		embeddings, err := amm.embedder.EmbedDocuments(ctx, []string{query})
		if err != nil {
			return nil, fmt.Errorf("failed to embed query: %w", err)
		}
		
		if len(embeddings) > 0 {
			results, err = amm.storage.SearchBySimilarity(ctx, amm.agentID, embeddings[0], limit)
			if err != nil {
				return nil, fmt.Errorf("failed to search by similarity: %w", err)
			}
		}
	} else {
		// Use keyword search
		results, err = amm.storage.Search(ctx, amm.agentID, query, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to search memories: %w", err)
		}
	}

	// Apply threshold filter if configured
	if amm.config.SearchThreshold > 0 {
		filteredResults := make([]*MemoryEntry, 0)
		for _, result := range results {
			if similarity, exists := result.Metadata["similarity"]; exists {
				if sim, ok := similarity.(float64); ok {
					if sim >= amm.config.SearchThreshold {
						filteredResults = append(filteredResults, result)
					}
				}
			} else {
				// If no similarity score, assume it passed keyword search and include
				filteredResults = append(filteredResults, result)
			}
		}
		results = filteredResults
	}

	// Update access count and last accessed time
	for _, entry := range results {
		entry.AccessCount++
		entry.LastAccessed = time.Now()
		
		// Update in storage
		if err := amm.storage.Update(ctx, entry); err != nil {
			fmt.Printf("Warning: failed to update access count for memory %s: %v\n", entry.ID, err)
		}
	}

	// Cache results
	amm.cacheMutex.Lock()
	// Keep cache size manageable
	if len(amm.searchCache) > 1000 {
		// Clear oldest entries
		amm.clearOldestCacheEntries()
	}
	amm.searchCache[cacheKey] = results
	amm.cacheMutex.Unlock()

	return results, nil
}

// GetMemoriesByCategory retrieves memories by category
func (amm *AIMemoryManager) GetMemoriesByCategory(ctx context.Context, category string) ([]*MemoryEntry, error) {
	return amm.storage.GetByCategory(ctx, amm.agentID, category)
}

// GetRecentMemories retrieves recent memories
func (amm *AIMemoryManager) GetRecentMemories(ctx context.Context, since time.Duration) ([]*MemoryEntry, error) {
	return amm.storage.GetRecent(ctx, amm.agentID, since)
}

// GetMemoriesByType retrieves memories by type
func (amm *AIMemoryManager) GetMemoriesByType(ctx context.Context, memoryType MemoryType) ([]*MemoryEntry, error) {
	return amm.storage.GetByType(ctx, amm.agentID, memoryType)
}

// GetCurrentContext retrieves the current context window
func (amm *AIMemoryManager) GetCurrentContext() []string {
	amm.mutex.RLock()
	defer amm.mutex.RUnlock()

	// Return a copy to prevent external modification
	contextCopy := make([]string, len(amm.contextWindow))
	copy(contextCopy, amm.contextWindow)
	
	return contextCopy
}

// startSummarizationProcess starts the periodic summarization process
func (amm *AIMemoryManager) startSummarizationProcess() {
	ticker := time.NewTicker(amm.config.SummarizationInterval)
	defer ticker.Stop()

	for range ticker.C {
		amm.performSummarization()
	}
}

// performSummarization performs memory summarization
func (amm *AIMemoryManager) performSummarization() {
	if !amm.config.AutoSummarization {
		return
	}

	// In a real implementation, this would:
	// 1. Identify groups of related memories
	// 2. Summarize them using an LLM
	// 3. Store the summary as a new memory
	// 4. Optionally remove the original detailed memories
	
	// For now, we'll just log this
	fmt.Println("Periodic memory summarization completed")
}

// startPruningProcess starts the periodic memory pruning process
func (amm *AIMemoryManager) startPruningProcess() {
	if amm.config.PruningInterval == 0 {
		return
	}

	ticker := time.NewTicker(amm.config.PruningInterval)
	defer ticker.Stop()

	for range ticker.C {
		amm.performPruning()
	}
}

// performPruning performs memory pruning to free up space
func (amm *AIMemoryManager) performPruning() {
	if !amm.config.EnableMemoryPruning {
		return
	}

	// In a real implementation, this would:
	// 1. Identify memories below the pruning importance threshold
	// 2. Check their last access time
	// 3. Remove the oldest low-importance memories
	
	// For now, we'll just log this
	fmt.Println("Periodic memory pruning completed")
}

// startArchivalProcess starts the periodic memory archival process
func (amm *AIMemoryManager) startArchivalProcess() {
	if amm.config.ArchivalThreshold == 0 {
		return
	}

	ticker := time.NewTicker(24 * time.Hour) // Check daily
	defer ticker.Stop()

	for range ticker.C {
		amm.performArchival()
	}
}

// performArchival performs memory archival for old memories
func (amm *AIMemoryManager) performArchival() {
	if !amm.config.EnableMemoryArchival {
		return
	}

	// In a real implementation, this would:
	// 1. Identify memories older than the archival threshold
	// 2. Move them to archival storage (different system or compressed format)
	// 3. Update references to point to archived location
	
	// For now, we'll just log this
	fmt.Println("Periodic memory archival completed")
}

// calculateCosineSimilarity calculates cosine similarity between two vectors
func calculateCosineSimilarity(vec1, vec2 []float32) float64 {
	if len(vec1) != len(vec2) {
		return 0.0
	}

	var dotProduct, magnitude1, magnitude2 float64
	for i := 0; i < len(vec1); i++ {
		dotProduct += float64(vec1[i] * vec2[i])
		magnitude1 += float64(vec1[i] * vec1[i])
		magnitude2 += float64(vec2[i] * vec2[i])
	}

	magnitude1 = math.Sqrt(magnitude1)
	magnitude2 = math.Sqrt(magnitude2)

	if magnitude1 == 0 || magnitude2 == 0 {
		return 0.0
	}

	return dotProduct / (magnitude1 * magnitude2)
}

// extractKeywords extracts keywords from content
func extractKeywords(content string) []string {
	words := strings.Fields(content)
	
	// Remove common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, 
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "about": true,
		"as": true, "from": true, "up": true, "into": true, "through": true,
		"over": true, "after": true, "under": true, "above": true, "below": true,
		"this": true, "that": true, "these": true, "those": true,
	}
	
	keywords := make([]string, 0)
	
	for _, word := range words {
		word = strings.ToLower(strings.Trim(word, ".,!?;:()[]{}"))
		if len(word) > 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	
	return uniqueStrings(keywords)
}

// uniqueStrings returns a slice with duplicate strings removed
func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// clearOldestCacheEntries clears the oldest half of cache entries
func (amm *AIMemoryManager) clearOldestCacheEntries() {
	// In a real implementation, we would track when cache entries were added
	// and remove the oldest ones
	// For now, we'll just clear half the entries randomly
	
	keys := make([]string, 0, len(amm.searchCache))
	for k := range amm.searchCache {
		keys = append(keys, k)
	}
	
	// Remove first half of keys alphabetically (as a simple "oldest" approach)
	sort.Strings(keys)
	for i, key := range keys {
		if i >= len(keys)/2 {
			break
		}
		delete(amm.searchCache, key)
	}
}

// Close shuts down the memory manager
func (amm *AIMemoryManager) Close() error {
	return amm.storage.Close()
}

// GetMemoryStats returns statistics about the memory system
func (amm *AIMemoryManager) GetMemoryStats(ctx context.Context) (map[string]interface{}, error) {
	amm.mutex.RLock()
	defer amm.mutex.RUnlock()

	stats := map[string]interface{}{
		"short_term_count": len(amm.shortTerm),
		"long_term_count":  len(amm.longTerm),
		"context_window_size": len(amm.contextWindow),
		"total_cached_searches": len(amm.searchCache),
		"config": map[string]interface{}{
			"enable_short_term": amm.config.EnableShortTermMemory,
			"enable_long_term": amm.config.EnableLongTermMemory,
			"short_term_limit": amm.config.ShortTermMemoryLimit,
			"long_term_limit": amm.config.LongTermMemoryLimit,
			"max_context_window": amm.config.MaxContextWindow,
			"enable_compression": amm.config.MemoryCompression,
			"enable_semantic_search": amm.config.EnableSemanticSearch,
			"enable_fuzzy_search": amm.config.EnableFuzzySearch,
		},
		"timestamp": time.Now().Unix(),
	}

	return stats, nil
}

// CreateCompressedMemory creates a compressed version of multiple memories
func (amm *AIMemoryManager) CreateCompressedMemory(ctx context.Context, memoryIDs []string) error {
	if !amm.config.MemoryCompression {
		return fmt.Errorf("memory compression is not enabled")
	}

	var contents []string
	var totalImportance float64
	var categories []string
	var sources []string

	for _, id := range memoryIDs {
		memory, err := amm.storage.Get(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get memory %s: %w", id, err)
		}

		contents = append(contents, memory.Content)
		totalImportance += memory.Importance
		categories = append(categories, memory.Category)
		sources = append(sources, memory.Source)
	}

	// In a real implementation, we would use an LLM to compress the information
	// For now, we'll just concatenate the contents
	compressedContent := strings.Join(contents, " ")

	// Calculate average importance
	avgImportance := totalImportance / float64(len(memoryIDs))

	// Create compressed memory
	compressedMemory := &MemoryEntry{
		ID:         uuid.New().String(),
		Type:       MemoryTypeDeclarative,
		Content:    compressedContent,
		Timestamp:  time.Now(),
		Importance: avgImportance,
		Category:   "compressed", // Indicate this is a compressed memory
		Source:     "compression_processor",
		Keywords:   extractKeywords(compressedContent),
		Metadata: map[string]interface{}{
			"original_count": len(memoryIDs),
			"original_ids":   memoryIDs,
			"compression_ratio": float64(len(compressedContent)) / float64(len(contents)),
		},
		AccessCount:  0,
		LastAccessed: time.Now(),
		Version:      1,
	}

	// Save the compressed memory
	if err := amm.storage.Save(ctx, compressedMemory); err != nil {
		return fmt.Errorf("failed to save compressed memory: %w", err)
	}

	// Update context window
	amm.mutex.Lock()
	amm.contextWindow = append(amm.contextWindow, compressedContent)
	if len(amm.contextWindow) > amm.config.MaxContextWindow {
		amm.contextWindow = amm.contextWindow[len(amm.contextWindow)-amm.config.MaxContextWindow:]
	}
	amm.mutex.Unlock()

	return nil
}