// internal/app/fetch.go
package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/chromedp/chromedp"
)

// QuestionRecord は CSV 1 行分（q_num,q_num_sub,question,answer,explain）を表す。
// answer は仕様上 true/false/空 の3値なので *bool にしている（nil=空）。
type QuestionRecord struct {
	QNum    int
	QNumSub string
	// question/explain は「全ての改行、空白を空文字に置換」した値
	Question string
	Answer   *bool
	Explain  string
}

// FetchQuestionExplanation は docs/仕様.md の「4. 問題の取得」に従い、
// https://www.musasi.jp/question/explanation/[q_num]#disabledHistoryback にアクセスして
// question/answer/explain を取得する。
func FetchQuestionExplanation(ctx context.Context, qNum int) (*QuestionRecord, error) {
	if qNum <= 0 {
		return nil, fmt.Errorf("qNum must be >= 1: %d", qNum)
	}

	url := fmt.Sprintf(
		"https://www.musasi.jp/question/explanation/%d#disabledHistoryback", qNum)

	var (
		rawQuestion string
		rawExplain  string
		rawAnswer   string
	)

	// 仕様の DOM 参照:
	// - question: document.querySelector('#questions .noBoth').textContent
	// - answer:   document.querySelector('#answer img')?.src
	// - explain:  document.querySelector('#explain .noBoth').textContent
	if err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),

		chromedp.Evaluate(`(() => {
			const el = document.querySelector('#questions .noBoth');
			return el ? (el.textContent ?? '') : '';
		})()`, &rawQuestion),

		chromedp.Evaluate(`(() => {
			const img = document.querySelector('#answer img');
			if (!img) return '';
			// src 属性値を優先（相対パスのこともあり得る）
			const v = img.getAttribute('src');
			return v ? v : (img.src ?? '');
		})()`, &rawAnswer),

		chromedp.Evaluate(`(() => {
			const el = document.querySelector('#explain .noBoth');
			return el ? (el.textContent ?? '') : '';
		})()`, &rawExplain),
	); err != nil {
		return nil, fmt.Errorf("explanation page scrape failed (q=%d): %w", qNum, err)
	}

	question := normalizeNoWhitespace(rawQuestion)
	explain := normalizeNoWhitespace(rawExplain)

	var ans *bool
	switch {
	case strings.Contains(rawAnswer, "true"):
		v := true
		ans = &v
	case strings.Contains(rawAnswer, "false"):
		v := false
		ans = &v
	default:
		// nil のまま（空）
	}

	return &QuestionRecord{
		QNum:     qNum,
		QNumSub:  "", // 仕様では基本空（複小問対応は後段で拡張）
		Question: question,
		Answer:   ans,
		Explain:  explain,
	}, nil
}

// normalizeNoWhitespace は「改行、空白を空文字に置換」に相当する処理。
// strings.Fields であらゆる空白（スペース/改行/タブ等）を落として結合する。
func normalizeNoWhitespace(s string) string {
	if s == "" {
		return ""
	}
	return strings.Join(strings.Fields(s), "")
}

// AnswerString は CSV 書き出し用の文字列化（true/false/空）。
func (r *QuestionRecord) AnswerString() string {
	if r == nil || r.Answer == nil {
		return ""
	}
	if *r.Answer {
		return "true"
	}
	return "false"
}
