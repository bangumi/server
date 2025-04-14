package person

import (
	"context"
	"strconv"
	"time"

	"encoding/json"
	"errors"
	"fmt"

	"github.com/bangumi/server/domain"
	"github.com/bangumi/server/domain/gerr"
	"github.com/bangumi/server/internal/model"
	"github.com/redis/rueidis"
	"github.com/trim21/errgo"
	"go.uber.org/zap"
)

const (
	prefixPersonByID          = "person:id:"
	prefixPersonBySubjectID   = "person:subject:"
	prefixPersonByCharacterID = "person:character:"
	defaultCacheTTL           = time.Hour * 1 // Default cache TTL set to 1 hour
)

// Helper function to generate Redis key for person by ID
func personIDKey(id model.PersonID) string {
	return prefixPersonByID + strconv.FormatUint(uint64(id), 10)
}

// Helper function to generate Redis key for persons related to a subject ID
func personSubjectIDKey(subjectID model.SubjectID) string {
	return prefixPersonBySubjectID + strconv.FormatUint(uint64(subjectID), 10)
}

// Helper function to generate Redis key for persons related to a character ID
func personCharacterIDKey(characterID model.CharacterID) string {
	return prefixPersonByCharacterID + strconv.FormatUint(uint64(characterID), 10)
}

// cachedRepo implements the CachedRepo interface using Redis.
type cachedRepo struct {
	repo Repo // Embed the underlying database repository
	rdb  rueidis.Client
	log  *zap.Logger
}

// NewCachedRepo creates a new cached repository for persons.
// It requires the underlying repository, a Redis client, and a logger.
func NewCachedRepo(repo Repo, rdb rueidis.Client, log *zap.Logger) CachedRepo {
	return &cachedRepo{
		repo: repo,
		rdb:  rdb, // Store the Redis client
		log:  log.Named("person.cachedRepo"),
	}
}

// Get implements CachedRepo.
func (c *cachedRepo) Get(ctx context.Context, id model.PersonID) (model.Person, error) {
	key := personIDKey(id)
	val, err := c.rdb.Do(ctx, c.rdb.B().Get().Key(key).Build()).AsBytes()

	if err == nil {
		// Cache hit
		var person model.Person
		if err := json.Unmarshal(val, &person); err != nil {
			c.log.Error("Failed to unmarshal person from cache", zap.String("key", key), zap.Error(err))
			// Fallback to DB if cache data is corrupted
		} else {
			c.log.Debug("Cache hit for person", zap.String("key", key))
			return person, nil
		}
	}

	if !rueidis.IsRedisNil(err) {
		// Redis error other than not found
		c.log.Error("Failed to get person from cache", zap.String("key", key), zap.Error(err))
		// Fallback to DB on Redis error
	} else {
		c.log.Debug("Cache miss for person", zap.String("key", key))
	}

	// Cache miss or error, fetch from repository
	person, err := c.repo.Get(ctx, id)
	if err != nil {
		// Handle repository errors (e.g., not found)
		// Don't cache "not found" errors unless specifically desired
		if errors.Is(err, gerr.ErrNotFound) {
			return model.Person{}, err
		}
		return model.Person{}, fmt.Errorf("failed to get person from repo: %w", err)
	}

	// Cache the result
	jsonData, err := json.Marshal(person)
	if err != nil {
		c.log.Error("Failed to marshal person for cache", zap.Uint64("id", uint64(id)), zap.Error(err))
		// Return data even if caching fails
		return person, nil
	}

	err = c.rdb.Do(ctx,
		c.rdb.B().Set().Key(key).Value(string(jsonData)).Ex(defaultCacheTTL).Build()).Error()
	if err != nil {
		c.log.Error("Failed to set person in cache", zap.String("key", key), zap.Error(err))
	}

	return person, nil
}

// GetByIDs
func (c *cachedRepo) GetByIDs(ctx context.Context, ids []model.PersonID) (map[model.PersonID]model.Person, error) {
	if len(ids) == 0 {
		return make(map[model.PersonID]model.Person), nil
	}

	results := make(map[model.PersonID]model.Person, len(ids))
	missedIDs := make([]model.PersonID, 0, len(ids))
	keys := make([]string, len(ids))
	keyToIDMap := make(map[string]model.PersonID, len(ids))

	for i, id := range ids {
		key := personIDKey(id)
		keys[i] = key
		keyToIDMap[key] = id
	}

	// Try fetching from cache
	mgetCmd := c.rdb.B().Mget().Key(keys...).Build()
	vals, err := c.rdb.Do(ctx, mgetCmd).ToArray()

	// rueidis MGET returns Nil error if *all* keys are missing
	if err != nil && !rueidis.IsRedisNil(err) {
		c.log.Error("Failed to MGET persons from cache", zap.Error(err))
		// Fallback to fetching all from DB on MGET error
		missedIDs = ids
	} else {
		// Process cache results
		for i, kv := range vals {
			key := keys[i]
			id := keyToIDMap[key]
			if kv.IsNil() {
				missedIDs = append(missedIDs, id)
				c.log.Debug("Cache miss for person", zap.String("key", key))
				continue
			}

			valStr, err := kv.ToString()
			if err != nil {
				c.log.Error("Failed to convert cache result to string", zap.String("key", key), zap.Error(err))
				missedIDs = append(missedIDs, id) // Treat conversion error as miss
				continue
			}

			var person model.Person
			if err := json.Unmarshal([]byte(valStr), &person); err != nil {
				c.log.Error("Failed to unmarshal person from cache", zap.String("key", key), zap.Error(err))
				missedIDs = append(missedIDs, id) // Treat unmarshal error as miss
			} else {
				results[id] = person
				c.log.Debug("Cache hit for person", zap.String("key", key))
			}
		}
	}

	// Fetch missed IDs from repository
	if len(missedIDs) > 0 {
		c.log.Debug("Fetching missed persons from repo", zap.Any("ids", missedIDs))
		repoResults, repoErr := c.repo.GetByIDs(ctx, missedIDs)
		if repoErr != nil {
			// If repo fails, return partial results from cache + error
			// Or return empty map + error depending on desired behavior
			return nil, errgo.Wrap(repoErr, "failed to get persons from repo")
		}

		// Add repo results to the final map and prepare for caching
		msetCmds := make([]rueidis.Completed, 0, len(repoResults))
		for id, person := range repoResults {
			results[id] = person // Add to final results

			jsonData, err := json.Marshal(person)
			if err != nil {
				c.log.Error("Failed to marshal person for cache", zap.Uint64("id", uint64(id)), zap.Error(err))
				continue // Skip caching this one if marshal fails
			}
			key := personIDKey(id)
			setCmd := c.rdb.B().Set().Key(key).Value(string(jsonData)).Ex(defaultCacheTTL).Build()
			msetCmds = append(msetCmds, setCmd)
			c.log.Debug("Prepared person for caching", zap.String("key", key))
		}

		// Cache the newly fetched items (pipeline SET commands)
		if len(msetCmds) > 0 {
			c.log.Debug("Caching fetched persons", zap.Int("count", len(msetCmds)))
			for _, resp := range c.rdb.DoMulti(ctx, msetCmds...) {
				if err := resp.Error(); err != nil {
					// Log individual SET errors, but don't fail the whole operation
					c.log.Error("Failed to set person in cache during MSET simulation", zap.Error(err))
				}
			}
		}
	}

	return results, nil
}

// GetCharacterRelated implements CachedRepo.
func (c *cachedRepo) GetCharacterRelated(
	ctx context.Context,
	characterID model.CharacterID,
) ([]domain.PersonCharacterRelation, error) {
	key := personCharacterIDKey(characterID)
	cmd := c.rdb.B().Get().Key(key).Build()
	val, err := c.rdb.Do(ctx, cmd).ToString()

	if err == nil {
		// Cache hit
		var relations []domain.PersonCharacterRelation
		if err := json.Unmarshal([]byte(val), &relations); err != nil {
			c.log.Error("Failed to unmarshal character relations from cache", zap.String("key", key), zap.Error(err))
			// Fallback to DB
		} else {
			c.log.Debug("Cache hit for character relations", zap.String("key", key))
			return relations, nil
		}
	}

	if !rueidis.IsRedisNil(err) {
		c.log.Error("Failed to get character relations from cache", zap.String("key", key), zap.Error(err))
		// Fallback to DB
	} else {
		c.log.Debug("Cache miss for character relations", zap.String("key", key))
	}

	// Cache miss or error, fetch from repository
	relations, err := c.repo.GetCharacterRelated(ctx, characterID)
	if err != nil {
		// Don't cache errors like "not found"
		if errors.Is(err, gerr.ErrNotFound) {
			return nil, err
		}
		return nil, errgo.Wrap(err, "failed to get character relations from repo")
	}

	// Cache the result (even if empty, to cache the absence of relations for this ID)
	jsonData, err := json.Marshal(relations)
	if err != nil {
		c.log.Error("Failed to marshal character relations for cache", zap.Uint64("characterID", uint64(characterID)), zap.Error(err))
		return relations, nil // Return data even if caching fails
	}

	setCmd := c.rdb.B().Set().Key(key).Value(string(jsonData)).Ex(defaultCacheTTL).Build()
	if setErr := c.rdb.Do(ctx, setCmd).Error(); setErr != nil {
		c.log.Error("Failed to set character relations in cache", zap.String("key", key), zap.Error(setErr))
	} else {
		c.log.Debug("Cached character relations", zap.String("key", key))
	}

	return relations, nil
}

// GetSubjectRelated implements CachedRepo.
func (c *cachedRepo) GetSubjectRelated(ctx context.Context, subjectID model.SubjectID) ([]domain.SubjectPersonRelation, error) {
	key := personSubjectIDKey(subjectID)
	cmd := c.rdb.B().Get().Key(key).Build()
	val, err := c.rdb.Do(ctx, cmd).ToString()

	if err == nil {
		// Cache hit
		var relations []domain.SubjectPersonRelation
		if err := json.Unmarshal([]byte(val), &relations); err != nil {
			c.log.Error("Failed to unmarshal subject relations from cache", zap.String("key", key), zap.Error(err))
			// Fallback to DB
		} else {
			c.log.Debug("Cache hit for subject relations", zap.String("key", key))
			return relations, nil
		}
	}

	if !rueidis.IsRedisNil(err) {
		c.log.Error("Failed to get subject relations from cache", zap.String("key", key), zap.Error(err))
		// Fallback to DB
	} else {
		c.log.Debug("Cache miss for subject relations", zap.String("key", key))
	}

	// Cache miss or error, fetch from repository
	relations, err := c.repo.GetSubjectRelated(ctx, subjectID)
	if err != nil {
		// Don't cache errors like "not found"
		if errors.Is(err, gerr.ErrNotFound) {
			return nil, err
		}
		return nil, errgo.Wrap(err, "failed to get subject relations from repo")
	}

	// Cache the result (even if empty)
	jsonData, err := json.Marshal(relations)
	if err != nil {
		c.log.Error("Failed to marshal subject relations for cache", zap.Uint64("subjectID", uint64(subjectID)), zap.Error(err))
		return relations, nil // Return data even if caching fails
	}

	setCmd := c.rdb.B().Set().Key(key).Value(string(jsonData)).Ex(defaultCacheTTL).Build()
	if setErr := c.rdb.Do(ctx, setCmd).Error(); setErr != nil {
		c.log.Error("Failed to set subject relations in cache", zap.String("key", key), zap.Error(setErr))
	} else {
		c.log.Debug("Cached subject relations", zap.String("key", key))
	}

	return relations, nil
}
