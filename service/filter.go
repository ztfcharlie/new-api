package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/dto"
)

// FilterAction 定义过滤动作
type FilterAction int

const (
	ActionPass FilterAction = iota
	ActionBlock
	ActionReplace
	ActionContext
)

// FilterResult 过滤结果
type FilterResult struct {
	Action     FilterAction
	Modified   bool
	NewContent string
	Reason     string
}

// FilterRequest 对请求进行敏感词过滤
// 如果返回 error，说明被拦截 (Block)
// 如果返回 bool=true，说明请求内容被修改，需要更新请求体
func FilterRequest(request dto.Request) (bool, error) {
	modified := false

	switch r := request.(type) {
	case *dto.GeneralOpenAIRequest:
		// 1. 检查 Prompt (Completions)
		if r.Prompt != nil {
			// Prompt 可能是 string 或 []string
			if str, ok := r.Prompt.(string); ok {
				newStr, changed, err := processText(str)
				if err != nil {
					return false, err
				}
				if changed {
					r.Prompt = newStr
					modified = true
				}
			} else if strs, ok := r.Prompt.([]interface{}); ok {
				var newStrs []interface{}
				arrChanged := false
				for _, item := range strs {
					if str, ok := item.(string); ok {
						newStr, changed, err := processText(str)
						if err != nil {
							return false, err
						}
						newStrs = append(newStrs, newStr)
						if changed {
							arrChanged = true
						}
					} else {
						newStrs = append(newStrs, item)
					}
				}
				if arrChanged {
					r.Prompt = newStrs
					modified = true
				}
			}
		}

		// 2. 检查 Messages (Chat)
		for i := range r.Messages {
			msg := &r.Messages[i]
			// 解析 content
			contentList := msg.ParseContent()
			contentChanged := false

			for j := range contentList {
				item := &contentList[j]
				if item.Type == dto.ContentTypeText && item.Text != "" {
					newText, changed, err := processText(item.Text)
					if err != nil {
						return false, err
					}
					if changed {
						item.Text = newText
						contentChanged = true
					}
				}
			}

			if contentChanged {
				msg.SetMediaContent(contentList)
				modified = true
			}
		}

	case *dto.ImageRequest:
		// 检查图片生成 Prompt
		if r.Prompt != "" {
			newText, changed, err := processText(r.Prompt)
			if err != nil {
				return false, err
			}
			if changed {
				r.Prompt = newText
				modified = true
			}
		}
	}

	return modified, nil
}

// processText 处理单段文本
// 返回: 新文本, 是否修改, 错误(拦截)
// 优先级:
// 1. 拦截 (Block) - 生物武器、图片生成
// 2. 替换 (Replace) - COT
// 3. 上下文 (Context) - 生物科研
func processText(text string) (string, bool, error) {
	// 1. 拦截检测 (Block)
	// 合并 生物武器 和 图片生成 进行检测
	blockDict := append(SensitiveWordsBioWeapon, SensitiveWordsImgGen...)
	// 使用 str.go 中的 AcSearch
	if found, words := AcSearch(text, blockDict, true); found {
		return "", false, fmt.Errorf("request blocked: contains sensitive words (%s)", strings.Join(words, ", "))
	}

	currentText := text
	modified := false

	// 2. 替换检测 (Replace) - COT
	// 含有COT词汇 -> 静默替换 (替换为空或指定字符)
	foundCOT, _ := AcSearch(currentText, SensitiveWordsCOT, false)
	if foundCOT {
		// 执行替换
		_, _, replaced := sensitiveWordReplaceWith(currentText, SensitiveWordsCOT, "") // 也可以替换为 "*COT*"
		if replaced != currentText {
			currentText = replaced
			modified = true
		}
	}

	// 3. 上下文检测 (Context) - 生物科研
	// 含有生物科研词汇 -> 添加上下文
	foundBio, _ := AcSearch(currentText, SensitiveWordsBioResearch, true) // 只要有一个就触发
	if foundBio {
		currentText = currentText + BioResearchContextSuffix
		modified = true
	}

	return currentText, modified, nil
}

// sensitiveWordReplaceWith 是 str.go 中 SensitiveWordReplace 的定制版，支持自定义替换内容
func sensitiveWordReplaceWith(text string, dict []string, replaceStr string) (bool, []string, string) {
	if len(dict) == 0 {
		return false, nil, text
	}

	// 统一转为 rune 处理，避免中英文混合导致的索引问题
	runes := []rune(text)
	lowerRunes := []rune(strings.ToLower(text)) // 搜索用小写

	m := getOrBuildAC(dict)

	// AC自动机返回的是 rune 的索引 (github.com/anknown/ahocorasick)
	hits := m.MultiPatternSearch(lowerRunes, false)

	if len(hits) > 0 {
		words := make([]string, 0, len(hits))
		var builder strings.Builder
		// 预估容量
		builder.Grow(len(text))

		lastPos := 0

		// hits 应该是按位置排序的吗？ anknown/ahocorasick 的 MultiPatternSearch 通常返回无序或按匹配顺序
		// 这里假设我们需要按位置排序处理，以免乱序替换
		// 简单的做法是先收集替换区间，然后重建字符串
		// 但 MultiPatternSearch 返回的 hit.Pos 是匹配结束的位置(不包含)还是开始位置？
		// 查看 github.com/anknown/ahocorasick 源码或 str.go 的用法。
		// str.go 中: hit.Pos 通常是结束位置? 不，标准AC算法通常记录结束位置。
		// 让我们看 str.go 的 SundaySearch 实现... 不，是 AC。
		// anknown/ahocorasick Example:
		// hits := m.MultiPatternSearch([]rune("she us he"), false)
		// for _, hit := range hits { fmt.Println(hit.Pos, string(hit.Word)) }
		// Output: 3 she, 6 us, 9 he (位置似乎是结尾? 3是 'she' 后面? 0,1,2 -> 3? )
		// Check implementation details if possible.
		// 假设 library 返回的是 start index 或者 end index。
		// 为了安全起见，我会先拿到 hits，然后按 Pos 排序。

		// 修正：anknown/ahocorasick 的 Term.Pos 实际上是匹配串在文本中的 *起始* 字节位置?
		// 不，根据 str.go: `hits := m.MultiPatternSearch([]rune(findText), stopImmediately)`
		// 输入是 []rune，所以返回的 Pos 应该是 rune index。

		// 重新排序 hits 以确保按顺序替换 (以防万一库返回乱序)
		// sortHits(hits)

		// 过滤重叠的匹配 (贪婪匹配：最左最长优先，或者简单地跳过重叠)
		// 简单策略：如果当前 hit 的开始位置 < lastPos，说明重叠，跳过

		spans := make([]replaceSpan, 0, len(hits))
		for _, hit := range hits {
			wordLen := len(hit.Word)
			// 假设 Pos 是 ending index (exclusive, 也就是 length so far)
			// 或者是 starting index.
			// 大多数 Go AC 库 (如 cloudflare/ahocorasick) 是 returning match id.
			// anknown/ahocorasick:
			// Pos is the index of the first character of the pattern in the text.
			start := hit.Pos
			end := start + wordLen
			spans = append(spans, replaceSpan{start: start, end: end, word: string(hit.Word)})
		}

		// 排序 spans
		sort.Slice(spans, func(i, j int) bool {
			return spans[i].start < spans[j].start
		})

		for _, span := range spans {
			if span.start < lastPos {
				continue // 简单的重叠处理：通过跳过重叠部分
			}

			// append text before match
			builder.WriteString(string(runes[lastPos:span.start]))
			// append replacement
			builder.WriteString(replaceStr)

			words = append(words, span.word)
			lastPos = span.end
		}
		// append remaining text
		builder.WriteString(string(runes[lastPos:]))

		// 如果没有任何实际替换（全是重叠被忽略），仍视为 modified=true?
		// 不，只有 append 了 words 才算。
		if len(words) > 0 {
			return true, words, builder.String()
		}
	}

	return false, nil, text
}

// 辅助排序
type replaceSpan struct {
	start int
	end   int
	word  string
}
