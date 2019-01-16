package busetabot

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/trace"
)

var HTTPFormat = propagation.HTTPFormat{}

// BusStop represents a bus stop.
type BusStop struct {
	BusStopCode string
	RoadName    string
	Description string
	Latitude    float64
	Longitude   float64
}

type NearbyBusStop struct {
	BusStop
	Distance float64
}

type InMemoryBusStopRepository struct {
	busStops    []BusStop
	busStopsMap map[string]*BusStop
	synonyms    map[string]string
}

func parentSpanFromContext(ctx context.Context) (spanContext trace.SpanContext, ok bool) {
	var r *http.Request
	r, ok = ctx.Value(requestKey{}).(*http.Request)
	if !ok {
		return
	}
	spanContext, ok = HTTPFormat.SpanContextFromRequest(r)
	return
}

func (r *InMemoryBusStopRepository) Get(ID string) *BusStop {
	busStop, ok := r.busStopsMap[ID]
	if ok {
		return busStop
	}
	return nil
}

// Nearby returns up to limit bus stops which are within a given radius from a point as well as their
// distance from that point.
func (r *InMemoryBusStopRepository) Nearby(ctx context.Context, lat, lon, radius float64, limit int) (nearby []NearbyBusStop) {
	if parent, ok := parentSpanFromContext(ctx); ok {
		_, span := trace.StartSpanWithRemoteParent(ctx, "InMemoryBusStopRepository/Nearby", parent)
		defer span.End()
	}

	for _, bs := range r.busStops {
		distance := EuclideanDistanceAtEquator(lat, lon, bs.Latitude, bs.Longitude)
		if distance <= radius {
			nearby = append(nearby, NearbyBusStop{
				BusStop:  bs,
				Distance: distance,
			})
		}
	}
	sort.Slice(nearby, func(i, j int) bool {
		return nearby[i].Distance < nearby[j].Distance
	})
	if limit <= 0 || limit > len(nearby) {
		limit = len(nearby)
	}
	return nearby[:limit]
}

func lowercaseTokens(tokens []string) (lowercased []string) {
	lowercased = make([]string, len(tokens))
	for i, token := range tokens {
		lowercased[i] = strings.ToLower(token)
	}
	return
}

func replaceSynonyms(synonyms map[string]string, tokens []string) []string {
	if synonyms == nil {
		return tokens
	}
	results := make([]string, len(tokens))
	for i, token := range tokens {
		if synonym, ok := synonyms[token]; ok {
			results[i] = synonym
		} else {
			results[i] = token
		}
	}
	return results
}

func (r *InMemoryBusStopRepository) Search(ctx context.Context, query string, limit int) []BusStop {
	if parent, ok := parentSpanFromContext(ctx); ok {
		_, span := trace.StartSpanWithRemoteParent(ctx, "InMemoryBusStopRepository/Search", parent)
		defer span.End()
	}

	if query == "" {
		if limit <= 0 || limit > len(r.busStops) {
			limit = len(r.busStops)
		}
		return r.busStops[:limit]
	}
	tokens := strings.Fields(query)
	if len(tokens) == 1 && len(tokens[0]) == 5 {
		code := tokens[0]
		if busStop, ok := r.busStopsMap[code]; ok {
			return []BusStop{
				*busStop,
			}
		}
	}
	tokens = replaceSynonyms(r.synonyms, lowercaseTokens(tokens))
	var hits []struct {
		Score int
		BusStop
	}
	for _, busStop := range r.busStops {
		descTokens := lowercaseTokens(strings.Fields(busStop.Description))
		roadTokens := lowercaseTokens(strings.Fields(busStop.RoadName))
		score := 0
		for _, token := range tokens {
			for _, descToken := range descTokens {
				if token == descToken {
					score += 2
				}
			}
			for _, roadToken := range roadTokens {
				if token == roadToken {
					score += 1
				}
			}
		}
		if score > 0 {
			hits = append(hits, struct {
				Score int
				BusStop
			}{Score: score, BusStop: busStop})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].Score == hits[j].Score {
			return hits[i].BusStopCode < hits[j].BusStopCode
		}
		return hits[i].Score > hits[j].Score
	})
	if limit <= 0 || limit > len(hits) {
		limit = len(hits)
	}
	results := make([]BusStop, limit)
	for i := 0; i < limit; i++ {
		results[i] = hits[i].BusStop
	}
	return results
}

func NewInMemoryBusStopRepository(busStops []BusStop, synonyms map[string]string) *InMemoryBusStopRepository {
	busStopsMap := make(map[string]*BusStop)
	for i := range busStops {
		bs := busStops[i]
		busStopsMap[bs.BusStopCode] = &bs
	}
	return &InMemoryBusStopRepository{
		busStops:    busStops,
		busStopsMap: busStopsMap,
		synonyms:    synonyms,
	}
}

func NewInMemoryBusStopRepositoryFromFile(path, synonymsPath string) (*InMemoryBusStopRepository, error) {
	busStopsJSONFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "error opening bus stops JSON file")
	}
	var busStopsJSON []BusStop
	err = json.NewDecoder(busStopsJSONFile).Decode(&busStopsJSON)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding bus stops JSON file")
	}
	return NewInMemoryBusStopRepository(busStopsJSON, nil), nil
}
