package front

import (
	"reflect"
	"testing"
)

func Test_findCurrentIndex(t *testing.T) {
	type args struct {
		selectedFile string
		files        []string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"test normal", args{"1.png", []string{"1.png", "2.png"}}, 0},
		{"test empty files", args{"1.png", []string{}}, -1},
		{"test empty files", args{"3.png", []string{"1.png", "2.png"}}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findCurrentIndex(tt.args.selectedFile, tt.args.files); got != tt.want {
				t.Errorf("findCurrentIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_takePreviews(t *testing.T) {
	type args struct {
		previewImagesLimit int
		files              []string
		currentIndex       int
	}

	manyFiles := []string {
		"1.png",
		"2.png",
		"3.png",
		"4.png",
		"5.png",
		"6.png",
		"7.png",
		"8.png",
		"9.png",
		"10.png",
	}

	tests := []struct {
		name         string
		args         args
		wantPreviews []string
		wantL        int
		wantR        int
	}{
		{"from start", args{5, manyFiles, 1 }, manyFiles[:5], 1, 5},
		{"middle", args{5, manyFiles, 4 }, manyFiles[2:7], 3, 7},
		{"from end", args{5, manyFiles, 9 }, manyFiles[5:10], 6, 10},
		{"empty files", args{5, []string{}, 0 }, nil, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPreviews, gotL, gotR := takePreviews(tt.args.previewImagesLimit, tt.args.files, tt.args.currentIndex)
			if !reflect.DeepEqual(gotPreviews, tt.wantPreviews) {
				t.Errorf("takePreviews() gotPreviews = %v, want %v", gotPreviews, tt.wantPreviews)
			}
			if gotL != tt.wantL {
				t.Errorf("takePreviews() gotL = %v, want %v", gotL, tt.wantL)
			}
			if gotR != tt.wantR {
				t.Errorf("takePreviews() gotR = %v, want %v", gotR, tt.wantR)
			}
		})
	}
}