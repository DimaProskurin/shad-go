//go:build !solution

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Record struct {
	Athlete string `json:"athlete"`
	Age     int    `json:"age"`
	Country string `json:"country"`
	Year    int    `json:"year"`
	Date    string `json:"date"`
	Sport   string `json:"sport"`
	Gold    int    `json:"gold"`
	Silver  int    `json:"silver"`
	Bronze  int    `json:"bronze"`
	Total   int    `json:"total"`
}

func LoadData(path string) ([]Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var recs []Record
	if err := json.Unmarshal(b, &recs); err != nil {
		return nil, err
	}
	return recs, nil
}

type Year = int
type Sport = string

type MedalsStats struct {
	Gold   int `json:"gold"`
	Silver int `json:"silver"`
	Bronze int `json:"bronze"`
	Total  int `json:"total"`
}

type SportStats map[Year]MedalsStats

func (s SportStats) TotalByYears() MedalsStats {
	total := MedalsStats{}
	for _, v := range s {
		total.Gold += v.Gold
		total.Silver += v.Silver
		total.Bronze += v.Bronze
		total.Total += v.Total
	}
	return total
}

type AthleteInfo struct {
	Name    string
	Country string
	// sport -> {"year" -> medals}
	MedalsBySport map[Sport]SportStats
}

type AthleteInfoOutput struct {
	Athlete      string               `json:"athlete"`
	Country      string               `json:"country"`
	Medals       MedalsStats          `json:"medals"`
	MedalsByYear map[Year]MedalsStats `json:"medals_by_year"`
}

type CountryInfoOutput struct {
	Country string `json:"country"`
	MedalsStats
}

func (a *AthleteInfo) InfoOutput(sportKey *Sport) AthleteInfoOutput {
	totalMedals := MedalsStats{}
	totalYearsInfo := make(map[Year]MedalsStats)
	for sport, yearsInfo := range a.MedalsBySport {
		if sportKey != nil && sport != *sportKey {
			continue
		}
		for year, stats := range yearsInfo {
			prev := totalYearsInfo[year]
			totalYearsInfo[year] = MedalsStats{
				Gold:   prev.Gold + stats.Gold,
				Silver: prev.Silver + stats.Silver,
				Bronze: prev.Bronze + stats.Bronze,
				Total:  prev.Total + stats.Total,
			}
			totalMedals.Gold += stats.Gold
			totalMedals.Silver += stats.Silver
			totalMedals.Bronze += stats.Bronze
			totalMedals.Total += stats.Total
		}
	}
	return AthleteInfoOutput{
		Athlete:      a.Name,
		Country:      a.Country,
		Medals:       totalMedals,
		MedalsByYear: totalYearsInfo,
	}
}

type Service struct {
	name2info  map[string]AthleteInfo
	rawRecords []Record
}

func New(recs []Record) *Service {
	name2info := prepareInfo(recs)
	return &Service{
		name2info:  name2info,
		rawRecords: recs,
	}
}

func (s *Service) athleteInfo(w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	name := params.Get("name")

	info, exists := s.name2info[name]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "athlete %q not found", name)
		return
	}

	b, err := json.Marshal(info.InfoOutput(nil))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

const DefaultLimit = 3

func (s *Service) topAthletesInSport(w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	sportKey := params.Get("sport")
	limit := DefaultLimit
	if params.Has("limit") {
		limitS := params.Get("limit")
		limit, err = strconv.Atoi(limitS)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, "wrong limit %q", limitS)
			return
		}
	}

	athleteInfos := s.getAllBySport(sportKey)
	if len(athleteInfos) == 0 {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "sport %q not found", sportKey)
		return
	}
	slices.SortFunc(athleteInfos, func(a, b AthleteInfo) int {
		aT := a.MedalsBySport[sportKey].TotalByYears()
		bT := b.MedalsBySport[sportKey].TotalByYears()
		switch {
		case aT.Gold != bT.Gold:
			return bT.Gold - aT.Gold
		case aT.Silver != bT.Silver:
			return bT.Silver - aT.Silver
		case aT.Bronze != bT.Bronze:
			return bT.Bronze - aT.Bronze
		default:
			return strings.Compare(a.Name, b.Name)
		}
	})
	athleteInfos = athleteInfos[:min(limit, len(athleteInfos))]

	out := make([]AthleteInfoOutput, 0)
	for _, a := range athleteInfos {
		out = append(out, a.InfoOutput(&sportKey))
	}
	b, err := json.Marshal(out)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func (s *Service) topCountriesInYear(w http.ResponseWriter, r *http.Request) {
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	yearS := params.Get("year")
	yearKey, err := strconv.Atoi(yearS)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "wrong year %q", yearS)
		return
	}
	limit := DefaultLimit
	if params.Has("limit") {
		limitS := params.Get("limit")
		limit, err = strconv.Atoi(limitS)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, "wrong limit %q", limitS)
			return
		}
	}

	countryStats := make(map[string]MedalsStats)
	for _, record := range s.rawRecords {
		if record.Year == yearKey {
			prev := countryStats[record.Country]
			countryStats[record.Country] = MedalsStats{
				Gold:   prev.Gold + record.Gold,
				Silver: prev.Silver + record.Silver,
				Bronze: prev.Bronze + record.Bronze,
				Total:  prev.Total + record.Total,
			}
		}
	}
	if len(countryStats) == 0 {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "year %q not found", yearKey)
		return
	}

	countryInfos := make([]CountryInfoOutput, 0)
	for country, stats := range countryStats {
		countryInfos = append(countryInfos, CountryInfoOutput{
			Country:     country,
			MedalsStats: stats,
		})
	}
	slices.SortFunc(countryInfos, func(a, b CountryInfoOutput) int {
		switch {
		case a.Gold != b.Gold:
			return b.Gold - a.Gold
		case a.Silver != b.Silver:
			return b.Silver - a.Silver
		case a.Bronze != b.Bronze:
			return b.Bronze - a.Bronze
		default:
			return strings.Compare(a.Country, b.Country)
		}
	})
	countryInfos = countryInfos[:min(limit, len(countryInfos))]

	b, err := json.Marshal(countryInfos)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(b)
}

func (s *Service) getAllBySport(key Sport) []AthleteInfo {
	res := make([]AthleteInfo, 0)
	for _, info := range s.name2info {
		if _, exists := info.MedalsBySport[key]; exists {
			res = append(res, info)
		}
	}
	return res
}

func prepareInfo(recs []Record) map[string]AthleteInfo {
	name2info := make(map[string]AthleteInfo)
	for _, rec := range recs {
		athleteInfo, exists := name2info[rec.Athlete]
		if !exists {
			medals := make(map[Sport]SportStats)
			medals[rec.Sport] = make(SportStats)
			medals[rec.Sport][rec.Year] = MedalsStats{
				Gold:   rec.Gold,
				Silver: rec.Silver,
				Bronze: rec.Bronze,
				Total:  rec.Total,
			}
			name2info[rec.Athlete] = AthleteInfo{
				Name:          rec.Athlete,
				Country:       rec.Country,
				MedalsBySport: medals,
			}
			continue
		}

		sportInfo, sportExists := athleteInfo.MedalsBySport[rec.Sport]
		if !sportExists {
			athleteInfo.MedalsBySport[rec.Sport] = make(SportStats)
			athleteInfo.MedalsBySport[rec.Sport][rec.Year] = MedalsStats{
				Gold:   rec.Gold,
				Silver: rec.Silver,
				Bronze: rec.Bronze,
				Total:  rec.Total,
			}
			continue
		}

		yearInfo := sportInfo[rec.Year]
		sportInfo[rec.Year] = MedalsStats{
			Gold:   yearInfo.Gold + rec.Gold,
			Silver: yearInfo.Silver + rec.Silver,
			Bronze: yearInfo.Bronze + rec.Bronze,
			Total:  yearInfo.Total + rec.Total,
		}
	}
	return name2info
}

func main() {
	port := flag.String("port", "", "port to run server on")
	dataPath := flag.String("data", "", "path to json file with data")
	flag.Parse()

	recs, err := LoadData(*dataPath)
	if err != nil {
		log.Fatalf("loading data %v", err)
	}

	srv := New(recs)

	mux := http.NewServeMux()
	mux.Handle("/athlete-info", http.HandlerFunc(srv.athleteInfo))
	mux.Handle("/top-athletes-in-sport", http.HandlerFunc(srv.topAthletesInSport))
	mux.Handle("/top-countries-in-year", http.HandlerFunc(srv.topCountriesInYear))

	if err := http.ListenAndServe(":"+*port, mux); err != nil {
		log.Fatalf("running server %v", err)
	}
}
