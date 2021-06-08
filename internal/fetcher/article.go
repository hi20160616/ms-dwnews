package fetcher

import (
	"crypto/md5"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/hi20160616/exhtml"
	"github.com/hi20160616/gears"
	"github.com/hi20160616/ms-dwnews/configs"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Article struct {
	Id            string
	Title         string
	Content       string
	WebsiteId     string
	WebsiteDomain string
	WebsiteTitle  string
	UpdateTime    *timestamppb.Timestamp
	U             *url.URL
	raw           []byte
	doc           *html.Node
}

var timeout = func() time.Duration {
	t, err := time.ParseDuration(configs.Data.MS.Timeout)
	if err != nil {
		log.Printf("[%s] timeout init error: %v", configs.Data.MS.Title, err)
		return time.Duration(1 * time.Minute)
	}
	return t
}()

func NewArticle() *Article {
	return &Article{
		WebsiteDomain: configs.Data.MS.Domain,
		WebsiteTitle:  configs.Data.MS.Title,
		WebsiteId:     fmt.Sprintf("%x", md5.Sum([]byte(configs.Data.MS.Domain))),
	}
}

// List get all articles from database
func (a *Article) List() ([]*Article, error) {
	return load()
}

// Get read database and return the data by rawurl.
func (a *Article) Get(id string) (*Article, error) {
	as, err := load()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		if a.Id == id {
			return a, nil
		}
	}
	return nil, fmt.Errorf("[%s] no article with id: %s, url: %s",
		configs.Data.MS.Title, id, a.U.String())
}

func (a *Article) Search(keyword ...string) ([]*Article, error) {
	as, err := load()
	if err != nil {
		return nil, err
	}

	as2 := []*Article{}
	for _, a := range as {
		for _, v := range keyword {
			v = strings.ToLower(strings.TrimSpace(v))
			switch {
			case a.Id == v:
				as2 = append(as2, a)
			case a.WebsiteId == v:
				as2 = append(as2, a)
			case strings.Contains(strings.ToLower(a.Title), v):
				as2 = append(as2, a)
			case strings.Contains(strings.ToLower(a.Content), v):
				as2 = append(as2, a)
			case strings.Contains(strings.ToLower(a.WebsiteDomain), v):
				as2 = append(as2, a)
			case strings.Contains(strings.ToLower(a.WebsiteTitle), v):
				as2 = append(as2, a)
			}
		}
	}
	return as2, nil
}

// fetchArticle fetch article by rawurl
func (a *Article) fetchArticle(rawurl string) (*Article, error) {
	var err error
	a.U, err = url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	// Dail
	a.raw, a.doc, err = exhtml.GetRawAndDoc(a.U, timeout)
	if err != nil {
		return nil, err
	}

	a.Id = fmt.Sprintf("%x", md5.Sum([]byte(rawurl)))

	a.Title, err = a.fetchTitle()
	if err != nil {
		return nil, err
	}

	a.UpdateTime, err = a.fetchUpdateTime()
	if err != nil {
		return nil, err
	}

	// filter work
	if a, err = a.filter(3); errors.Is(err, ErrTimeOverDays) {
		return nil, err
	}

	// content should be the last step to fetch
	a.Content, err = a.fetchContent()
	if err != nil {
		return nil, err
	}

	a.Content, err = a.fmtContent(a.Content)
	if err != nil {
		return nil, err
	}
	return a, nil

}

func (a *Article) fetchTitle() (string, error) {
	n := exhtml.ElementsByTag(a.doc, "title")
	if n == nil {
		return "", fmt.Errorf("[%s] getTitle error, there is no element <title>", configs.Data.MS.Title)
	}
	title := n[0].FirstChild.Data
	if strings.Contains(title, "[图集]") {
		return "", fmt.Errorf("[%s] ignore pic news: %s", configs.Data.MS.Title, a.U.String())
	}
	title = strings.TrimSpace(strings.ReplaceAll(title, "｜多维新闻", ""))
	gears.ReplaceIllegalChar(&title)
	return title, nil
}

func (a *Article) fetchUpdateTime() (*timestamppb.Timestamp, error) {
	if a.doc == nil {
		return nil, errors.Errorf("[%s] fetchUpdateTime: doc is nil: %s", configs.Data.MS.Title, a.U.String())
	}
	metas := exhtml.MetasByName(a.doc, "parsely-pub-date")
	cs := []string{}
	for _, meta := range metas {
		for _, a := range meta.Attr {
			if a.Key == "content" {
				cs = append(cs, a.Val)
			}
		}
	}
	if len(cs) <= 0 {
		return nil, fmt.Errorf("[%s] fetchUpdateTime got nothing.", configs.Data.MS.Title)
	}
	t, err := time.Parse(time.RFC3339, cs[0])
	if err != nil {
		return nil, err
	}
	return timestamppb.New(t), nil
}

func shanghai(t time.Time) time.Time {
	loc := time.FixedZone("UTC", 8*60*60)
	return t.In(loc)
}

var ErrTimeOverDays error = errors.New("article update time out of range")
var ErrSameArticleExist error = errors.New("article title exist")

// filter work for ignore articles by conditions
// TODO: filter redundancy articles by title
func (a *Article) filter(days int) (*Article, error) {
	// if article time out of days, return nil and `ErrTimeOverDays`
	// param days means fetch news during days from befor now.
	during := func(days int, ts *timestamppb.Timestamp) bool {
		t := shanghai(ts.AsTime())
		if time.Now().Day()-t.Day() <= days {
			return true
		}
		return false
	}
	// if during return false rt nil, and error as ErrTimeOverDays
	if !during(days, a.UpdateTime) {
		return nil, ErrTimeOverDays
	}

	return a, nil
}

func (a *Article) fetchContent() (string, error) {
	if a.doc == nil {
		return "", errors.Errorf("[%s] fetchContent: doc is nil: %s",
			configs.Data.MS.Title, a.U.String())
	}
	doc := a.doc
	body := ""
	// Fetch content nodes
	nodes := exhtml.ElementsByTag(doc, "article")
	if len(nodes) == 0 {
		return "", fmt.Errorf("[%s] There is no tag named `<article>` from: %s",
			configs.Data.MS.Title, a.U.String())
	}
	articleDoc := nodes[0].FirstChild
	plist := exhtml.ElementsByTag(articleDoc, "p")
	if articleDoc.FirstChild != nil &&
		articleDoc.FirstChild.Data == "div" { // to fetch the summary block
		// body += fmt.Sprintf("\n > %s  \n", plist[0].FirstChild.Data) // redundant summary
		body += fmt.Sprintf("\n > ")
	}
	for _, v := range plist { // the last item is `推荐阅读：`
		if v.FirstChild == nil {
			continue
		} else if v.FirstChild.FirstChild != nil && v.FirstChild.Data == "strong" {
			if d := v.FirstChild.FirstChild.Data; !strings.Contains(d, "↓↓↓") ||
				!strings.Contains(d, "点击浏览") {
				body += fmt.Sprintf("\n**%s**  \n", d)
			}
			if t := v.FirstChild.NextSibling; t != nil && t.Type == html.TextNode {
				body += t.Data
			}
		} else {
			ok := true

			for _, a := range v.Parent.Attr {
				if a.Key == "class" {
					switch a.Val {
					// if it is a info for picture, igonre!
					case "sc-bdVaJa iHZvIS":
						ok = false
					// if it is a twitter content, ignore!
					case "twitter-tweet":
						ok = false
					}
				}
			}
			if ok {
				body += v.FirstChild.Data + "  \n"
			}
		}
	}
	rp := strings.NewReplacer("strong", "", "**推荐阅读：**  \n", "")
	body = rp.Replace(body)
	return body, nil
}

func (a *Article) fmtContent(body string) (string, error) {
	var err error
	title := "# " + a.Title + "\n\n"
	lastupdate := shanghai(a.UpdateTime.AsTime()).Format("LastUpdate: [02.01] [1504H]")
	webTitle := fmt.Sprintf(" @ [%s](/list/?v=%[1]s): [%[2]s](http://%[2]s)", a.WebsiteTitle, a.WebsiteDomain)
	u, err := url.QueryUnescape(a.U.String())
	if err != nil {
		u = a.U.String() + "\n\nunescape url error:\n" + err.Error()
	}

	body = title +
		lastupdate +
		webTitle + "\n\n" +
		"---\n" +
		body + "\n\n" +
		"原地址：" + fmt.Sprintf("[%s](%[1]s)", u)
	return body, nil
}
