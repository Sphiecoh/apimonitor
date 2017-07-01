package db

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// Store represent a bolt db datastore
type Store struct {
	*bolt.DB
	TestBucket   []byte
	ResultBucket []byte
}
type ApiResult struct {
	Status    int           `json:"status,omitempty"`
	Name      string        `json:"name,omitempty"`
	Error     string        `json:"error,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`
	Timestamp time.Time     `json:"timestamp,omitempty"`
	TestID    string        `json:"test_id,omitempty"`
}
type ApiTest struct {
	URL  string `json:"url,omitempty"`
	Cron string `json:"cron,omitempty"`
	Name string `json:"name,omitempty"`
	ID   string `json:"id,omitempty"`
}
type Results []*ApiResult

// NewStore creates a new store.
func NewStore(path string) (*Store, error) {
	config := &bolt.Options{Timeout: 1 * time.Second}
	d, err := bolt.Open(path, 0600, config)
	if err != nil {
		return nil, errors.Wrapf(err, "Opening store %s failed", path)
	}

	return &Store{d, []byte("tests"), []byte("results")}, nil
}

// CreateBuckets creates a new buckets
func (db *Store) CreateBuckets() error {
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(db.TestBucket)
		if err != nil {
			return errors.Wrapf(err, "Could not create test bucket")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(db.ResultBucket)
		if err != nil {
			return errors.Wrapf(err, "Could not create result bucket")
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

// RemoveBucket deletes bucket by name
func (db *Store) RemoveBucket(name []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket(name)
	})
}
func (db *Store) Put(key string, bucket, data []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		return b.Put([]byte(key), data)

	})
}

func (s *Store) GetAllTests() ([]*ApiTest, error) {
	result := make([]*ApiTest, 0)
	err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.TestBucket)

		b.ForEach(func(k, v []byte) error {
			apitest := new(ApiTest)
			err := json.Unmarshal(v, apitest)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to unmashal json for key %s", string(k)))
			}
			result = append(result, apitest)
			return nil
		})
		return nil
	})

	return result, err
}

func (s *Store) GetResultsByTest(id string) ([]*ApiResult, error) {
	result := make([]*ApiResult, 0)
	err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.ResultBucket)
		res := b.Get([]byte(id))
		if len(res) > 0 {
			err := json.Unmarshal(res, &result)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to unmashal json for key %s", id))
			}

		}
		return nil
	})

	return result, err
}
func (s *Store) DeleteTest(id string) error {
	return s.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.TestBucket)
		return b.Delete([]byte(id))
	})
}

// Run runs the API test
func (test *ApiTest) Run() *ApiResult {
	start := time.Now()
	result := &ApiResult{
		Name:      test.Name,
		Timestamp: time.Now(),
		Status:    500,
		TestID:    test.ID,
	}
	response, err := http.DefaultClient.Get(test.URL)
	result.Duration = time.Since(start)
	if err != nil {
		result.Error = err.Error()
		return result
	} else if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		res, _ := ioutil.ReadAll(response.Body)

		result.Error = string(res)
	}
	result.Status = response.StatusCode
	return result
}

func (s *Store) SaveResult(result *ApiResult) error {
	results, err := s.GetResultsByTest(result.TestID)

	if err != nil {
		return errors.Wrapf(err, "Could not get results for test %s", result.TestID)
	}
	results = append(results, result)
	data, jsonerr := json.Marshal(results)
	if jsonerr != nil {
		return errors.Wrapf(jsonerr, "Failed to mashal results")
	}
	return s.Put(result.TestID, s.ResultBucket, data)

}

func GenerateID() string {
	return uuid.NewV4().String()
}
