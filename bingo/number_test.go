package bingo

import "testing"

func TestNumberString(t *testing.T) {
	for _, test := range numberTests {
		if want, got := test.text, test.n.String(); want != got {
			t.Errorf("Number(%v): wanted %v, got %v", int(test.n), want, got)
		}
	}
}

func TestNumberStringInvalid(t *testing.T) {
	tests := []Number{
		0,
		-1,
		76,
		100,
	}
	for _, n := range tests {
		if want, got := "?", n.String(); want != got {
			t.Errorf("Number(%v): wanted %v, got %v", int(n), want, got)
		}
	}
}

func TestNumberColumn(t *testing.T) {
	for _, test := range numberTests {
		if want, got := test.column, test.n.Column(); want != got {
			t.Errorf("Number(%v) Column: wanted %v, got %v", int(test.n), want, got)
		}
	}
}

func TestNumberValue(t *testing.T) {
	for _, test := range numberTests {
		if want, got := test.value, test.n.Value(); want != got {
			t.Errorf("Number(%v) Value: wanted %v, got %v", int(test.n), want, got)
		}
	}
}

var numberTests = []struct {
	n      Number
	text   string
	column int
	value  int
}{
	{1, "B 1", 0, 1},
	{2, "B 2", 0, 2},
	{3, "B 3", 0, 3},
	{4, "B 4", 0, 4},
	{5, "B 5", 0, 5},
	{6, "B 6", 0, 6},
	{7, "B 7", 0, 7},
	{8, "B 8", 0, 8},
	{9, "B 9", 0, 9},
	{10, "B 10", 0, 10},
	{11, "B 11", 0, 11},
	{12, "B 12", 0, 12},
	{13, "B 13", 0, 13},
	{14, "B 14", 0, 14},
	{15, "B 15", 0, 15},
	{16, "I 16", 1, 16},
	{17, "I 17", 1, 17},
	{18, "I 18", 1, 18},
	{19, "I 19", 1, 19},
	{20, "I 20", 1, 20},
	{21, "I 21", 1, 21},
	{22, "I 22", 1, 22},
	{23, "I 23", 1, 23},
	{24, "I 24", 1, 24},
	{25, "I 25", 1, 25},
	{26, "I 26", 1, 26},
	{27, "I 27", 1, 27},
	{28, "I 28", 1, 28},
	{29, "I 29", 1, 29},
	{30, "I 30", 1, 30},
	{31, "N 31", 2, 31},
	{32, "N 32", 2, 32},
	{33, "N 33", 2, 33},
	{34, "N 34", 2, 34},
	{35, "N 35", 2, 35},
	{36, "N 36", 2, 36},
	{37, "N 37", 2, 37},
	{38, "N 38", 2, 38},
	{39, "N 39", 2, 39},
	{40, "N 40", 2, 40},
	{41, "N 41", 2, 41},
	{42, "N 42", 2, 42},
	{43, "N 43", 2, 43},
	{44, "N 44", 2, 44},
	{45, "N 45", 2, 45},
	{46, "G 46", 3, 46},
	{47, "G 47", 3, 47},
	{48, "G 48", 3, 48},
	{49, "G 49", 3, 49},
	{50, "G 50", 3, 50},
	{51, "G 51", 3, 51},
	{52, "G 52", 3, 52},
	{53, "G 53", 3, 53},
	{54, "G 54", 3, 54},
	{55, "G 55", 3, 55},
	{56, "G 56", 3, 56},
	{57, "G 57", 3, 57},
	{58, "G 58", 3, 58},
	{59, "G 59", 3, 59},
	{60, "G 60", 3, 60},
	{61, "O 61", 4, 61},
	{62, "O 62", 4, 62},
	{63, "O 63", 4, 63},
	{64, "O 64", 4, 64},
	{65, "O 65", 4, 65},
	{66, "O 66", 4, 66},
	{67, "O 67", 4, 67},
	{68, "O 68", 4, 68},
	{69, "O 69", 4, 69},
	{70, "O 70", 4, 70},
	{71, "O 71", 4, 71},
	{72, "O 72", 4, 72},
	{73, "O 73", 4, 73},
	{74, "O 74", 4, 74},
	{75, "O 75", 4, 75},
}
