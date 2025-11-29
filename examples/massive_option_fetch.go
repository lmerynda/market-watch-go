package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	massive "github.com/massive-com/client-go/v2/rest"
	mm "github.com/massive-com/client-go/v2/rest/models"
)

// This small example fetches 1-minute aggregates for an option contract and writes CSV:
// timestamp,open,high,low,close,volume
func main() {
	var (
		apiKey         = os.Getenv("MASSIVE_API_KEY")
		underlying     = flag.String("underlying", "COIN", "Underlying ticker (e.g. COIN)")
		contractType   = flag.String("type", "call", "contract type: call|put")
		expirationDate = flag.String("expiration", "2025-11-14", "Expiration date (YYYY-MM-DD)")
		strike         = flag.Float64("strike", 265, "Strike price")
		outFile        = flag.String("out", "option_aggregates.csv", "Output CSV file")
		mult           = flag.Int("mult", 1, "Aggregate multiplier (1 = 1 minute)")
	)
	flag.Parse()

	if apiKey == "" {
		log.Fatal("MASSIVE_API_KEY not set in environment")
	}

	c := massive.New(apiKey)

	// Find matching contract (take first)
	// parse expiration date
	expDate := mustParseDate(*expirationDate)

	params := mm.ListOptionsContractsParams{}.
		WithUnderlyingTicker(mm.EQ, *underlying).
		WithContractType(string(mmmContractType(*contractType))).
		WithExpirationDate(mm.EQ, mm.Date(expDate)).
		WithStrikePrice(mm.EQ, *strike).
		WithExpired(true).
		WithLimit(10)
	iter := c.ListOptionsContracts(context.Background(), params)
	var contractTicker string
	for iter.Next() {
		item := iter.Item()
		// match strike approximately (float compare)
		if floatEquals(item.StrikePrice, *strike) {
			contractTicker = item.Ticker
			break
		}
		// fallback to first
		if contractTicker == "" {
			contractTicker = item.Ticker
		}
	}
	if iter.Err() != nil {
		log.Fatalf("error listing contracts: %v", iter.Err())
	}

	if contractTicker == "" {
		log.Fatalf("no contract found for %s %s %v", *underlying, *expirationDate, *strike)
	}

	// Build from/to for the date
	from := mustParseDate(*expirationDate)
	to := from.Add(24 * time.Hour)

	// Get 1-minute aggs
	base := mm.ListAggsParams{
		Ticker:     contractTicker,
		Multiplier: *mult,
		Timespan:   mm.Minute,
		From:       mm.Millis(from),
		To:         mm.Millis(to),
	}
	aggParams := base.WithAdjusted(true).WithOrder(mm.Asc).WithLimit(50000)

	outIter := c.ListAggs(context.Background(), aggParams)

	f, err := os.Create(*outFile)
	if err != nil {
		log.Fatalf("failed to create output: %v", err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()

	// header
	w.Write([]string{"timestamp", "open", "high", "low", "close", "volume"})

	for outIter.Next() {
		a := outIter.Item()
		ts := time.Time(a.Timestamp).UTC().Format(time.RFC3339)
		w.Write([]string{
			ts,
			fmt.Sprintf("%.6f", a.Open),
			fmt.Sprintf("%.6f", a.High),
			fmt.Sprintf("%.6f", a.Low),
			fmt.Sprintf("%.6f", a.Close),
			strconv.FormatFloat(a.Volume, 'f', -1, 64),
		})
	}
	if outIter.Err() != nil {
		log.Fatalf("failed to fetch aggregates: %v", outIter.Err())
	}

	log.Printf("Wrote aggregates for %s to %s", contractTicker, *outFile)
}

func floatEquals(a, b float64) bool {
	const eps = 1e-8
	return (a-b) < eps && (b-a) < eps
}

func mustParseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		log.Fatalf("invalid date: %v", err)
	}
	return t
}

// mmmContractType returns model-compatible contract type
func mmmContractType(t string) mm.ContractType {
	if t == "put" {
		return mm.ContractPut
	}
	return mm.ContractCall
}
