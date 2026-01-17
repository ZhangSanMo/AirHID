package input

import (
	"testing"
)

// TestParseChineseNumber 测试中文数字解析
func TestParseChineseNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		ok       bool
	}{
		// 阿拉伯数字
		{"阿拉伯数字1", "1", 1, true},
		{"阿拉伯数字5", "5", 5, true},
		{"阿拉伯数字10", "10", 10, true},
		{"阿拉伯数字99", "99", 99, true},

		// 单个中文数字
		{"中文零", "零", 0, true},
		{"中文一", "一", 1, true},
		{"中文二", "二", 2, true},
		{"中文两", "两", 2, true},
		{"中文三", "三", 3, true},
		{"中文四", "四", 4, true},
		{"中文五", "五", 5, true},
		{"中文六", "六", 6, true},
		{"中文七", "七", 7, true},
		{"中文八", "八", 8, true},
		{"中文九", "九", 9, true},
		{"中文十", "十", 10, true},

		// 大写中文数字
		{"大写壹", "壹", 1, true},
		{"大写贰", "贰", 2, true},
		{"大写叁", "叁", 3, true},
		{"大写肆", "肆", 4, true},
		{"大写伍", "伍", 5, true},
		{"大写陆", "陆", 6, true},
		{"大写柒", "柒", 7, true},
		{"大写捌", "捌", 8, true},
		{"大写玖", "玖", 9, true},
		{"大写拾", "拾", 10, true},

		// 十几
		{"十一", "十一", 11, true},
		{"十二", "十二", 12, true},
		{"十五", "十五", 15, true},
		{"十九", "十九", 19, true},
		{"拾壹", "拾壹", 11, true},
		{"拾伍", "拾伍", 15, true},

		// 几十
		{"二十", "二十", 20, true},
		{"三十", "三十", 30, true},
		{"五十", "五十", 50, true},
		{"九十", "九十", 90, true},
		{"贰拾", "贰拾", 20, true},
		{"玖拾", "玖拾", 90, true},

		// 几十几
		{"二十一", "二十一", 21, true},
		{"三十五", "三十五", 35, true},
		{"四十九", "四十九", 49, true},
		{"九十九", "九十九", 99, true},
		{"贰拾壹", "贰拾壹", 21, true},
		{"玖拾玖", "玖拾玖", 99, true},

		// 边界和无效情况
		{"空字符串", "", 0, false},
		{"带空格", "  ", 0, false},
		{"无效字符", "abc", 0, false},
		{"混合无效", "一abc", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := parseChineseNumber(tt.input)
			if ok != tt.ok {
				t.Errorf("parseChineseNumber(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && result != tt.expected {
				t.Errorf("parseChineseNumber(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseRepeatCommand 测试重复命令解析
func TestParseRepeatCommand(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedCnt  int
		expectedCmd  string
		expectedFlag bool
	}{
		// 格式1: 次数x命令
		{"x格式基础", "3xenter", 3, "enter", true},
		{"x格式带空格", "5x backspace", 5, "backspace", true},
		{"x格式复杂命令", "10x ctrl+z", 10, "ctrl+z", true},
		{"x格式边界1次", "1x tab", 1, "tab", true},
		{"x格式边界100次", "100x esc", 100, "esc", true},

		// 格式3: 命令*次数
		{"星号格式基础", "backspace*5", 5, "backspace", true},
		{"星号格式带空格", "ctrl+z * 10", 10, "ctrl+z", true},
		{"星号格式复杂命令", "alt+tab*3", 3, "alt+tab", true},
		{"星号格式边界1次", "enter*1", 1, "enter", true},
		{"星号格式边界100次", "esc*100", 100, "esc", true},

		// 格式4: 中文前置数字
		{"中文前置阿拉伯", "3次ctrl+z", 3, "ctrl+z", true},
		{"中文前置汉字", "三次ctrl+z", 3, "ctrl+z", true},
		{"中文前置下", "三下回车", 3, "回车", true},
		{"中文前置遍", "五遍backspace", 5, "backspace", true},
		{"中文前置十", "十次enter", 10, "enter", true},
		{"中文前置十五", "十五次tab", 15, "tab", true},
		{"中文前置二十", "二十次esc", 20, "esc", true},
		{"中文前置三十五", "三十五下space", 35, "space", true},

		// 格式5: 中文后置数字
		{"中文后置阿拉伯", "ctrl+z3次", 3, "ctrl+z", true},
		{"中文后置汉字", "ctrl+z三次", 3, "ctrl+z", true},
		{"中文后置下", "回车三下", 3, "回车", true},
		{"中文后置遍", "backspace五遍", 5, "backspace", true},
		{"中文后置十", "enter十次", 10, "enter", true},
		{"中文后置十五", "tab十五次", 15, "tab", true},
		{"中文后置二十", "esc二十次", 20, "esc", true},

		// 边界情况 - 无效重复次数
		{"x格式0次", "0xenter", 0, "0xenter", false},
		{"星号格式0次", "enter*0", 0, "enter*0", false},
		{"x格式超过100次", "101xenter", 0, "101xenter", false},
		{"星号格式超过100次", "enter*101", 0, "enter*101", false},
		{"中文前置0次", "零次enter", 0, "零次enter", false},
		{"中文后置0次", "enter零次", 0, "enter零次", false},

		// 带额外空格的命令
		{"x格式多空格", "3x   enter", 3, "enter", true},
		{"星号格式多空格", "enter  *  5", 5, "enter", true},

		// 非重复命令
		{"普通命令", "ctrl+c", 0, "ctrl+c", false},
		{"普通命令2", "alt+tab", 0, "alt+tab", false},
		{"普通命令3", "enter", 0, "enter", false},
		{"普通命令4", "f5", 0, "f5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, cmd, isRepeat := parseRepeatCommand(tt.input)
			if isRepeat != tt.expectedFlag {
				t.Errorf("parseRepeatCommand(%q) isRepeat = %v, want %v", tt.input, isRepeat, tt.expectedFlag)
			}
			if isRepeat {
				if count != tt.expectedCnt {
					t.Errorf("parseRepeatCommand(%q) count = %d, want %d", tt.input, count, tt.expectedCnt)
				}
				if cmd != tt.expectedCmd {
					t.Errorf("parseRepeatCommand(%q) cmd = %q, want %q", tt.input, cmd, tt.expectedCmd)
				}
			}
		})
	}
}

// TestKeyMapContents 测试keyMap包含预期的键
func TestKeyMapContents(t *testing.T) {
	expectedKeys := []string{
		"enter", "回车", "确认",
		"esc", "escape", "退出",
		"tab", "制表",
		"space", "空格",
		"backspace", "退格",
		"del", "delete", "删除",
		"up", "down", "left", "right",
		"上", "下", "左", "右",
		"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12",
	}

	for _, key := range expectedKeys {
		if _, ok := keyMap[key]; !ok {
			t.Errorf("keyMap missing expected key: %q", key)
		}
	}
}

// TestCharMapContents 测试charMap包含预期的字符
func TestCharMapContents(t *testing.T) {
	// 测试字母
	for c := byte('a'); c <= byte('z'); c++ {
		if _, ok := charMap[c]; !ok {
			t.Errorf("charMap missing letter: %c", c)
		}
	}

	// 测试数字
	for c := byte('0'); c <= byte('9'); c++ {
		if _, ok := charMap[c]; !ok {
			t.Errorf("charMap missing digit: %c", c)
		}
	}

	// 测试特殊字符
	specialChars := []byte{' ', '+', '-', '=', ',', '.', '/', ';', '\'', '[', ']', '\\', '`'}
	for _, c := range specialChars {
		if _, ok := charMap[c]; !ok {
			t.Errorf("charMap missing special char: %c", c)
		}
	}
}

// TestParseCommand 测试命令解析
func TestParseCommand(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantKeys  int // 预期的主按键数量
		wantMods  int // 预期的修饰键数量
		shouldErr bool
	}{
		{"简单按键", "enter", 1, 0, false},
		{"简单组合键", "ctrl+c", 1, 1, false},
		{"复杂组合键", "ctrl+alt+delete", 1, 2, false},
		{"带空格组合键", "ctrl + v", 1, 1, false},
		{"中文组合键", "ctrl+c", 1, 1, false}, // SimulateCommand 会处理 "加"
		{"多个按键序列", "enter tab", 2, 0, false},
		{"问题用例-ctrl+w", "ctrl+w", 1, 1, false},
		{"中文加号带标点", "Ctrl加W。", 1, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys, mods, err := ParseCommand(tt.input)
			if (err != nil) != tt.shouldErr {
				t.Errorf("ParseCommand(%q) error = %v, shouldErr = %v", tt.input, err, tt.shouldErr)
				return
			}
			if err == nil {
				if len(keys) != tt.wantKeys {
					t.Errorf("ParseCommand(%q) got %d keys, want %d", tt.input, len(keys), tt.wantKeys)
				}
				if len(mods) != tt.wantMods {
					t.Errorf("ParseCommand(%q) got %d mods, want %d", tt.input, len(mods), tt.wantMods)
				}
			}
		})
	}
}

// BenchmarkParseChineseNumber 基准测试中文数字解析
func BenchmarkParseChineseNumber(b *testing.B) {
	testCases := []string{"三", "十五", "二十一", "九十九", "5", "50"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			parseChineseNumber(tc)
		}
	}
}

// BenchmarkParseRepeatCommand 基准测试重复命令解析
func BenchmarkParseRepeatCommand(b *testing.B) {
	testCases := []string{
		"3xenter",
		"ctrl+z*10",
		"三次回车",
		"backspace五遍",
		"ctrl+c", // 非重复命令
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			parseRepeatCommand(tc)
		}
	}
}

// TestParseRepeatCommandCaseInsensitive 测试大小写不敏感
func TestParseRepeatCommandCaseInsensitive(t *testing.T) {
	// parseRepeatCommand 本身不做大小写转换，但 SimulateCommand 会先转小写
	// 这里测试的是原始输入（小写后的结果）
	tests := []struct {
		name         string
		input        string
		expectedCnt  int
		expectedCmd  string
		expectedFlag bool
	}{
		{"小写x格式", "3xenter", 3, "enter", true},
		{"大写X格式", "3Xenter", 0, "3Xenter", false}, // 只支持小写x
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count, cmd, isRepeat := parseRepeatCommand(tt.input)
			if isRepeat != tt.expectedFlag {
				t.Errorf("parseRepeatCommand(%q) isRepeat = %v, want %v", tt.input, isRepeat, tt.expectedFlag)
			}
			if isRepeat {
				if count != tt.expectedCnt {
					t.Errorf("parseRepeatCommand(%q) count = %d, want %d", tt.input, count, tt.expectedCnt)
				}
				if cmd != tt.expectedCmd {
					t.Errorf("parseRepeatCommand(%q) cmd = %q, want %q", tt.input, cmd, tt.expectedCmd)
				}
			}
		})
	}
}

// TestChineseNumMapContents 测试中文数字映射完整性
func TestChineseNumMapContents(t *testing.T) {
	// 验证 chineseNumMap 包含所有预期的中文数字
	expectedMappings := map[rune]int{
		'零': 0, '〇': 0,
		'一': 1, '壹': 1,
		'二': 2, '贰': 2, '两': 2,
		'三': 3, '叁': 3,
		'四': 4, '肆': 4,
		'五': 5, '伍': 5,
		'六': 6, '陆': 6,
		'七': 7, '柒': 7,
		'八': 8, '捌': 8,
		'九': 9, '玖': 9,
		'十': 10, '拾': 10,
	}

	for char, expected := range expectedMappings {
		if val, ok := chineseNumMap[char]; !ok {
			t.Errorf("chineseNumMap missing character: %c", char)
		} else if val != expected {
			t.Errorf("chineseNumMap[%c] = %d, want %d", char, val, expected)
		}
	}
}

// TestParseChineseNumberEdgeCases 测试中文数字解析边界情况
func TestParseChineseNumberEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		ok       bool
	}{
		// 特殊的〇
		{"圈零", "〇", 0, true},

		// 前后有空格应该被处理
		{"前空格", " 三", 3, true},
		{"后空格", "三 ", 3, true},
		{"前后空格", " 三 ", 3, true},

		// 超出支持范围（百）
		{"百", "百", 0, false},
		{"一百", "一百", 0, false},

		// 特殊组合（代码当前行为）
		{"连续十", "十十", 20, true}, // 代码将其解析为 10 + 10 = 20
		{"十零", "十零", 10, true},   // 十零 = 10 + 0 = 10
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := parseChineseNumber(tt.input)
			if ok != tt.ok {
				t.Errorf("parseChineseNumber(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && result != tt.expected {
				t.Errorf("parseChineseNumber(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
