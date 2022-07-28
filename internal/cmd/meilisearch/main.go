package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/meilisearch/meilisearch-go"
	"github.com/schollz/progressbar/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/bangumi/server/internal/model"
	"github.com/bangumi/server/internal/pkg/generic/slice"
	"github.com/bangumi/server/internal/web/res"
)

const pk = "id"

type DumpSubject struct {
	ID       int    `json:"id"`
	Type     uint8  `json:"type"`
	Name     string `json:"name"`
	NameCN   string `json:"name_cn"`
	Infobox  string `json:"infobox"`
	Platform uint16 `json:"platform"`
	Summary  string `json:"summary"`
	Nsfw     bool   `json:"nsfw"`
}

var errSkip = errors.New("skip")

var errNoRedirect = errors.New("auto redirect is disabled")

func main() {
	m, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://root:example@192.168.1.3:27017/"))
	if err != nil {
		panic(err)
	}

	h := resty.New()
	// h.SetProxy("http://127.0.0.1:7890")
	h.SetHeader(fiber.HeaderAuthorization, "Bearer 8XgJpAtY5N1chIKBmrkAnlC4dnsEGeovtQSN827Y")
	h.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
		return errNoRedirect
	}))
	h.SetRetryCount(4)
	// h.SetRedirectPolicy(resty.DomainCheckRedirectPolicy("api.bgm.tv"))

	client := meilisearch.NewClient(meilisearch.ClientConfig{
		Host:    "http://192.168.1.3:7700",
		APIKey:  "masterKey",
		Timeout: time.Second * 5,
	})

	// _, err := client.Index("subjects").UpdateSearchableAttributes(&[]string{
	// 	"name",
	// 	"tag",
	// 	"name_cn",
	// 	"summary",
	// })
	// if err != nil {
	// 	panic(err)
	// }
	//
	// os.Exit(0)

	f, err := os.Open("./subject.jsonlines")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	c := m.Database("bangumi").Collection("subjects")

	var ch = make(chan DumpSubject)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range ch {
				var r, err = GetFullDoc(h, c, v.ID)
				if errors.Is(err, errSkip) {
					continue
				}

				// w, err := wiki.Parse(v.Infobox)
				// if err != nil {
				// 	w = wiki.Wiki{}
				// }

				// doc := map[string]any{
				// 	pk:        v.ID,
				// 	"name":    v.Name,
				// 	"type":    v.Type,
				// 	"name_cn": v.NameCN,
				// 	"wiki":    w,
				// 	"nsfw":    v.Nsfw,
				// 	"sort":    r.Rating.Rank,
				// 	"summary": v.Summary,
				// 	"tag":     slice.Map(r.Tags, func(t res.SubjectTag) string { return t.Name }),
				// 	"poster":  fmt.Sprintf("https://api.bgm.tv/v0/subjects/%d/image?type=large&vv.jpg", v.ID),
				// }

				var date = ""
				if r.Date != nil {
					date = *r.Date
				}

				name := []string{v.Name}
				if r.NameCN != "" {
					name = append(name, r.NameCN)
				}

				doc := FinalSubject{
					ID:           v.ID,
					Summary:      v.Summary,
					Date:         date,
					Tag:          slice.Map(r.Tags, func(t res.SubjectTag) string { return t.Name }),
					Name:         name,
					GamePlatform: nil,
					Record:       Record{},
					Heat:         heat(r.Collection),
					Score:        r.Rating.Score,
					Rank:         r.Rating.Rank,
					Platform:     v.Platform,
					Type:         v.Type,
					NSFW:         v.Nsfw,
				}

				_, err = client.Index("subjects").AddDocuments(doc, pk)
				if err != nil {
					panic(err)
				}
			}
		}()
	}

	// We're recording marks-per-1second
	// counter := ratecounter.NewRateCounter(1 * time.Second)
	// var obj []map[string]any
	scanner := bufio.NewScanner(f)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	b := progressbar.NewOptions(-1,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(30),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
	)

	for scanner.Scan() {
		line := scanner.Text()
		// defer func() {s
		var v DumpSubject
		err := json.Unmarshal([]byte(line), &v)
		if err != nil {
			panic(err)
		}

		if v.ID < 392300 {
			continue
		}

		ch <- v
		// fmt.Println(v.ID)
		b.Set(v.ID)
		// b.Add(1)

		// counter.Incr(1)
		// fmt.Println(v.ID, counter.Rate())
	}

	close(ch)
	wg.Wait()
}

type FileCache struct {
	Res    APISubject `json:"res"`
	ID     int        `json:"id"`
	Status int        `json:"status"`
}

func GetFullDoc(h *resty.Client, c *mongo.Collection, id int) (APISubject, error) {
	f, err := os.Open(fmt.Sprintf(`D:\d\dump\%d.json`, id))
	if err == nil {
		defer f.Close()
		var c FileCache
		if err = json.NewDecoder(f).Decode(&c); err != nil {
			fmt.Println(id)
			panic(err)
		}

		if c.ID == id {
			return c.Res, nil
		}

		fmt.Println("bad file cache", id)
	}

	if !errors.Is(err, os.ErrNotExist) {
		panic(err)
	}

	var ctx, cancel = context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	doc := c.FindOne(ctx, bson.M{"_id": id})
	if doc.Err() != mongo.ErrNoDocuments {
		if doc.Err() != nil && doc.Err() != mongo.ErrNoDocuments {
			panic(doc.Err())
		}

		var r APISubject
		err := doc.Decode(&r)
		if err != nil {
			panic(err)
		}

		return r, nil
	}

	var r APISubject
	fmt.Println("fetch", id, "from http")
	resp, err := h.R().SetResult(&r).Get(fmt.Sprintf("https://api.bgm.tv/v0/subjects/%d", id))
	if err != nil {
		if errors.Is(err, errNoRedirect) {
			return APISubject{}, errSkip
		}
		panic(err)
	}

	if resp.StatusCode() != 200 {
		panic(resp.String())
	}

	_, err = c.UpdateByID(context.Background(), id, bson.M{"$set": r}, options.Update().SetUpsert(true))
	if err != nil {
		panic(err)
	}

	return r, nil
}

type APISubject struct {
	Date          *string                   `json:"date"`
	Platform      *string                   `json:"platform"`
	Image         res.SubjectImages         `json:"images"`
	Summary       string                    `json:"summary"`
	Name          string                    `json:"name"`
	NameCN        string                    `json:"name_cn"`
	Tags          []res.SubjectTag          `json:"tags"`
	Rating        res.Rating                `json:"rating"`
	TotalEpisodes int64                     `json:"total_episodes"`
	Collection    res.SubjectCollectionStat `json:"collection"`
	ID            model.SubjectID           `json:"id"`
	Eps           uint32                    `json:"eps"`
	Volumes       uint32                    `json:"volumes"`
	Redirect      model.SubjectID           `json:"-"`
	Locked        bool                      `json:"locked"`
	NSFW          bool                      `json:"nsfw"`
	TypeID        model.SubjectType         `json:"type"`
}

type FinalSubject struct {
	ID           int      `json:"id"`
	Summary      string   `json:"summary"`
	Date         string   `json:"date,omitempty"`
	Tag          []string `json:"tag,omitempty"`
	Name         []string `json:"name"`
	GamePlatform []string `json:"game_platform"`
	Record       Record   `json:"record"`
	Heat         uint32   `json:"heat,omitempty"`
	Score        float64  `json:"score"`
	Rank         uint32   `json:"rank"`
	Platform     uint16   `json:"platform,omitempty"`
	Type         uint8    `json:"type"`
	NSFW         bool     `json:"nsfw"`
}

type Record struct {
	Date   time.Time   `json:"date"`
	Image  string      `json:"image"`
	Name   string      `json:"name"`
	NameCN string      `json:"name_cn"`
	Tags   []model.Tag `json:"tags"`
	Score  float64     `json:"score"`
	ID     uint32      `json:"id"`
	Rank   uint32      `json:"rank"`
}

func heat(s res.SubjectCollectionStat) uint32 {
	return s.OnHold + s.Doing + s.Dropped + s.Wish + s.Collect
}
