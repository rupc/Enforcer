package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/rupc/Enforcer/atypes"
	"github.com/rupc/Enforcer/core/pa"
)

type SpecialTx struct {
	Member string `json:"member"`
	Number string `json:"number"`
}

type FilteredBlock struct {
	Number      string      `json:"number"`
	SpecialTxes []SpecialTx `json:"special_txes"`
	TotalTx     uint64      `json:"total_tx"`
}

func main() {

	// Read data file
	// For each block b, execute DoPrefixAgreement(b)
	inputDataFiles := []string{
		"./fidelity0_1.json",
	}

	outputDataFiles := []string{
		"./fidelity0_1.out",
	}

	// Select a client profile
	// fileIndex := 4

	for fileIndex := 0; fileIndex < len(inputDataFiles); fileIndex++ {
		paConfig := pa.PrefixAgreementConfig{
			ChainID:         "mychannel",
			NumberOfMembers: 10,
		}

		prefixAgreement := pa.InitializePrefixAgreement(paConfig)

		// Read input filtered block data
		inputData, err := ioutil.ReadFile(inputDataFiles[fileIndex])
		if err != nil {
			fmt.Println(err)
			return
		}

		// Export output prefix analysis data
		outputFile, err := os.OpenFile(outputDataFiles[fileIndex], os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		}

		datawriter := bufio.NewWriter(outputFile)

		var FilteredBlocks []FilteredBlock
		err = json.Unmarshal(inputData, &FilteredBlocks)

		if err != nil {
			fmt.Println(err)
			return
		}

		for _, v := range FilteredBlocks {
			// fmt.Printf("%s\n", v.Number)
			// fmt.Printf("%+v\n", v.SpecialTxes)
			// fmt.Printf("%d\n", v.TotalTx)

			var specialTxes []*atypes.SpecialTransaction

			for _, t := range v.SpecialTxes {
				number, _ := strconv.ParseUint(t.Number, 10, 64)
				tx := atypes.SpecialTransaction{
					ChainID:   "mychannel",
					Number:    number,
					Member:    t.Member,
					Hash:      "0xabcd",
					Signature: nil,
				}
				specialTxes = append(specialTxes, &tx)
			}

			number, _ := strconv.ParseUint(v.Number, 10, 64)
			block := &atypes.DistilledBlock{
				Number:   number,
				ChainID:  "mychannel",
				Hash:     "0xabcd",
				PrevHash: "0xcdef",
				TxSet:    specialTxes,
			}

			// Measure execution time...
			start := time.Now()
			prefixAgreement.DoPrefixAgreement(block)
			paResult := prefixAgreement.DoPrefixAgreement(block)

			elapsed := time.Since(start) / time.Microsecond

			line := fmt.Sprintf("%s %d %d %d %d\n", v.Number, paResult.PreparedHeight, paResult.CommitHeight, len(specialTxes), elapsed)

			// _, _ = datawriter.WriteString(line)
			fmt.Printf("%s", line)
		}

		PrintMemUsage()
		runtime.GC()
		datawriter.Flush()
		outputFile.Close()

	}

}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
