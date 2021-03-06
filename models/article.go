package models

import (
	"fmt"

	"github.com/BeanCookie/magic-box-api/pkg/util"
	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/gjson"
)

const JUEJIN = "juejin"
const CSDN = "csdn"

type Article struct {
	Model

	ID           string `gorm:"primary_key" json:"id"`
	Title        string `json:"title"`
	CoverImage   string `json:"cover_image"`
	BriefContent string `json:"brief_content"`
	CollectCount int    `json:"collect_count"`
	CommentCount int    `json:"comment_count"`
	ViewCount    int    `json:"view_count"`

	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`

	// Category   Category `json:"category"`
	Platform string `json:"platform"`
}

func MapArticle(items []*Article, fn func(index int, item *Article) *Article) []*Article {
	mappedItems := []*Article{}
	for index, value := range items {
		mappedItems = append(mappedItems, fn(index, value))
	}
	return mappedItems
}

func AddJuejinArticle(data gjson.Result) error {
	log.Info().Msgf("add article %s", data.Get("title").String())
	article := Article{
		ID:           data.Get("article_id").String(),
		Title:        data.Get("title").String(),
		CoverImage:   data.Get("cover_image").String(),
		BriefContent: data.Get("brief_content").String(),
		CollectCount: int(data.Get("collect_count").Int()),
		CommentCount: int(data.Get("comment_count").Int()),
		ViewCount:    int(data.Get("view_count").Int()),
		Platform:     JUEJIN,
	}
	if err := db.Create(&article).Error; err != nil {
		log.Error().Msgf("%v %v", article, data, err)
		return err
	}
	return nil
}

func UpdateJuejinArticle(article Article, data gjson.Result) {
	updateArticle := Article{CollectCount: int(data.Get("collect_count").Int()), CommentCount: int(data.Get("comment_count").Int()), ViewCount: int(data.Get("view_count").Int())}
	db.Model(&article).Select("collect_count", "comment_count", "view_count").Updates(updateArticle)
}

func UpdateCsdnArticle(article Article, data gjson.Result) {
	updateArticle := Article{CollectCount: int(data.Get("favorCount").Int()), CommentCount: int(data.Get("commentCount").Int()), ViewCount: int(data.Get("viewCount").Int())}
	db.Model(&article).Select("collect_count", "comment_count", "view_count").Updates(updateArticle)
}

func AddCsdnArticle(data gjson.Result) error {
	log.Info().Msgf("add article %s", data.Get("articleTitle").String())
	picList := data.Get("picList").Array()
	coverImage := ""
	if len(picList) > 0 {
		coverImage = picList[0].String()
	}
	id := util.ParseCsdnId(data.Get("articleDetailUrl").String())

	article := Article{
		ID:           id,
		Title:        data.Get("articleTitle").String(),
		UserName:     data.Get("userName").String(),
		CollectCount: int(data.Get("favorCount").Int()),
		CommentCount: int(data.Get("commentCount").Int()),
		ViewCount:    int(data.Get("viewCount").Int()),
		CoverImage:   coverImage,
		Platform:     CSDN,
	}
	if err := db.Create(&article).Error; err != nil {
		log.Error().Msgf("%v %v", article, err)
		return err
	}
	return nil
}

func FirstArticleByIdAndPlatform(id string, platform string) (Article, error) {
	var article Article
	err := db.Select("id").Where("id = ? AND platform = ? AND deleted_on = ? ", id, platform, 0).First(&article).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return article, err
	}
	return article, nil
}

func ExistArticleByTitleAndPlatform(title string, platform string) (bool, error) {
	var article Article
	err := db.Select("title").Where("title = ? AND platform = ? AND deleted_on = ? ", title, platform, 0).First(&article).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	if article.ID != "" {
		return true, nil
	}
	return false, nil
}

func GetArticles(page int, size int, maps map[string]interface{}) ([]*Article, error) {
	var articles []*Article
	title, ok := maps["title"]
	delete(maps, "title")
	query := db.Debug().Where(maps)
	if ok && title != "" {
		query = query.Where("title LIKE ?", fmt.Sprintf("%s%v%s", "%", title, "%"))
	}
	err := query.Offset(page).Limit(size).Order("created_on desc").Find(&articles).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return articles, nil
}

func GetArticleTotal(maps map[string]interface{}) (uint, error) {
	var total uint
	title, ok := maps["title"]
	delete(maps, "title")
	query := db.Debug().Where(maps)
	if ok && title != "" {
		query = query.Where("title LIKE ?", fmt.Sprintf("%s%v%s", "%", title, "%"))
	}
	if err := query.Model(&Article{}).Where(maps).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
