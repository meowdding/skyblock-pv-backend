package auctions

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	routeUtils "skyblock-pv-backend/routes/utils"
	"skyblock-pv-backend/utils"
	"skyblock-pv-backend/utils/nbt"
	"slices"
	"strings"
	"time"
)

func GetCachedAuctions(ctx *routeUtils.RouteContext) (*string, error) {
	data, err := ctx.GetFromCache(nil, "auctions", "cached")
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func FetchAll(ctx *routeUtils.RouteContext) error {
	if !ctx.IsCached("auctions", "index") {
		fetch(*ctx)
	} else {
		println("Data cached :3")
	}

	auctions, err := ctx.GetAll("auctions.index")
	if err != nil {
		return err
	}
	auctionList := make([]AuctionStruct, len(auctions))
	for i, auctionJson := range auctions {
		var auction AuctionStruct
		err = json.Unmarshal([]byte(auctionJson), &auction)
		if err != nil {
			return err
		}
		auctionList[i] = auction
	}

	items := calculateAverage(auctionList)
	data, err := json.Marshal(*items)
	if err != nil {
		return err
	}

	_ = ctx.AddToCache("auctions", "cached", data, time.Hour*2)
	println("Finished fetching & caching auctions.")
	return nil
}

func calculateAverage(auctions []AuctionStruct) *map[string]ItemInfo {
	var items = make(map[string][]int64)

	fmt.Printf("Searching %d auctions\n", len(auctions))
	pets := make([]AuctionStruct, 0)
	for _, auction := range auctions {
		if !auction.Bin {
			continue
		}
		item, err := auction.GetItem()
		if err != nil {
			continue
		}
		sbId := item.GetSbId()
		if sbId == nil {
			continue
		}
		if *sbId == "RUNE" {
			continue // not supported atm :)
		}
		if *sbId == "PET" {
			pets = append(pets, auction)
			continue
		}
		priceList := items[*sbId]
		if priceList == nil {
			priceList = make([]int64, 0)
		}

		items[*sbId] = append(priceList, auction.StartingBid)
	}

	actualItems := make(map[string]ItemInfo)
	for key, cost := range items {

		slices.Sort(cost)
		actualItems[key] = ItemInfo{
			Highest: cost[len(cost)-1],
			Lowest:  cost[0],
			Median:  calculateMedian(cost),
			Mean:    calculateArithmeticMean(cost),
		}
	}

	return &actualItems
}

func write(name string, a interface{}) {
	file, err := os.Create(fmt.Sprintf("%s.json", name))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	d, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		panic(err)
	}

	file.Write(d)
}

func calculateArithmeticMean(list []int64) float64 {
	var sum float64 = 0
	for _, i := range list {
		sum += float64(i)
	}
	return sum / float64(len(list))
}

func calculateMedian(list []int64) int64 {
	if len(list) == 0 {
		return 0
	} else if len(list) <= 2 {
		return list[0]
	} else if len(list)%2 == 0 {
		middle := len(list) / 2
		return (list[middle] + list[middle+1]) / 2
	} else {
		return list[len(list)/2]
	}
}

type ItemInfo struct {
	Lowest  int64   `json:"lowest"`
	Highest int64   `json:"highest"`
	Median  int64   `json:"median"`
	Mean    float64 `json:"mean"`
}

func cacheAll(ctx *routeUtils.RouteContext, auction *AuctionRespond) {
	for _, auctionStruct := range auction.Auctions {
		cache(*ctx, &auctionStruct)
	}
}

func cache(ctx routeUtils.RouteContext, auction *AuctionStruct) {
	data, err := json.Marshal(auction)
	if err != nil {
		println(err.Error())
		return
	}
	_ = ctx.AddToCache("auctions.index", auction.Id, data, time.Hour*7)
}

func fetch(ctx routeUtils.RouteContext) {
	data, err := fetchPage(ctx, 0)
	if err != nil {
		println("so sad :C")
		return
	}
	cacheAll(&ctx, data)

	for i := range data.TotalPages {
		if i == 0 {
			continue
		}
		data, err = fetchPage(ctx, i)
		if err != nil {
			fmt.Printf("so sad^%d", i)
			return
		}
		cacheAll(&ctx, data)
	}

	err = ctx.AddToCache("auctions", "index", "yay :3", time.Hour*7)
	if err != nil {
		panic(err)
	}
}

func fetchPage(ctx routeUtils.RouteContext, page int) (*AuctionRespond, error) {
	fmt.Printf("Fetching Page number %d\n", page)
	res, err := routeUtils.GetFromHypixel(ctx, fmt.Sprintf("/v2/skyblock/auctions?page=%d", page), false)
	if err != nil {
		println("Error fetching data from hypixel")
		return nil, err
	}

	var auctionRespond AuctionRespond
	err = json.NewDecoder(strings.NewReader(*res)).Decode(&auctionRespond)
	if err != nil {
		println("Error decoding data from hypixel")
		return nil, err
	}

	return &auctionRespond, nil
}

type AuctionRespond struct {
	Success       bool            `json:"success"`
	Page          int             `json:"page"`
	TotalPages    int             `json:"totalPages"`
	TotalAuctions int             `json:"totalAuctions"`
	LastUpdated   int64           `json:"lastUpdated"`
	Auctions      []AuctionStruct `json:"auctions"`
}

type AuctionStruct struct {
	Bin         bool   `json:"bin"`
	Id          string `json:"uuid"`
	StartingBid int64  `json:"starting_bid"`
	ItemName    string `json:"item_name"`
	ItemLore    string `json:"item_lore"`
	ItemBytes   string `json:"item_bytes"`
}

func (auction *AuctionStruct) GetItem() (*utils.Item, error) {
	data, err := base64.StdEncoding.DecodeString(auction.ItemBytes)
	if err != nil {
		return nil, err
	}
	tag, err := nbt.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return &utils.Item{Compound: tag.AsCompound().Get("i").AsList().GetValues()[0].(*nbt.Compound)}, nil
}
