package zhenai

import (
	"../../engine"
	"../../model/zhenai"
	"fmt"
	"regexp"
)

/*
	<div class="m-content-box m-des" data-v-8b1eac0c>

<span data-v-8b1eac0c>
河北人，北京户口，硕士毕业，国企人事工作。有贷款购房，没有车，薪酬税前二十。没有不良嗜好，酒烟很少沾，平时爬爬山，打打游戏，玩玩狼人杀。
</span>
</div>
*/
var osRe = regexp.MustCompile(`<div class="m-content-box m-des" data-v-8b1eac0c><span data-v-8b1eac0c>([^<]+)</span></div>`)

/*
*
<div class="purple-btns" data-v-8b1eac0c>

	     <div class="m-btn purple" data-v-8b1eac0c>
	     未婚</div>
	     <div class="m-btn purple" data-v-8b1eac0c>
	     28岁</div>
	     <div class="m-btn purple" data-v-8b1eac0c>
	     射手座(11.22-12.21)</div>
	     <div class="m-btn purple" data-v-8b1eac0c>
	     177cm</div>
	     <div class="m-btn purple" data-v-8b1eac0c>
	     85kg</div>
	     <div class="m-btn purple" data-v-8b1eac0c>
	     工作地:北京丰台区</div>
	     <div class="m-btn purple" data-v-8b1eac0c>
	     月收入:2-5万</div>
	     <div class="m-btn purple" data-v-8b1eac0c>
	     人事主管</div>
	     <div class="m-btn purple" data-v-8b1eac0c>
	     硕士</div>
	</div>
*/
var purpleRe = regexp.MustCompile(`<div class="m-btn purple" data-v-8b1eac0c>([^<]+)</div>`)

var pinkRe = regexp.MustCompile(`<div class="m-btn pink" data-v-8b1eac0c>([^<]+)</div>`)

var conditionRe = regexp.MustCompile(` <div class="m-btn" data-v-8b1eac0c>([^<]+)</div>`)

var favoriteFoodRe = regexp.MustCompile(`<div class="question f-fl" data-v-8b1eac0c>喜欢的一道菜：</div><div class="answer f-fl" data-v-8b1eac0c>([^<]+)</div>`)

var favoritePersonRe = regexp.MustCompile(`<div class="question f-fl" data-v-8b1eac0c>欣赏的一个名人：</div><div class="answer f-fl" data-v-8b1eac0c>([^<]+)</div>`)

var favoriteSongRe = regexp.MustCompile(`<div class="question f-fl" data-v-8b1eac0c>喜欢的一首歌：</div><div class="answer f-fl" data-v-8b1eac0c>([^<]+)</div>`)

var favoriteBookRe = regexp.MustCompile(`<div class="question f-fl" data-v-8b1eac0c>喜欢的一本书：</div><div class="answer f-fl" data-v-8b1eac0c>([^<]+)</div>`)

var favoriteThingRe = regexp.MustCompile(`<div class="question f-fl" data-v-8b1eac0c>喜欢做的事：</div><div class="answer f-fl" data-v-8b1eac0c>([^<]+)</div>`)

func ParseProfile(contents []byte, name string) engine.ParseResult {
	match := osRe.FindSubmatch(contents)
	profile := zhenai.Profile{}

	profile.Name = name
	fmt.Printf("get person profile 1: %v", profile)
	if match != nil {
		profile.Os = string(match[1])
	} else {

	}
	purpleResults := purpleRe.FindAllSubmatch(contents, -1)
	if purpleResults != nil {
		for i := 0; i < len(purpleResults); i++ {
			profile.PurpleList = append(profile.PurpleList, string(purpleResults[i][1]))
		}
	} else {

	}
	pinkResults := pinkRe.FindAllSubmatch(contents, -1)
	if pinkResults != nil {
		for i := 0; i < len(pinkResults); i++ {
			profile.PinkList = append(profile.PinkList, string(pinkResults[i][1]))
		}
	} else {

	}
	conditionResults := conditionRe.FindAllSubmatch(contents, -1)
	if conditionResults != nil {
		for i := 0; i < len(conditionResults); i++ {
			profile.ConditionList = append(profile.ConditionList, string(conditionResults[i][1]))
		}
	} else {

	}
	favoriteBook := favoriteBookRe.FindSubmatch(contents)
	if favoriteBook != nil {
		profile.FavoriteBook = string(favoriteBook[1])
	} else {

	}
	favoriteSong := favoriteSongRe.FindSubmatch(contents)
	if favoriteSong != nil {
		profile.FavoriteSong = string(favoriteSong[1])
	} else {

	}
	favoriteFood := favoriteFoodRe.FindSubmatch(contents)
	if favoriteFood != nil {
		profile.FavoriteFood = string(favoriteFood[1])
	} else {

	}
	favoriteThing := favoriteThingRe.FindSubmatch(contents)
	if favoriteThing != nil {
		profile.FavoriteThing = string(favoriteThing[1])
	} else {

	}
	favoritePerson := favoritePersonRe.FindSubmatch(contents)
	if favoritePerson != nil {
		profile.FavoritePerson = string(favoritePerson[1])
	} else {

	}
	fmt.Printf("get person profile : %v", profile)
	return engine.ParseResult{
		Items: []interface{}{profile},
	}
}
