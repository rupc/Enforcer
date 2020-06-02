package fetcher

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/rupc/Enforcer/atypes"
	"github.com/tidwall/gjson"
)

const configFile = "./config_fetch.json"

type Fetcher struct {
	client *redis.Client
	config gjson.Result
}

func GetConfig() string {
	dat, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Print("error", err)
	}

	return string(dat)
}

func (f *Fetcher) Initialize(r *redis.Client) {
	// Read config file
	// dat, err := ioutil.ReadFile(configFile)
	// if err != nil {
	// log.Print("error", err)
	// }

	// config := gjson.Parse(string(dat))

	f.client = r
	// f.config = config
}

// start a subscriber to fetch a new block
func (f *Fetcher) StartNewBlockSubscriber(BlockArrivalChannel chan<- *atypes.DistilledBlock) {
	go func() {
		// ch := pubsub.Channel()
		log.Print("wait event")
		pubsub := f.client.PSubscribe("newblock*")
		defer pubsub.Close()
		ch := pubsub.ChannelSize(100)
		for {
			// msg, _ := pubsub.ReceiveMessage()
			msg := <-ch
			// get a block number
			number, _ := strconv.ParseInt(msg.Payload, 10, 64)

			// get a block
			fetchedBlock := f.FetchBlock(number)

			BlockArrivalChannel <- fetchedBlock
			// log.Printf("received pattern[%s], payload[%s]", number, msg.Pattern, msg.Payload)
			log.Printf("received newblock[%+v], pattern[%s], payload[%s]", fetchedBlock, msg.Pattern, msg.Payload)
		}
	}()
}

func (f *Fetcher) Publish() {
	go func() {
		for {
			// err := f.client.Set("xx11", "0", 0).Err()
			f.client.Publish("newblock0000000000003000", "3000")
			// log.Print("write xx11")
			// if err != nil {
			// log.Print(err)
			// }
			time.Sleep(150 * time.Millisecond)
		}
	}()
}

func (f *Fetcher) FetchBlock(number int64) *atypes.DistilledBlock {
	fetchedBlock := f.FetchBlockRange(number, number)
	return fetchedBlock[0]
}

func (f *Fetcher) FetchBlockRange(start, end int64) []*atypes.DistilledBlock {
	var fillteredBlocks []*atypes.DistilledBlock

	for i := start; i <= end; i++ {
		blockKey := fmt.Sprintf("%016d", i)

		// fmt.Printf("before read %s\n", blockKey)
		block, err := f.client.Get(blockKey).Result()
		// fmt.Printf("after read %s\n", blockKey)

		if err != nil {
			fmt.Printf("nil value on %d\n", i)
			panic(err)
		}

		parsed := gjson.Parse(block)

		filteredBlock := &atypes.DistilledBlock{
			ChainID:    "basschain",
			Number:     parsed.Get("header.number").Uint(),
			TotalTxNum: len(parsed.Get("data.data").Array()),
		}

		var specialTxes []*atypes.SpecialTransaction

		for _, tx := range parsed.Get("data.data").Array() {
			// check tx is ascc ?
			if tx.Get("payload.data.actions.0.payload.action.proposal_response_payload.extension.chaincode_id.name").String() == "ascc" {
				value := gjson.Parse(tx.Get("payload.data.actions.0.payload.action.proposal_response_payload.extension.results.ns_rwset.0.rwset.writes.0.value").String())
				member := value.Get("member").String()
				number := value.Get("number").Uint()

				// check member is in target_member

				specialTxes = append(specialTxes, &atypes.SpecialTransaction{
					Member: member,
					Number: number,
				})
			}
		}
		// fmt.Printf("len(filteredBlock) %d\n", len(specialTxes))
		filteredBlock.TxSet = specialTxes
		fillteredBlocks = append(fillteredBlocks, filteredBlock)
		// fmt.Printf("one loop\n")
	}

	// for _, b := range fillteredBlocks {
	// fmt.Printf("%+v\n", b)
	// for _, tx := range b.TxSet {
	// fmt.Printf("%+v %+v\n", b, tx)
	// }
	// }

	return fillteredBlocks
}

// func main() {
//     config := gjson.Parse(GetConfig())
//     client := redis.NewClient(&redis.Options{
//         Addr:     config.Get("redis_db_path").String(),
//         Password: "", // no password set
//         DB:       0,  // use default DB
//     })

//     pong, err := client.Ping().Result()

//     if err != nil {
//         fmt.Println(pong, err)
//     }

//     fetcher := &Fetcher{
//         client: client,
//         config: config,
//     }

//     fetcher.NewBlockSubscriber()
//     // fetcher.Publish()

//     // fetcher.FetchBlockRange(13740, 13742)

//     // cnt := 0
//     for {
//         // fmt.Printf("%d\n", cnt)
//         // cnt++
//         time.Sleep(1 * time.Second)
//     }

//     // fetcher.FetchBlockRange(config.Get("start_blk_num").Int(), config.Get("end_blk_num").Int())

//     // err = client.Set("key", "value", 0).Err()
//     // if err != nil {
//     // panic(err)
//     // }

//     // val, err := client.Get("0000000000005000").Result()
//     // if err != nil {
//     // panic(err)
//     // }
//     // fmt.Println("key", val)

//     // val2, err := client.Get("key2").Result()
//     // if err == redis.Nil {
//     // fmt.Println("key2 does not exist")
//     // } else if err != nil {
//     // panic(err)
//     // } else {
//     // fmt.Println("key2", val2)
//     // }
// }
