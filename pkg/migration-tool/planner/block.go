package planner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/prompb"
	"github.com/schollz/progressbar/v3"
	"github.com/timescale/promscale/pkg/migration-tool/utils"
)

// Block represents an in-memory storage for data that is fetched by the reader.
type Block struct {
	id                    int64
	mint                  int64 // inclusive.
	maxt                  int64 // exclusive.
	done                  bool
	pbarDescriptionPrefix string
	pbar                  *progressbar.ProgressBar
	timeseries            []*prompb.TimeSeries
	numBytesCompressed    int
	numBytesUncompressed  int
	pbarMux               *sync.Mutex
	plan                  *Plan // We keep a copy of plan so that each block hsa the authority to update the stats of the planner.
}

// Fetch starts fetching the samples from remote read storage based on the matchers.
func (b *Block) Fetch(context context.Context, client *utils.Client, mint, maxt int64, matchers []*labels.Matcher) (*prompb.QueryResult, error) {
	timeRangeMinutesDelta := (maxt - mint) / time.Minute.Milliseconds()
	readRequest, err := utils.CreatePrombQuery(mint, maxt, matchers)
	if err != nil {
		return nil, fmt.Errorf("create promb query: %w", err)
	}
	result, bytesCompressed, bytesUncompressed, err := client.Read(context, readRequest, fmt.Sprintf("%s | time-range: %d mins", b.pbarDescriptionPrefix, timeRangeMinutesDelta))
	if err != nil {
		return nil, fmt.Errorf("executing client-read: %w", err)
	}
	b.timeseries = result.Timeseries
	// We set compressed bytes in block since those are the bytes that will be pushed over the network to the write storage after snappy compression.
	// The pushed bytes are not exactly the bytesCompressed since while pushing, we add the progress metric. But,
	// the size of progress metric along with the sample is negligible. So, it is safe to consider bytesCompressed
	// in such a scenario.
	b.numBytesCompressed = bytesCompressed
	b.numBytesUncompressed = bytesUncompressed
	b.plan.update(bytesUncompressed)
	return result, nil
}

// MergeProgressSeries returns the block's time-series after appending a sample to the progress-metric and merging
// with the time-series of the block.
func (b *Block) MergeProgressSeries(ts *prompb.TimeSeries) []*prompb.TimeSeries {
	b.SetDescription(fmt.Sprintf("pushing %.2f...", float64(b.numBytesCompressed)/float64(utils.Megabyte)), 1)
	ts.Samples = []prompb.Sample{{Timestamp: b.Maxt(), Value: 1}} // One sample per block.
	b.timeseries = append(b.timeseries, ts)
	return b.timeseries
}

func (b *Block) SetDescription(description string, proceed int) {
	if b.pbarDescriptionPrefix == "" {
		return
	}
	b.pbarMux.Lock()
	defer b.pbarMux.Unlock()
	_ = b.pbar.Add(proceed)
	b.pbar.Describe(fmt.Sprintf("%s | %s", b.pbarDescriptionPrefix, description))
}

// Done updates the text and sets the spinner to done.
func (b *Block) Done() error {
	if b.pbarDescriptionPrefix == "" {
		return nil
	}
	b.SetDescription(
		fmt.Sprintf("pushed %.2f MB. Memory footprint: %.2f MB.", float64(b.numBytesCompressed)/float64(utils.Megabyte), float64(b.numBytesUncompressed)/float64(utils.Megabyte)),
		1,
	)
	b.done = true
	if err := b.pbar.Finish(); err != nil {
		return fmt.Errorf("finish block-lifecycle: %w", err)
	}
	return nil
}

// Mint returns the mint of the block (inclusive).
func (b *Block) Mint() int64 {
	return b.mint
}

// Maxt returns the maxt of the block (exclusive).
func (b *Block) Maxt() int64 {
	return b.maxt
}
