package kaleido

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	. "github.com/rupc/Enforcer/atypes"
)

type Provider struct{}

func (p *Provider) Distill(block interface{}) (*DistilledBlock, error) {
	return nil, nil
}

func (p *Provider) Extract(block interface{}) []*SpecialTransaction {
	return nil
}

func (p *Provider) Tailor(tx SpecialTransaction) interface{} {
	return nil
}

func (p *Provider) Submit(tx SpecialTransaction) TxID {
	return ""
}

func (p *Provider) GetNewBlockListener(chainID, srcAddr string) <-chan *DistilledBlock {
	return nil
}

func (p *Provider) Initialize(chainID, blockFetchAddress, txSendAddress string) error {
	return nil
}

func (p *Provider) GetBlockByNumber(number uint64) *DistilledBlock {
	urlStr := "http://0.0.0.0:5000/api/v1/GetBlock/" + strconv.FormatInt(int64(number), 10)
	resp, err := http.Get(urlStr)
	// resp, err := http.Get("http://localhost/get-block")
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(data))
	block := &DistilledBlock{}
	// if err := json.Unmarshal(data, &block); err != nil {
	// panic(err)
	// }
	return block
}

func (p *Provider) SendTransaction(tx SpecialTransaction) string {
	urlStr := "http://0.0.0.0:5000/api/v1/SendTransaction"
	pbytes, _ := json.Marshal(tx)
	buff := bytes.NewBuffer(pbytes)
	resp, err := http.Post(urlStr, "application/json", buff)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	// Response 체크.
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	str := string(respBody)
	// println(str)
	return str
}

// func (p *Provider) Transform() *atypes.Block {

// }
