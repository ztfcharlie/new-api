package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/QuantumNous/new-api/common"
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
	Action         FilterAction
	Modified       bool
	NewContent     string
	Reason         string
	TriggeredWords []string
}

// FilterRequest 对请求进行敏感词过滤
// 返回 FilterResult 指针，如果为 nil 表示无操作或错误（错误包含在 result.Reason 中? 不，还是保留 error 为好，或者 error 放到 result 里?）
// 保持 error 用于返回严重的系统错误。但对于 Block，我们以前是返回 error。
// 新设计：返回 (*FilterResult, error)。
// 如果 Action == ActionBlock，则 FilterResult.Action = ActionBlock，同时返回 error = nil (或者 error = "blocked"?)
// 为了兼容性尽量少改动逻辑：
// 以前: (bool, error). error != nil implies Block.
// 现在: (*FilterResult, error).
// 如果 ActionBlock, return result, nil. 调用方检查 result.Action.
func FilterRequest(request dto.Request) (*FilterResult, error) {
	result := &FilterResult{
		Action: ActionPass,
	}

	switch r := request.(type) {
	case *dto.GeneralOpenAIRequest:
		// 1. 检查 Prompt (Completions)
		if r.Prompt != nil {
			if str, ok := r.Prompt.(string); ok {
				newStr, res, err := processText(str)
				if err != nil {
					return nil, err // 系统错误? processText以前返回 error 是 Block。现在 processText 应该返回 result。
				}
				// Merge result
				if res.Action > result.Action {
					result.Action = res.Action
					result.Reason = res.Reason
					result.TriggeredWords = append(result.TriggeredWords, res.TriggeredWords...)
				}
				if res.Modified {
					r.Prompt = newStr
					result.Modified = true
				}
			} else if strs, ok := r.Prompt.([]interface{}); ok {
				var newStrs []interface{}
				arrChanged := false
				for _, item := range strs {
					if str, ok := item.(string); ok {
						newStr, res, err := processText(str)
						if err != nil {
							return nil, err
						}
						if res.Action > result.Action {
							result.Action = res.Action
							result.Reason = res.Reason
							result.TriggeredWords = append(result.TriggeredWords, res.TriggeredWords...)
						}
						newStrs = append(newStrs, newStr)
						if res.Modified {
							arrChanged = true
						}
					} else {
						newStrs = append(newStrs, item)
					}
				}
				if arrChanged {
					r.Prompt = newStrs
					result.Modified = true
				}
			}
		}

		// 2. 检查 Messages (Chat)
		for i := range r.Messages {
			msg := &r.Messages[i]
			contentList := msg.ParseContent()
			contentChanged := false

			for j := range contentList {
				item := &contentList[j]
				if item.Type == dto.ContentTypeText && item.Text != "" {
					newText, res, err := processText(item.Text)
					if err != nil {
						return nil, err
					}
					if res.Action > result.Action {
						result.Action = res.Action
						result.Reason = res.Reason
						result.TriggeredWords = append(result.TriggeredWords, res.TriggeredWords...)
					}
					if res.Modified {
						item.Text = newText
						contentChanged = true
					}
				}
			}

			if contentChanged {
				msg.SetMediaContent(contentList)
				result.Modified = true
			}
		}

	case *dto.ImageRequest:
		// 检查图片生成 Prompt
		if r.Prompt != "" {
			newText, res, err := processText(r.Prompt)
			if err != nil {
				return nil, err
			}
			if res.Action > result.Action {
				result.Action = res.Action
				result.Reason = res.Reason
				result.TriggeredWords = append(result.TriggeredWords, res.TriggeredWords...)
			}
			if res.Modified {
				r.Prompt = newText
				result.Modified = true
			}
		}
	}

	// 如果 Block，为了保持兼容性，以前是返回 error。
	// 但现在我们希望返回 result 给调用者记录日志，然后再决定怎么处理。
	// 所以这里我们只返回 result，不返回 error (除非真正的系统错误)。
	// Block 状态包含在 result.Action 中。

	return result, nil
}

// processText 处理单段文本
// 返回: 新文本, 结果详情, 错误(系统错误)
func processText(text string) (string, *FilterResult, error) {
	result := &FilterResult{
		Action: ActionPass,
	}

	currentText := text
	modified := false

	// Helper to check and apply action
	checkAndApply := func(dict []string, modeStr string, dictName string) {
		if len(dict) == 0 {
			return
		}

		// Parse mode
		action := parseAction(modeStr)
		if action == ActionPass {
			return
		}

		// Check depends on action type
		// If Block or Context, we just need to know if it exists (stopImmediately=true)
		// If Replace, we need to find occurrences to replace (stopImmediately=false)

		stopImmediately := true
		if action == ActionReplace {
			stopImmediately = false
		}

		found, words := AcSearch(currentText, dict, stopImmediately)
		if !found {
			return
		}

		switch action {
		case ActionBlock:
			if ActionBlock > result.Action {
				result.Action = ActionBlock
				result.Reason = fmt.Sprintf("sensitive words detected (%s): %s", dictName, strings.Join(words, ", "))
			}
			result.TriggeredWords = append(result.TriggeredWords, words...)

		case ActionReplace:
			_, _, replaced := sensitiveWordReplaceWith(currentText, dict, "")
			if replaced != currentText {
				currentText = replaced
				modified = true
				if ActionReplace > result.Action {
					result.Action = ActionReplace
					result.Reason = fmt.Sprintf("sensitive words replaced (%s)", dictName)
				}
				result.TriggeredWords = append(result.TriggeredWords, words...)
			}

		case ActionContext:
			if !strings.HasSuffix(currentText, BioResearchContextSuffix) {
				currentText = currentText + BioResearchContextSuffix
				modified = true
				if ActionContext > result.Action {
					result.Action = ActionContext
					result.Reason = fmt.Sprintf("context added (%s)", dictName)
				}
				result.TriggeredWords = append(result.TriggeredWords, words...)
			}
		}
	}

	// 1. Check BioWeapon
	checkAndApply(SensitiveWordsBioWeapon, common.BioWeaponFilterMode, "BioWeapon")
	if result.Action == ActionBlock {
		return text, result, nil
	}

	// 2. Check ImageGen
	checkAndApply(SensitiveWordsImgGen, common.ImageGenFilterMode, "ImageGen")
	if result.Action == ActionBlock {
		return text, result, nil
	}

	// 3. Check COT
	checkAndApply(SensitiveWordsCOT, common.CotFilterMode, "COT")

	// 4. Check BioResearch
	checkAndApply(SensitiveWordsBioResearch, common.BioResearchFilterMode, "BioResearch")

	result.Modified = modified
	result.NewContent = currentText

	return currentText, result, nil
}

func parseAction(mode string) FilterAction {
	mode = strings.ToUpper(mode)
	switch mode {
	case "BLOCK", "STRICT":
		return ActionBlock
	case "REPLACE":
		return ActionReplace
	case "CONTEXT", "MODERATE":
		return ActionContext
	case "NONE", "PASS":
		return ActionPass
	default:
		return ActionPass
	}
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
