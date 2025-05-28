package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/nutsdb/nutsdb"
)

// APIKeyStore defines the interface for API key storage operations
type APIKeyStore interface {
	SaveAPIKey(email, apiKey string, expiry ...time.Duration) error
	GetAPIKey(email string) (string, error)
	DeleteAPIKey(email string) error
	HasAPIKey(email string) (bool, error)
	GetDefaultExpiry() time.Duration
	Close() error
}

// CacheConfig holds configuration for the NutsDB cache
type CacheConfig struct {
	DBPath        string
	BucketName    string
	DefaultExpiry time.Duration
}

// APIKeyEntry represents an API key with metadata
type APIKeyEntry struct {
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

// Cache represents the cache service
type Cache struct {
	db     *nutsdb.DB
	config CacheConfig
}

// NewCache creates a new cache instance with configuration from environment variables
func NewCache() (*Cache, error) {
	config := CacheConfig{
		DBPath:        "./data/cache",
		BucketName:    "apikeys",
		DefaultExpiry: time.Duration(24) * time.Hour,
	}

	// Get configuration from environment variables with defaults
	dbPath := os.Getenv("CACHE_DB_PATH")
	bucketName := os.Getenv("CACHE_BUCKET_NAME")
	defaultExpiryStr := os.Getenv("CACHE_DEFAULT_EXPIRY_HOURS")

	if len(dbPath) > 0 {
		config.DBPath = dbPath
	}

	if len(bucketName) > 0 {
		config.BucketName = bucketName
	}

	if len(defaultExpiryStr) > 0 {
		defaultExpiry, err := strconv.Atoi(defaultExpiryStr)
		if err == nil {
			config.DefaultExpiry = time.Duration(defaultExpiry) * time.Hour
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(config.DBPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Initialize NutsDB
	opt := nutsdb.DefaultOptions
	opt.Dir = config.DBPath
	db, err := nutsdb.Open(opt)
	if err != nil {
		return nil, fmt.Errorf("failed to open NutsDB: %w", err)
	}

	return &Cache{
		db:     db,
		config: config,
	}, nil
}

// SaveAPIKey stores an API key for the given email with expiration
func (c *Cache) SaveAPIKey(email, apiKey string, expiry ...time.Duration) error {
	// Use default expiry if not provided
	expiryDuration := c.config.DefaultExpiry
	if len(expiry) > 0 {
		expiryDuration = expiry[0]
	}

	entry := APIKeyEntry{
		APIKey:    apiKey,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal API key entry: %w", err)
	}

	err = c.db.Update(func(tx *nutsdb.Tx) error {
		if tx.ExistBucket(nutsdb.DataStructureBTree, c.config.BucketName) {
			return nil // Bucket already exists, no need to create it
		}
		return tx.NewBucket(nutsdb.DataStructureBTree, c.config.BucketName)
	})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	err = c.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(c.config.BucketName, []byte(email), data, uint32(expiryDuration.Seconds()))
	})

	if err != nil {
		return fmt.Errorf("failed to save API key: %w", err)
	}

	return nil
}

func (c *Cache) GetDefaultExpiry() time.Duration {
	return c.config.DefaultExpiry
}

// GetAPIKey retrieves an API key for the given email
func (c *Cache) GetAPIKey(email string) (string, error) {
	var apiKey string

	err := c.db.View(func(tx *nutsdb.Tx) error {
		entry, err := tx.Get(c.config.BucketName, []byte(email))
		if err != nil {
			return err
		}

		var keyEntry APIKeyEntry
		if err := json.Unmarshal(entry, &keyEntry); err != nil {
			return err
		}

		apiKey = keyEntry.APIKey
		return nil
	})

	if err != nil {
		if err == nutsdb.ErrKeyNotFound {
			return "", fmt.Errorf("no API key found for email: %s", email)
		}
		return "", fmt.Errorf("failed to get API key: %w", err)
	}

	return apiKey, nil
}

// DeleteAPIKey removes an API key for the given email
func (c *Cache) DeleteAPIKey(email string) error {
	err := c.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Delete(c.config.BucketName, []byte(email))
	})

	if err != nil {
		if err == nutsdb.ErrKeyNotFound {
			return nil // Key doesn't exist, consider it a success
		}
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	return nil
}

// HasAPIKey checks if an API key exists for the given email
func (c *Cache) HasAPIKey(email string) (bool, error) {
	var exists bool

	err := c.db.View(func(tx *nutsdb.Tx) error {
		_, err := tx.Get(c.config.BucketName, []byte(email))
		if err != nil {
			if err == nutsdb.ErrKeyNotFound {
				exists = false
				return nil
			}
			return err
		}
		exists = true
		return nil
	})

	if err != nil {
		return false, fmt.Errorf("failed to check API key existence: %w", err)
	}

	return exists, nil
}

// Close closes the database connection
func (c *Cache) Close() error {
	if err := c.db.Close(); err != nil {
		return fmt.Errorf("failed to close cache database: %w", err)
	}
	return nil
}
