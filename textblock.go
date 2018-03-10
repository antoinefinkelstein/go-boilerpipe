package boilerpipe

import (
	"bytes"
	"math"
)

type Label int

//go:generate stringer -type=Label
const (
	LabelIndicatesEndOfText Label = iota
	LabelMightBeContent
	LabelVeryLikelyContent
	LabelTitle
	LabelList
	LabelHeading
	LabelHeading1
	LabelHeading2
	LabelHeading3
)

type LabelStack struct {
	labels []Label
}

func NewLabelStack() *LabelStack {
	return &LabelStack{
		labels: make([]Label, 0),
	}
}

func (x *LabelStack) Len() int {
	return len(x.labels)
}

func (x *LabelStack) Pop() (label Label, ok bool) {
	if len(x.labels) == 0 {
		return
	}
	label = x.labels[len(x.labels)-1]
	ok = true
	x.labels = x.labels[:len(x.labels)-1]
	return
}

func (x *LabelStack) PopAll() (labels []Label) {
	if x.Len() == 0 {
		return
	}

	labels = make([]Label, x.Len())
	for i, j := x.Len(), 0; i > 0; i-- {
		labels[j] = x.labels[i-1]
		j++
	}
	x.labels = nil
	x.labels = make([]Label, 0)
	return
}

func (x *LabelStack) Push(labels ...Label) {
	x.labels = append(x.labels, labels...)
}

type TextBlock struct {
	Text string

	OffsetBlocksStart int
	OffsetBlocksEnd   int

	NumWords               int
	NumLinkedWords         int
	NumWordsInWrappedLines int
	NumWrappedLines        int
	TagLevel               int

	IsContent bool

	Labels map[Label]bool
}

var (
	textBlockEmptyStart *TextBlock
	textBlockEmptyEnd   *TextBlock
)

func init() {
	textBlockEmptyStart = NewTextBlock()
	textBlockEmptyStart.OffsetBlocksStart = math.MinInt32
	textBlockEmptyStart.OffsetBlocksEnd = math.MinInt32

	textBlockEmptyEnd = NewTextBlock()
	textBlockEmptyEnd.OffsetBlocksStart = math.MaxInt32
	textBlockEmptyEnd.OffsetBlocksEnd = math.MaxInt32
}

func NewTextBlock() (tb *TextBlock) {
	tb = new(TextBlock)
	tb.Labels = make(map[Label]bool)
	return
}

func (tb *TextBlock) AddLabels(labels ...Label) *TextBlock {
	for _, label := range labels {
		tb.Labels[label] = true
	}
	return tb
}

func (tb *TextBlock) HasLabel(label Label) bool {
	_, hasLabel := tb.Labels[label]
	return hasLabel
}

func (tb *TextBlock) MergeNext(next *TextBlock) {
	buf := bytes.NewBufferString(tb.Text)
	buf.WriteRune('\n')
	buf.WriteString(next.Text)
	tb.Text = buf.String()

	tb.NumWords += next.NumWords
	tb.NumLinkedWords += next.NumLinkedWords

	tb.NumWordsInWrappedLines += next.NumWordsInWrappedLines
	tb.NumWrappedLines += next.NumWrappedLines

	tb.OffsetBlocksStart = int(math.Min(float64(tb.OffsetBlocksStart), float64(next.OffsetBlocksStart)))
	tb.OffsetBlocksEnd = int(math.Min(float64(tb.OffsetBlocksEnd), float64(next.OffsetBlocksEnd)))

	tb.IsContent = tb.IsContent || next.IsContent

	// TODO
	//if (containedTextElements == null) {
	//  containedTextElements = (BitSet) next.containedTextElements.clone();
	//} else {
	//  containedTextElements.or(next.containedTextElements);
	//}

	for k, v := range next.Labels {
		tb.Labels[k] = v
	}

	tb.TagLevel = int(math.Min(float64(tb.TagLevel), float64(next.TagLevel)))
}

func (tb *TextBlock) LinkDensity() float64 {
	if tb.NumWords == 0 {
		return 0.0
	}
	return float64(tb.NumLinkedWords) / float64(tb.NumWords)
}

func (tb *TextBlock) TextDensity() float64 {
	return float64(tb.NumWordsInWrappedLines) / float64(tb.NumWrappedLines)
}
