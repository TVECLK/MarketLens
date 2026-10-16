package crawler

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"hash/fnv"
	"log"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

const shingleSize = 3

const maxHash32 = ^uint32(0)

// JobData is the text fields a MinHash signature is built from.
type JobData struct {
	Employer    string
	JobRole     string
	Location    string
	Description string
}

// LSHIndex is one LSH band's bucket key.
type LSHIndex struct {
	BandNo    int
	BucketKey string
}

// GenerateMinHashAndLSH computes the MinHash signature and LSH bucket
// keys for a job.
func GenerateMinHashAndLSH(job JobData, numPerm, numBands int) ([]uint32, []LSHIndex) {
	text := strings.ToLower(strings.Join([]string{
		job.Employer, job.JobRole, job.Location, job.Description,
	}, " "))
	text = strings.Join(strings.Fields(text), " ")

	shingles := shingle(text, shingleSize)

	sig := make([]uint32, numPerm)
	for i := range sig {
		sig[i] = maxHash32
	}

	for s := range shingles {
		b := []byte(s)
		for i := 0; i < numPerm; i++ {
			h := seededHash32(uint32(i), b)
			if h < sig[i] {
				sig[i] = h
			}
		}
	}

	hashesPerBand := numPerm / numBands
	lshIndexes := make([]LSHIndex, 0, numBands)
	for band := 0; band < numBands; band++ {
		start := band * hashesPerBand
		end := start + hashesPerBand
		bandHashes := sig[start:end]

		buf := make([]byte, 4*len(bandHashes))
		for i, h := range bandHashes {
			binary.LittleEndian.PutUint32(buf[i*4:], h)
		}
		sum := md5.Sum(buf)
		bucketKey := "b" + strconv.Itoa(band) + "_" + hex.EncodeToString(sum[:])[:16]

		lshIndexes = append(lshIndexes, LSHIndex{BandNo: band, BucketKey: bucketKey})
	}

	return sig, lshIndexes
}

// seededHash32 gives MinHash num_perm independent hash functions by
// mixing the permutation index into an FNV-1a hash alongside the
// shingle bytes.
func seededHash32(seed uint32, shingleBytes []byte) uint32 {
	var seedBuf [4]byte
	binary.LittleEndian.PutUint32(seedBuf[:], seed)

	h := fnv.New32a()
	h.Write(seedBuf[:])
	h.Write(shingleBytes)

	return h.Sum32()
}

// shingle returns unique k-length shingles, indexed by unicode code
// point (rune) to match Python 3's str slicing semantics.
func shingle(text string, k int) map[string]struct{} {
	runes := []rune(text)
	set := make(map[string]struct{})
	for i := 0; i+k <= len(runes); i++ {
		set[string(runes[i:i+k])] = struct{}{}
	}
	return set
}

// locationsCompatible: exact match, then substring, then stopword-
// filtered word overlap.
func locationsCompatible(locA, locB string) bool {
	if locA == "" || locB == "" {
		return true
	}
	if locA == locB {
		return true
	}
	if strings.Contains(locB, locA) || strings.Contains(locA, locB) {
		return true
	}

	stopwords := map[string]struct{}{
		"the": {}, "of": {}, "in": {}, "at": {}, "and": {}, "a": {},
	}
	wordsA := wordSet(locA, stopwords)
	wordsB := wordSet(locB, stopwords)
	for w := range wordsA {
		if _, ok := wordsB[w]; ok {
			return true
		}
	}
	return false
}

func wordSet(s string, stop map[string]struct{}) map[string]struct{} {
	out := make(map[string]struct{})
	for _, w := range strings.Fields(s) {
		if _, isStop := stop[w]; !isStop {
			out[w] = struct{}{}
		}
	}
	return out
}

// ToPQArray converts a []uint32 signature into a pq.Int64Array 
func ToPQArray(sig []uint32) pq.Int64Array {
	out := make(pq.Int64Array, len(sig))
	for i, v := range sig {
		out[i] = int64(int32(v))
	}
	return out
}

// FromPQArray reverses ToPQArray.
func FromPQArray(arr pq.Int64Array) []uint32 {
	out := make([]uint32, len(arr))
	for i, v := range arr {
		out[i] = uint32(int32(v))
	}
	return out
}

// CheckDuplicate looks up candidate jobs sharing any of the given LSH
// bucket keys and returns whether the current signature is a near-
// duplicate of one of them (isDuplicate, matchedJobID — 0 if none).
// Errors from the repository are logged and treated as "no duplicate
// found" rather than propagated, matching the original Python
// try/except behavior.
func CheckDuplicate(
	repo DuplicateLookupRepository,
	lshIndexes []LSHIndex,
	currentSig []uint32,
	incomingLocation string,
	jaccardThreshold float64,
) (bool, int64) {
	bucketKeys := make([]string, len(lshIndexes))
	for i, idx := range lshIndexes {
		bucketKeys[i] = idx.BucketKey
	}

	jobs, err := repo.GetJobsByBucketKeys(bucketKeys)
	if err != nil {
		log.Printf("duplicate lookup failed: %v", err)
		return false, 0
	}

	incLoc := strings.ToLower(strings.TrimSpace(incomingLocation))

	for _, job := range jobs {
		dbLoc := strings.ToLower(strings.TrimSpace(job.Location))
		if !locationsCompatible(dbLoc, incLoc) {
			continue
		}

		dbSig := FromPQArray(job.MetaData.MinhashSignature)
		if len(dbSig) != len(currentSig) {
			continue
		}

		var intersection int
		for i := range currentSig {
			if currentSig[i] == dbSig[i] {
				intersection++
			}
		}
		union := len(currentSig) + len(dbSig) - intersection

		var jaccardValue float64
		if union > 0 {
			jaccardValue = float64(intersection) / float64(union)
		}

		if jaccardValue >= jaccardThreshold {
			log.Printf("duplicate found: jaccard=%.2f matched_job_id=%d", jaccardValue, job.ID)
			return true, int64(job.ID)
		}
	}

	return false, 0
}