package auctions

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	routeUtils "skyblock-pv-backend/routes/utils"
	"skyblock-pv-backend/utils"
	"skyblock-pv-backend/utils/nbt"
	"strings"
	"time"
)

func FetchAll(ctx *routeUtils.RouteContext) {
	if !ctx.IsCached("auctions", "index") {
		fetch(*ctx)
	} else {
		println("Data cached :3")
	}

	auctions, err := ctx.GetAll("auctions.index")
	if err != nil {
		panic(err)
	}
	for _, auctionJson := range auctions {
		var auction AuctionStruct
		err = json.Unmarshal([]byte(auctionJson), &auction)
		if err != nil {
			panic(err)
		}
		item, err := auction.GetItem()
		if err != nil {
			fmt.Printf("%+v\n", err)
			continue
		}
		fmt.Printf("%s\n", item.GetSbId())
	}
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
	_ = ctx.AddToCache("auctions.index", auction.Id, data, time.Hour*3)
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

	err = ctx.AddToCache("auctions", "index", "yay :3", time.Hour*5)
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
	return &utils.Item{Compound: tag.AsCompound()}, nil
}
