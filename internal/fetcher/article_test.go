package fetcher

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/hi20160616/exhtml"
	"github.com/hi20160616/ms-dwnews/configs"
	"github.com/pkg/errors"
)

// pass test
func TestFetchArticle(t *testing.T) {
	tests := []struct {
		url string
		err error
	}{
		{"https://www.dwnews.com/%E4%B8%AD%E5%9B%BD/60245064/%E6%B9%96%E5%8C%97%E9%AB%98%E8%80%83%E7%94%9F%E6%8B%8D%E9%A2%98%E4%B8%8A%E4%BC%A0%E5%90%8D%E5%AD%97%E6%9B%9D%E5%85%89%E6%95%99%E8%82%B2%E9%83%A8%E7%AD%89%E5%A4%9A%E6%96%B9%E7%B4%A7%E6%80%A5%E5%9B%9E%E5%BA%94%E5%9B%BE", ErrTimeOverDays},
		{"https://www.dwnews.com/%E4%B8%AD%E5%9B%BD/60244065/%E4%B8%AD%E5%9B%BD%E6%B0%91%E9%97%B4%E5%90%81%E4%B8%96%E5%8D%AB%E8%B0%83%E6%9F%A5%E7%BE%8E%E7%94%9F%E7%89%A9%E5%AE%9E%E9%AA%8C%E5%AE%A4%E5%8C%97%E4%BA%AC%E4%B8%BE%E4%BE%8B%E7%A7%8D%E7%A7%8D%E7%96%91%E7%82%B9", nil},
	}
	for _, tc := range tests {
		a := NewArticle()
		a, err := a.fetchArticle(tc.url)
		if err != nil {
			if !errors.Is(err, ErrTimeOverDays) {
				t.Error(err)
			} else {
				fmt.Println("ignore old news pass test: ", tc.url)
			}
		} else {
			fmt.Println("pass test: ", a.Content)
		}
	}
}

func TestFetchTitle(t *testing.T) {
	tests := []struct {
		url   string
		title string
	}{
		{
			"https://www.dwnews.com/%E4%B8%AD%E5%9B%BD/60245064/%E6%B9%96%E5%8C%97%E9%AB%98%E8%80%83%E7%94%9F%E6%8B%8D%E9%A2%98%E4%B8%8A%E4%BC%A0%E5%90%8D%E5%AD%97%E6%9B%9D%E5%85%89%E6%95%99%E8%82%B2%E9%83%A8%E7%AD%89%E5%A4%9A%E6%96%B9%E7%B4%A7%E6%80%A5%E5%9B%9E%E5%BA%94%E5%9B%BE",
			"湖北高考生拍题上传名字曝光　教育部等多方紧急回应[图]｜中国",
		},
		{
			"https://www.dwnews.com/%E4%B8%AD%E5%9B%BD/60244065/%E4%B8%AD%E5%9B%BD%E6%B0%91%E9%97%B4%E5%90%81%E4%B8%96%E5%8D%AB%E8%B0%83%E6%9F%A5%E7%BE%8E%E7%94%9F%E7%89%A9%E5%AE%9E%E9%AA%8C%E5%AE%A4%E5%8C%97%E4%BA%AC%E4%B8%BE%E4%BE%8B%E7%A7%8D%E7%A7%8D%E7%96%91%E7%82%B9",
			"中国民间吁世卫调查美生物实验室　北京回应列举众多疑点｜中国",
		},
	}
	for _, tc := range tests {
		a := NewArticle()
		u, err := url.Parse(tc.url)
		if err != nil {
			t.Error(err)
		}
		a.U = u
		// Dail
		a.raw, a.doc, err = exhtml.GetRawAndDoc(a.U, timeout)
		if err != nil {
			t.Error(err)
		}
		got, err := a.fetchTitle()
		if err != nil {
			if !errors.Is(err, ErrTimeOverDays) {
				t.Error(err)
			} else {
				fmt.Println("ignore pass test: ", tc.url)
			}
		} else {
			if tc.title != got {
				t.Errorf("\nwant: %s\n got: %s", tc.title, got)
			}
		}
	}

}

func TestFetchUpdateTime(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{
			"https://www.dwnews.com/%E4%B8%AD%E5%9B%BD/60245064/%E6%B9%96%E5%8C%97%E9%AB%98%E8%80%83%E7%94%9F%E6%8B%8D%E9%A2%98%E4%B8%8A%E4%BC%A0%E5%90%8D%E5%AD%97%E6%9B%9D%E5%85%89%E6%95%99%E8%82%B2%E9%83%A8%E7%AD%89%E5%A4%9A%E6%96%B9%E7%B4%A7%E6%80%A5%E5%9B%9E%E5%BA%94%E5%9B%BE",
			"2021-06-08 13:45:00 +0800 UTC",
		},
		{
			"https://www.dwnews.com/%E4%B8%AD%E5%9B%BD/60244065/%E4%B8%AD%E5%9B%BD%E6%B0%91%E9%97%B4%E5%90%81%E4%B8%96%E5%8D%AB%E8%B0%83%E6%9F%A5%E7%BE%8E%E7%94%9F%E7%89%A9%E5%AE%9E%E9%AA%8C%E5%AE%A4%E5%8C%97%E4%BA%AC%E4%B8%BE%E4%BE%8B%E7%A7%8D%E7%A7%8D%E7%96%91%E7%82%B9",
			"2021-06-02 16:54:00 +0800 UTC",
		},
	}
	var err error
	if err := configs.Reset("../../"); err != nil {
		t.Error(err)
	}

	for _, tc := range tests {
		a := NewArticle()
		a.U, err = url.Parse(tc.url)
		if err != nil {
			t.Error(err)
		}
		// Dail
		a.raw, a.doc, err = exhtml.GetRawAndDoc(a.U, timeout)
		if err != nil {
			t.Error(err)
		}
		tt, err := a.fetchUpdateTime()
		if err != nil {
			t.Error(err)
		} else {
			ttt := tt.AsTime()
			got := shanghai(ttt)
			if got.String() != tc.want {
				t.Errorf("\nwant: %s\n got: %s", tc.want, got.String())
			}
		}
	}
}

func TestFetchContent(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{
			"https://www.dwnews.com/%E4%B8%AD%E5%9B%BD/60245064/%E6%B9%96%E5%8C%97%E9%AB%98%E8%80%83%E7%94%9F%E6%8B%8D%E9%A2%98%E4%B8%8A%E4%BC%A0%E5%90%8D%E5%AD%97%E6%9B%9D%E5%85%89%E6%95%99%E8%82%B2%E9%83%A8%E7%AD%89%E5%A4%9A%E6%96%B9%E7%B4%A7%E6%80%A5%E5%9B%9E%E5%BA%94%E5%9B%BE",
			"2021-06-08 13:45:00 +0800 UTC",
		},
		{
			"https://www.dwnews.com/%E4%B8%AD%E5%9B%BD/60244065/%E4%B8%AD%E5%9B%BD%E6%B0%91%E9%97%B4%E5%90%81%E4%B8%96%E5%8D%AB%E8%B0%83%E6%9F%A5%E7%BE%8E%E7%94%9F%E7%89%A9%E5%AE%9E%E9%AA%8C%E5%AE%A4%E5%8C%97%E4%BA%AC%E4%B8%BE%E4%BE%8B%E7%A7%8D%E7%A7%8D%E7%96%91%E7%82%B9",
			"2021-06-02 16:54:00 +0800 UTC",
		},
	}
	var err error
	if err := configs.Reset("../../"); err != nil {
		t.Error(err)
	}

	for _, tc := range tests {
		a := NewArticle()
		a.U, err = url.Parse(tc.url)
		if err != nil {
			t.Error(err)
		}
		// Dail
		a.raw, a.doc, err = exhtml.GetRawAndDoc(a.U, timeout)
		if err != nil {
			t.Error(err)
		}
		c, err := a.fetchContent()
		if err != nil {
			t.Error(err)
		}
		fmt.Println(c)
	}
}
