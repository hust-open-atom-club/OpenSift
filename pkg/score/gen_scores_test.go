package score

import (
	"math"
	"testing"
)

func TestCalculateDistScore(t *testing.T) {
	distScore := &DistScore{
		DistImpact:   0.5,
		DistPageRank: 0.5,
	}

	expectedScore := (weights["distScore"]["dist_impact"] * distScore.DistImpact) + (weights["distScore"]["dist_pagerank"] * distScore.DistPageRank)
	distScore.CalculateDistScore("log")

	if distScore.DistScore != expectedScore {
		t.Errorf("Expected score %v, but got %v", expectedScore, distScore.DistScore)
	}
}

func TestLogNormalize(t *testing.T) {
	value := 10.0
	threshold := 100.0

	expected := math.Log(value+1) / math.Log(math.Max(value, threshold)+1)
	actual := LogNormalize(value, threshold)

	if actual != expected {
		t.Errorf("Expected %v, but got %v", expected, actual)
	}
}
